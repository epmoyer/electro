# Standard library
from pathlib import Path
import json
import shutil
import shutil
import re
from datetime import datetime, date

# Library
from prettyprinter import pformat
import markdown
from bs4 import BeautifulSoup
from result import Ok, Err, Result
from loguru import logger
from pytest import mark

# Local
from electro.app_config import CONFIG, OUTPUT_FORMATS
from electro.console import CONSOLE
from electro.warnings import WARNINGS
from electro.paths import PATH_THEMES, PATH_JS, PATH_SEARCH_RESULTS_MD
from electro.html_snippets import build_snippet_notice_start, SNIPPET_NOTICE_END
from electro.simplepack import simplepack
from electro.inline_images import make_html_images_inline
from electro.inline_fonts import make_html_fonts_inline
from electro.inline_icons import make_html_icons_inline

print = CONSOLE.print
MAX_HEADING_DEPTH = 6


def build_project(path_build) -> Result[str, str]:
    if not path_build.exists():
        # -------------------------
        # Bad path passed
        # -------------------------
        return Err(
            f'Path "{path_build}" does not exist. Expected a path to an electro project directory'
            ' or to an electro project file (i.e. ".json" file).'
        )
    if path_build.is_dir():
        # -------------------------
        # Directory passed
        # -------------------------
        path_project_directory = path_build
        path_project_file = path_project_directory / Path(CONFIG['project_filename'])
        if not path_project_file.exists():
            return Err(f'Project file {path_project_file} not found.')
    else:
        # -------------------------
        # Project file path passed
        # -------------------------
        path_project_file = path_build
        if path_project_file.suffix != '.json':
            return Err(f'Expected project file ("{path_project_file}") to have a ".json" extension.')
        path_project_directory = path_project_file.parent

    with open(path_project_file, 'r') as file:
        project_config = json.load(file)
    CONFIG['project_config'] = project_config
    logger.info(f'Project Config:\n{pformat(project_config)}')

    # -----------------------
    # Determine output format
    # -----------------------
    CONFIG['output_format'] = project_config.get('output_format', 'static_site')
    if 'pack' in project_config:
        WARNINGS.warning(
            'The "pack" option is deprecated. Use "output_format" instead, and specify'
            f' one of: {OUTPUT_FORMATS}'
        )
        if project_config.get('pack', False):
            WARNINGS.warning(
                'Implicitly interpreting deprecated `"pack": true` option as'
                ' `"output_format": "single_file"`.'
            )
            CONFIG['output_format'] = 'single_file'
    if CONFIG['output_format'] not in OUTPUT_FORMATS:
        return Err(
            f'Project file specified an output_format of "{CONFIG["output_format"]}". '
            f'Expected one of: {OUTPUT_FORMATS}.'
        )

    # -----------------------
    # Get options
    # -----------------------
    CONFIG['enable_newline_to_break'] = project_config.get('enable_newline_to_break', False)

    # -----------------------
    # Determine site dir
    # -----------------------
    result = get_deprecated(project_config, 'output_directory', 'site_directory')
    if isinstance(result, Err):
        return result
    path_output_directory = path_project_directory / Path(result.value)
    if not path_output_directory.is_dir():
        return Err(f'Site directory {path_output_directory} does not exist.')
    CONFIG['path_project_directory'] = path_project_directory
    CONFIG['path_output_directory'] = path_output_directory

    # -----------------------
    # Determine template dir
    # -----------------------
    path_theme_directory = PATH_THEMES / project_config['theme']
    if not path_theme_directory.is_dir():
        return Err(f'Theme directory {path_theme_directory} does not exist.')
    CONFIG['path_theme_directory'] = path_theme_directory

    # -----------------------
    # Build menu and pase markdown
    # -----------------------
    builder = SiteBuilder()
    for navigation_descriptor in project_config['navigation']:
        result = builder.add_navigation_descriptor(navigation_descriptor)
        if isinstance(result, Err):
            return result
    result = builder.render_site()
    if isinstance(result, Err):
        return result

    if CONFIG['output_format'] == 'single_file':
        pack_site(path_output_directory)
    return Ok()


def pack_site(path_output_directory):
    print("Packing...")

    path_file = path_output_directory / Path('index.raw.html')
    path_file_stage1 = path_output_directory / Path("index.packed.stage1.html")
    path_file_stage2 = path_output_directory / Path("index.packed.stage2.html")
    path_file_stage3 = path_output_directory / Path("index.packed.stage3.html")
    path_file_packed = path_output_directory / Path("index.html")
    print(f'packing {path_file.name} to {path_file_packed}...')
    simplepack(path_file, path_file_stage1, uglify=False)
    print(f'Inlining images to: {path_file_stage2}...')
    make_html_images_inline(str(path_file_stage1), str(path_file_stage2))
    print(f'Inlining html fonts to: {path_file_stage3}...')
    make_html_fonts_inline(path_file_stage2, path_file_stage3)
    print(f'Inlining html icons to: {path_file_packed}...')
    make_html_icons_inline(path_file_stage3, path_file_packed)
    print("Packing complete.")


class SiteBuilder:
    def __init__(self):
        self.menu_html = ''
        self.site_documents = {}
        self.search_index = {
            "config": {
                "lang": ["en"],
                "min_search_length": 3,
                "prebuild_index": False,
                "separator": r"[\s\-]+",
            },
            "docs": [],
        }
        self.substitutions = {}
        self.menu_builder = MenuBuilder()

    def add_navigation_descriptor(self, navigation_descriptor) -> Result[str, str]:
        section_name = navigation_descriptor.get('section')
        self.menu_builder.add_section(section_name)
        self.menu_html += '<ul class="menu-tree">\n'
        documents_dict = navigation_descriptor.get('documents')
        if documents_dict is None:
            return Err(f'No "documents" key in navigation descriptor {navigation_descriptor}.')
        for menu_name, md_document_name in documents_dict.items():
            document_name = md_document_name_to_document_name(md_document_name)
            path_markdown = CONFIG['path_project_directory'] / Path('docs') / Path(md_document_name)
            result = self.build_document(path_markdown, document_name)
            if isinstance(result, Err):
                return result
            link_url = f"{document_name}.html" if CONFIG['output_format'] == 'static_site' else None
            self.menu_builder.add_item(0, menu_name, link_url=link_url, document_name=document_name)
            self._build_subheading_menus(document_name)
        return Ok()


    def _build_subheading_menus(self, document_name):
        document_html = self.site_documents[document_name]['html']
        soup = BeautifulSoup(document_html, 'lxml')
        for heading in soup.find_all(['h2', 'h3']):
            html_tag = heading.name
            heading_text = heading.text.strip()
            heading_id = heading_text_to_id(heading_text)
            level = int(html_tag[1]) - 1
            link_url = f"{document_name}.html#{heading_id}" if CONFIG['output_format'] == 'static_site' else None
            self.menu_builder.add_item(level, heading_text, link_url=link_url, heading_id=heading_id)

    def build_document(self, path_markdown, document_name) -> Result[str, str]:
        if not path_markdown.exists():
            return Err(f'Source markdown document {path_markdown} does not exist.')
        with open(path_markdown, 'r') as file:
            document_markdown = file.read()

        # --------------------
        # Pre-parser
        # --------------------
        result = self.pre_parse_markdown(document_markdown)
        if isinstance(result, Err):
            return result
        document_markdown = result.value

        # --------------------
        # Render Markdown
        # --------------------
        extensions = [
            'tables',
            'fenced_code',
            'electro.mdx_urlize:UrlizeExtension',
            'codehilite',
            'attr_list',
        ]
        if CONFIG['enable_newline_to_break']:
            # Newlines in markdown will be interpreted as line breaks.
            extensions.append('nl2br')
        document_html = markdown.markdown(document_markdown, extensions=extensions,)

        # --------------------
        # Pre-parser
        # --------------------
        document_html = self.post_parse_html(document_html)

        # ---------------------
        # Modify HTML
        # ---------------------

        # Fix inter-document links
        inter_document_links = re.findall(r'<a href="\S*.md(?:\#\S*)?">', document_html)
        for link in list(set(inter_document_links)):
            document_html = document_html.replace(link, link.replace('.md', '.html'))

        # Wrap images
        img_tags = re.findall(r'<img .*?>', document_html)
        for img_tag in list(set(img_tags)):
            document_html = document_html.replace(
                img_tag, f'<div class="img-wrapper">{img_tag}</div>'
            )

        # Add id tags to headings
        headings = re.findall(r'<h\d>.*<\/h\d>', document_html)
        # print(md_document_name)
        for heading in headings:
            core = heading[4:-5]
            tag_start = heading[:3]
            _id = heading_text_to_id(core)
            replacement = heading.replace(tag_start, f'{tag_start} id="{_id}"')
            # print(f'   {heading}')
            # print(f'      {core}')
            # print(f'      {tag_start}')
            # print(f'         {_id}')
            # print(f'         {replacement}')
            document_html = document_html.replace(heading, replacement)

        # Add footer text
        result = get_deprecated(CONFIG['project_config'], 'footer', 'copyright')
        if isinstance(result, Err):
            return result
        if result.value:
            document_html += '<hr />\n' f'<div class="footer">{result.value}</div>'

        # ---------------------
        # Search
        # ---------------------
        if path_markdown.name not in CONFIG['project_config'].get('exclude_from_search', []):
            self.add_document_to_search(document_name, document_html)

        self.site_documents[document_name] = {'path_markdown': path_markdown, 'html': document_html}
        return Ok()

    def pre_parse_markdown(self, markdown) -> Result[str, str]:
        project_config = CONFIG['project_config']
        markdown = self._fix_bullet_list_starts(markdown)
        if project_config.get('strip_frontmatter', False):
            markdown = self._strip_frontmatter(markdown)
        if project_config.get('number_headings', False):
            at_level = project_config.get('number_headings_at_level', 1)
            markdown = add_heading_numbers(markdown, at_level=at_level)
        markdown = self._parse_replacements(markdown)
        result = self._parse_notices(markdown)
        if isinstance(result, Err):
            return result
        markdown = result.value
        markdown = self._parse_experimental(markdown)
        return Ok(markdown)

    def _fix_bullet_list_starts(self, markdown):
        """Fix bullet lists by injecting a blank line if previous line was text.

        Many markdown engines recognize a line starting with "- " or "* " as an
        unordered list line, even if the preceding line was paragraph text. The
        python `markdown` package does not.  We inject a blank line where needed
        to force lists to be recognized.

        The generally used VSCode markdown preview extension ("Markdown Preview Github Styling")
        treats all "- " and "* " prefixed lines as lists, so matching that behavior here 
        ensures that people composing in VSCode will get the same output they see in
        VSCode's previewer.
        """
        previous_was_list = False
        previous_was_blank = True
        out_lines = []
        for line in markdown.splitlines():
            stripped = line.strip()
            is_list = (stripped.startswith('- ') or stripped.startswith('* '))
            if is_list and not previous_was_list and not previous_was_blank:
                # Insert a blank line to force this list to be recognized.
                out_lines.append('')
            out_lines.append(line)
            previous_was_list = is_list
            previous_was_blank = not bool(stripped)
        return '\n'.join(out_lines)

    def _strip_frontmatter(self, markdown):
        found_start = False
        found_end = False
        out_lines = []
        for line in markdown.splitlines():
            if not found_start:
                if re.match(r'^---\s*$', line):
                    found_start = True
                    continue
            elif not found_end:
                if re.match(r'^---\s*$', line):
                    found_end = True
                continue
            out_lines.append(line)
        return '\n'.join(out_lines)

    def _parse_replacements(self, markdown):
        replacements = CONFIG['project_config'].get('replacements', ())
        for replacement in replacements:
            markdown = markdown.replace(replacement['find'], replacement['replace'])
        return markdown

    def _parse_notices(self, markdown) -> Result[str, str]:
        notice_start_types = re.findall(r'{{% notice (\S*) %}}', markdown)
        logger.debug(f'{notice_start_types=}')
        for notice_start_type in notice_start_types:
            index = str(len(self.substitutions))
            html_temporary = f'<div class="PRE-PARSER-SUBSTITUTION-{index}"></div>'
            result = build_snippet_notice_start(notice_start_type)
            if isinstance(result, Err):
                return result
            substitution = result.value
            self.substitutions[html_temporary] = substitution
            markdown = markdown.replace(
                r'{{% notice ' + notice_start_type + r' %}}', html_temporary
            )

        NOTICE_END_ITEM = r'{{% /notice %}}'
        if NOTICE_END_ITEM in markdown:
            index = str(len(self.substitutions))
            html_temporary = f'<div class="PRE-PARSER-SUBSTITUTION-{index}"></div>'
            substitution = SNIPPET_NOTICE_END
            self.substitutions[html_temporary] = substitution
            markdown = markdown.replace(NOTICE_END_ITEM, html_temporary)
        return Ok(markdown)

    def _parse_experimental(self, markdown):
        index = str(len(self.substitutions))
        html_temporary = f'<div class="PRE-PARSER-SUBSTITUTION-{index}"></div>'
        substitution = '<div class="change-bar">'
        self.substitutions[html_temporary] = substitution
        markdown = markdown.replace(
                r':change_bar_start', html_temporary
            )

        index = str(len(self.substitutions))
        html_temporary = f'<div class="PRE-PARSER-SUBSTITUTION-{index}"></div>'
        substitution = '</div>'
        self.substitutions[html_temporary] = substitution
        markdown = markdown.replace(
                r':change_bar_end', html_temporary
            )
        return markdown

    def post_parse_html(self, html):
        for text_old, text_new in self.substitutions.items():
            logger.debug(f'Subsitituting: {text_old=} {text_new=}')
            html = html.replace(text_old, text_new)
        return html

    def add_document_to_search(self, document_name, document_html):
        logger.debug(f'add_document_to_search(): {document_name}')
        soup = BeautifulSoup(document_html, 'lxml')
        document_title = None
        for element in soup.find_all(["h1"]):
            document_title = element.text.strip()
            break
        if document_title is None:
            WARNINGS.warning(
                f'No h1 tag found in {document_name}. Cannot extract document title for search.'
            )
            document_title = '(Unknown)'
        base_location = f'{document_name}.html'

        current_location = base_location
        current_heading_text = None
        section_text = ''
        for element in soup.find_all(['h2', 'h3', 'p', 'li', 'th', 'td']):
            if element.name in ['p', 'li', 'th', 'td']:
                # Paragraph, list, table
                section_text += (' ' if section_text else '') + element.text.strip()
            else:
                # Heading
                if section_text:
                    self._add_search_item(
                        document_title, current_location, current_heading_text, section_text
                    )
                    section_text = ''
                current_heading_text = element.text.strip()
                heading_id = heading_text_to_id(current_heading_text)
                logger.debug(f'   {element.name} : {current_heading_text} : {heading_id}')
                current_location = f'{base_location}#{heading_id}'
        # Commit any remaining text
        if section_text:
            self._add_search_item(
                document_title, current_location, current_heading_text, section_text
            )

    def _add_search_item(self, title, location, heading, text):
        doc_descriptor = {'title': title, 'location': location, 'heading': heading, 'text': text}
        self.search_index['docs'].append(doc_descriptor)

    def _render_document(self, template_html, path_document_out, content_html, document_name) -> Result[str, str]:
        print(f'Building {path_document_out}...')

        project_config = CONFIG['project_config']

        document_html = template_html.replace(r'{{% content %}}', content_html)

        # Items replaced here will also target user content, since user content has been merged by
        # now.
        result = get_deprecated(project_config, 'master_title', 'site_name', required=True)
        if isinstance(result, Err):
            return result
        document_html = document_html.replace(r'{{% master_title %}}', result.value)
        document_html = document_html.replace(r'{{% sidebar_menu %}}', self.menu_html)
        document_html = document_html.replace(r'{{% current_document_name %}}', document_name)
        # NOTE: We do a weird thing here. Note that the text we are replacing INCLUDES the single
        #       quotes surrounding it.  That violates the cleanliness of how template substitution
        #       works, but allows us to replace the value in the template with a json boolean.
        #
        document_html = document_html.replace(
            r"'{{% single_file %}}'", to_json_bool(CONFIG['output_format'] == 'single_file')
        )
        document_html = document_html.replace(
            r'{{% watermark %}}', project_config.get("watermark", "")
        )
        document_html = document_html.replace(r'{{% electro_version %}}', CONFIG['version'])
        document_html = document_html.replace(
            r'{{% timestamp %}}', datetime.now().astimezone().replace(microsecond=0).isoformat()
        )
        document_html = document_html.replace(r'{{% year %}}', str(date.today().year))

        with open(path_document_out, 'w') as file:
            file.write(document_html)
        return Ok()

    def render_site(self) -> Result[str, str]:
        project_config = CONFIG['project_config']
        if project_config.get('level_1_headings_are_document_titles', False):
            self.menu_builder.cull_items_above(1)
        else:
            self.menu_builder.cull_items_below(1)
        self.menu_builder.dump()
        menu_html = self.menu_builder.render_html()
        # with open('DEBUG_menu_html.html', 'w') as file:
        #     file.write(menu_html)
        self.menu_html = menu_html

        path_output_directory = CONFIG['path_output_directory']
        path_theme_directory = CONFIG['path_theme_directory']
        path_project_directory = CONFIG['path_project_directory']

        # -------------------
        # Copy CSS
        # -------------------
        path_css_source = path_theme_directory / Path('style.css')
        path_css_destination = path_output_directory / Path('style.css')
        shutil.copy(path_css_source, path_css_destination)
        path_css_source = path_theme_directory / Path('fonts.css')
        path_css_destination = path_output_directory / Path('fonts.css')
        shutil.copy(path_css_source, path_css_destination)
        path_css_source = path_theme_directory / Path('fontawesome.css')
        path_css_destination = path_output_directory / Path('fontawesome.css')
        shutil.copy(path_css_source, path_css_destination)

        # -------------------
        # Copy CSS overlay
        # -------------------
        path_css_destination = path_output_directory / Path('overlay.css')
        path_css_source = path_project_directory / Path('docs') / Path('overlay.css')
        if not path_css_source.exists():
            path_css_source = path_theme_directory / Path('overlay.css')
        shutil.copy(path_css_source, path_css_destination)
        # Append customizations to end of CSS overlay
        append_css_customizations(path_css_destination)

        # -------------------
        # Copy Images
        # -------------------
        path_image_source_dir = path_project_directory / Path('docs') / Path('img')
        path_image_destination_dir = path_output_directory / Path('img')
        copy_directory_contents(path_image_source_dir, path_image_destination_dir)

        # -------------------
        # Copy Fonts
        # -------------------
        path_fonts_source_dir = path_theme_directory / Path('fonts')
        path_image_destination_dir = path_output_directory / Path('fonts')
        copy_directory_contents(path_fonts_source_dir, path_image_destination_dir)

        # -------------------
        # Copy Attachments
        # -------------------
        path_attachment_source_dir = path_project_directory / Path('docs') / Path('attachments')
        path_attachments_destination_dir = path_output_directory / Path('attachments')
        copy_directory_contents(path_attachment_source_dir, path_attachments_destination_dir)

        # -------------------
        # Copy Favicon
        # -------------------
        path_favicon_source = path_theme_directory / Path('favicon.ico')
        path_favicon_destination = path_output_directory / Path('img') / Path('favicon.ico')
        shutil.copy(path_favicon_source, path_favicon_destination)

        # -------------------
        # Copy js
        # -------------------
        path_js_resource_source_dir = PATH_JS
        path_js_destination_dir = path_output_directory / Path('js')
        copy_directory_contents(path_js_resource_source_dir, path_js_destination_dir)

        path_js_theme_source_dir = path_theme_directory / Path('js')
        copy_directory_contents(path_js_theme_source_dir, path_js_destination_dir)

        # -------------------
        # Build search results doc
        # -------------------
        result = self.build_document(PATH_SEARCH_RESULTS_MD, 'search')
        if isinstance(result, Err):
            return result

        # -------------------
        # Build site pages
        # -------------------
        path_template = path_theme_directory / Path('template.html')
        with open(path_template, 'r') as file:
            template_html = file.read()

        if CONFIG['output_format'] == 'single_file':
            # -------------------
            # Single-file document
            # -------------------
            pages_html = ""
            style_html = ""
            for document_name, document_info in self.site_documents.items():
                pages_html += f'<div class="content-page" id="{document_name}" {style_html}>'
                # pages_html += f'(content from {document_name})<br><br>'
                pages_html += document_info['html']
                pages_html += '</div>'
                # Start all subsequent pages as hidden
                style_html = 'style="display: none"'
            path_site_document = path_output_directory / Path('index.raw.html')
            result = self._render_document(template_html, path_site_document, pages_html, "Document")
            if isinstance(result, Err):
                return result
        else:
            # -------------------
            # Static site
            # -------------------
            for document_name, document_info in self.site_documents.items():
                path_site_document = path_output_directory / Path(f'{document_name}.html')
                result = self._render_document(
                    template_html, path_site_document, document_info['html'], document_name
                )
                if isinstance(result, Err):
                    return result

        # -------------------
        # Save search index
        # -------------------
        path_search_directory = path_output_directory / Path('search')
        path_search_directory.mkdir(parents=True, exist_ok=True)
        path_search_index = path_search_directory / Path('search_index.js')
        # TODO: Remove this JSON indent after everything is working?
        search_js = "App.searchData = " + json.dumps(self.search_index, indent=4)
        with open(path_search_index, 'w') as file:
            file.write(search_js)
        return Ok()


MAX_MENU_DEPTH = 3


class MenuItem:
    def __init__(self, display_text, link_url, document_name, heading_id):
        self.display_text = display_text
        self.link_url = link_url
        self.document_name = document_name
        self.heading_id = heading_id
        self.children = []


class MenuSection:
    def __init__(self, display_text):
        self.display_text = display_text
        # self.document_name = document_name
        self.last_child_at_level = [None] * MAX_MENU_DEPTH
        self.children = []

    def add(self, level, display_text, heading_id, link_url, document_name):
        """Add a menu item.

        level is 0 based.
        """
        new_item = MenuItem(display_text, link_url, document_name, heading_id)
        if level == 0:
            self.children.append(new_item)
            self.last_child_at_level[0] = new_item
        else:
            parent = self.last_child_at_level[level - 1]
            parent.children.append(new_item)
            # Clear "last child" of all levels deeper than this one.
            # NOTE: This is not strictly necessary, but it will
            #       defensively keep us from creating a weird tree if the
            #       input is badly formed.
            for i in range(level + 1, MAX_MENU_DEPTH):
                self.last_child_at_level[i] = None
            self.last_child_at_level[level] = new_item
        pass


class MenuBuilder:
    def __init__(self):
        self.sections = []
        self.current_document_name = None

    def add_section(self, display_text):
        section = MenuSection(display_text)
        self.sections.append(section)

    def add_item(self, level, display_text, heading_id=None, link_url=None, document_name=None):
        if document_name:
            self.current_document_name = document_name
        section = self.sections[-1]
        section.add(level, display_text, heading_id, link_url, self.current_document_name)

    def dump(self, display=False):
        for section in self.sections:
            self._dump_recursive(section, display)

    def _dump_recursive(self, node, display, level=0):
        """Dump node and children recursively.

        Args:
            node (MenuItem or MenuSection): A tree node
            display (bool): If True, dump to screen. Will always dump to log.
        """
        aux_data = {}
        # if isinstance(node, MenuSection):
        #     aux_data['document_name'] = node.document_name
        if isinstance(node, MenuItem):
            aux_data['document_name'] = node.document_name
            aux_data['link_url'] = node.link_url
            aux_data['heading_id'] = node.heading_id
        indent = '   ' * level
        text = f'{indent}🟡 "{node.display_text}" :: {aux_data}'
        logger.debug(text)
        if display:
            print(text)
        for child in node.children:
            self._dump_recursive(child, display, level + 1)

    def cull_items_above(self, level):
        """Remove all menu items ABOVE level (i.e. at an indent LESS than level).
        
        NOTE: level is the "item" level, and does not include the section. 
        """
        logger.info(f"cull_items_above(): {level}")
        for section in self.sections:
            # NOTE: level is the "item" level depth, and does not include the section, but we
            #       are recursing a tree where level 0 is the section node, so we add
            #       1 to the passed in level.
            section.children = self._cull_items_above_recursive(level + 1, 0, section)

    def _cull_items_above_recursive(self, cull_level, current_level, node):
        if cull_level == current_level + 1:
            return node.children
        retained_items = []
        for child in node.children:
            retained_items += self._cull_items_above_recursive(cull_level, current_level + 1, child)
        return retained_items

    def cull_items_below(self, level):
        """Remove all menu items BELOW level (i.e. at an indent GREATER than level).
        
        NOTE: level is the "item" level, and does not include the section. 
        """
        logger.info(f"cull_items_below(): {level}")
        for section in self.sections:
            # NOTE: level is the "item" level depth, and does not include the section, but we
            #       are recursing a tree where level 0 is the section node, so we add
            #       1 to the passed in level.
            self._cull_items_below_recursive(level + 1, 0, section)

    def _cull_items_below_recursive(self, cull_level, current_level, node):
        if cull_level == current_level:
            node.children = []
            return

        for child in node.children:
            self._cull_items_below_recursive(cull_level, current_level + 1, child)

    def render_html(self):
        html = ''
        for section in self.sections:
            html += self._render_html_section(section)
        return html

    def _render_html_section(self, section):
        lines = []
        if section.display_text:
            lines.append(f'<div class="section-heading">{section.display_text}</div>')
        lines.append('<ul class="menu-tree">')
        lines += self._render_html_lines_children(0, section.children)
        lines.append('</ul>')
        return '\n'.join(lines)

    def _render_html_lines_children(self, level, children):
        lines = []
        for child in children:
            # Submenu items
            if level == 0 and child.children:
                submenu_lines = (
                    ['<ul class="nested">']
                    + self._render_html_lines_children(level + 1, child.children)
                    + ['</ul>']
                )
                caret_visible = True
            else:
                submenu_lines = []
                caret_visible = False
            classes = ['level-0'] if level == 0 else ['no-child']
            class_list = ' '.join(classes)
            class_statement = f'class="{class_list}"'
            heading_id_statement = (
                f'data-target-heading-id={child.heading_id}' if child.heading_id else ''
            )

            lines.append('<li>')
            lines.append(
                '<span'
                f' {class_statement}'
                f' id="menuitem_doc_{child.document_name}"'
                f' data-document-name="{child.document_name}"'
                f' {heading_id_statement}'
                '>'
            )
            lines.append(
                format_menu_heading(
                    child.display_text,
                    include_caret_space=(level == 0),
                    caret_visible=caret_visible,
                    link_url=child.link_url,
                    is_level_two=(level > 0)
                )
            )
            lines.append('</span>')
            lines += submenu_lines
            lines.append('</li>')
        return lines


def md_document_name_to_document_name(md_document_name):
    return Path(md_document_name).stem


def to_json_bool(python_bool):
    """Return a json compliant boolean string (given a python bool)."""
    return 'true' if python_bool else 'false'


def format_menu_heading(
    text,
    include_caret_space=False,
    caret_visible=False,
    link_url=None,
    is_level_two=False,
):
    """Given a menu heading, split it into two divs if it has a numeric prefix.

    For headings that start with a section number (e.g. "1.5 Study Results") we
    will split the heading into two pieces and wrap them each in a div.
    """
    caret_item_content = ''
    number_item_content = ''
    text_item_content = text
    core_content = ''

    if include_caret_space:
        caret_item_content = (
            '<i class="caret fa fa-angle-right"></i>'
            if caret_visible
            else '<i class="caret-placeholder"></i>'
        )

    NBSP = "\xa0"
    text = text.replace(NBSP, " ")
    if " " in text:
        pieces = split_if_numbered(text)
        if pieces:
            number_item_content, text_item_content = pieces

    if number_item_content:
        classes = "number-item" + (" level-two" if is_level_two else "")
        number_item_content = f'<div class="{classes}">{number_item_content}</div>'
    text_item_content = f'<div class="text-item">{text_item_content}</div>'
    core_content = f'<div class="core">{number_item_content}{text_item_content}</div>'
    if link_url:
        core_content = f'<a href="{link_url}">{core_content}</a>'

    html = f'<div class="menu-item-container">{caret_item_content}{core_content}</div>'

    logger.debug(f'format_menu_heading: result: "{html}"')
    return html


def split_if_numbered(text):
    """Given a menu heading, split it into two strings if it has a numeric prefix.

    For headings that start with a section number (e.g. "1.5 Study Results") we
    will split the heading into two pieces and return a tuple.

    Otherwise return None
    """
    pieces = text.split()
    heading_number = pieces[0]
    result = None
    if re.match(r'^[\d\.]+$', heading_number):
        remaining = ' '.join(pieces[1:])
        result = (heading_number, remaining)
    logger.debug(f'split_if_numbered(): {heading_number=} {text=} => {result}')
    return result


def heading_text_to_id(text):
    NBSP = '\u00a0'
    original_text = text
    _id = ''
    dash_appended = False
    text = text.replace('&nbsp;', ' ')
    text = text.replace(NBSP, ' ')
    # NOTE: This transformation is subtle. We need heading_text_to_id() to output identical
    #       id strings for both "raw" heading text (parsed from markdown text directly) and
    #       For heading text which has already been pre-processed by the markdown converter.
    #       The markdown converter changes '&' to '&amp;', so in order to make both ids
    #       Identical we change it BACK to '&', which will then get DROPPED by the
    #       ID conversion.
    text = text.replace('&amp;', '&')
    # Replace decimal with dashes so that heading numbers like "3.12" vs "31.2" remain
    # unique.
    text = text.replace('.', '-')
    for char in text.lower():
        if char == ' ':
            if not dash_appended:
                _id += '-'
                dash_appended = True
        elif re.match(r'[a-z0-9\-]', char):
            _id += char
            dash_appended = False

    # Combine multiple dashes into single dash
    _id = re.sub(r'\-+', '-', _id)

    logger.debug(f'heading_text_to_id() "{original_text}" -> "{_id}"')
    return _id


def copy_directory_contents(source_directory, target_directory):
    logger.debug(f'Copying directory "{source_directory}" to "{target_directory}"')
    target_directory.mkdir(parents=True, exist_ok=True)
    paths_source_files = source_directory.glob('*')
    for path_source_file in sorted(list(paths_source_files)):
        logger.debug(f'   {path_source_file.name}')
        path_destination_file = target_directory / Path(path_source_file.name)
        shutil.copy(path_source_file, path_destination_file)


def add_heading_numbers(markdown, at_level=1):
    """Add heading numbers to a markdown file.

    Args:
        markdown (str): Markdown text

    Returns:
        out_lines (list of strings): List of the lines in the new markdown file.
    """

    heading_manager = HeadingManager(at_level)
    out_lines = []
    for line in markdown.splitlines():
        if not line.startswith("#"):
            out_lines.append(line)
            continue
        pieces = line.split()
        level = len(pieces[0])
        if level < at_level:
            out_lines.append(line)
            continue
        heading_number_text = heading_manager.get(level)
        heading_text = " ".join(pieces[1:])
        line = f'{pieces[0]} {heading_number_text}&nbsp;&nbsp;&nbsp;&nbsp;' f'{heading_text}'
        out_lines.append(line)
    return '\n'.join(out_lines)


class HeadingManager:
    """Assign heading numbers to document headings.

    This class is a tool which returns an assigned heading number string (of the form
    "1.2.3" etc.) to the headings of a markdown document.  The class maintains a count of the
    current heading number for all indent depths.  Calling .get(level) for each document
    heading in sequence, from document top to bottom, returns the appropriate heading number
    for that heading, and updates the counters appropriately.
    """

    def __init__(self, at_level):
        """Initialize."""
        self.at_level = at_level
        self.heading_number = {level: 0 for level in range(1, MAX_HEADING_DEPTH + 1)}

    def get(self, level):
        """Get the heading number for the current heading (with depth of level)."""

        # Bump this heading level
        self.heading_number[level] += 1
        # Reset all deeper heading levels
        for i in range(level + 1, MAX_HEADING_DEPTH):
            self.heading_number[i] = 0

        # Build heading number text
        digits = [str(self.heading_number[i]) for i in range(self.at_level, level + 1)]
        return ".".join(digits)


def append_css_customizations(path_css_overlay):
    new_lines = []
    project_config = CONFIG['project_config']

    width = project_config.get('menu_level_two_number_prefix_width', None)
    if width:
        new_lines += [
            '.menu-item-container .core .number-item.level-two {',
            f'   width: {width};',
            '}',
        ]
    if new_lines:
        new_lines = [''] + new_lines + ['']
        text = '\n'.join(new_lines)
        with open(path_css_overlay, 'a') as file:
            file.write(text)

def get_deprecated(config_dict, key, deprecated_key, default=None, required=True) -> Result[str, str]:
    if key in config_dict and deprecated_key in config_dict:
        return Err(
            f'Key "{key}" and deprecated key "{deprecated_key}" both present in config.'
            ' Remove deprecated key.')
    if deprecated_key in config_dict:
        WARNINGS.warning(f'Key "{deprecated_key}" has been deprecated.  Use "{key}" instead.')
        return Ok(config_dict[deprecated_key])
    if required and key not in config_dict:
        return Err(f'Required key "{key}" not present in config.')
    return Ok(str(config_dict.get(key, default)))
