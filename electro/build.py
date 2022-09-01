# Standard library
from pathlib import Path
import json
import shutil
import shutil
import re
from textwrap import indent
from datetime import datetime, date

# Library
from prettyprinter import pformat
import markdown
from bs4 import BeautifulSoup

# Local
from loguru import logger
from pytest import mark
from electro.app_config import CONFIG, OUTPUT_FORMATS
from electro.faults import FAULTS
from electro.paths import PATH_THEMES, PATH_JS, PATH_SEARCH_RESULTS_MD
from electro.html_snippets import build_snippet_notice_start, SNIPPET_NOTICE_END
from electro.simplepack import simplepack
from electro.inline_images import make_html_images_inline
from electro.inline_fonts import make_html_fonts_inline
from electro.inline_icons import make_html_icons_inline

pprint = CONFIG['console_pprint']


def build_project(path_build):
    if not path_build.exists():
        # -------------------------
        # Bad path passed
        # -------------------------
        FAULTS.error(
            f'Path "{path_build}" does not exist. Expected a path to an electro project directory'
            ' or to an electro project file (i.e. ".json" file).'
        )
        return
    if path_build.is_dir():
        # -------------------------
        # Directory passed
        # -------------------------
        path_project_directory = path_build
        path_project_file = path_project_directory / Path(CONFIG['project_filename'])
        if not path_project_file.exists():
            FAULTS.error(f'Project file {path_project_file} not found.')
            return
    else:
        # -------------------------
        # Project file path passed
        # -------------------------
        path_project_file = path_build
        if path_project_file.suffix != '.json':
            FAULTS.error(
                f'Expected project file ("{path_project_file}") to have a ".json" extension.'
            )
            return
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
        FAULTS.warning(
            'The "pack" option is deprecated. Use "output_format" instead, and specify'
            f' one of: {OUTPUT_FORMATS}'
        )
        if project_config.get('pack', False):
            FAULTS.warning(
                'Implicitly interpreting deprecated `"pack": true` option as'
                ' `"output_format": "single_file"`.'
            )
            CONFIG['output_format'] = 'single_file'
    if CONFIG['output_format'] not in OUTPUT_FORMATS:
        FAULTS.error(
            f'Project file specified an output_format of "{CONFIG["output_format"]}". '
            f'Expected one of: {OUTPUT_FORMATS}.'
        )
        return

    # -----------------------
    # Get options
    # -----------------------
    CONFIG['enable_newline_to_break'] = project_config.get('enable_newline_to_break', False)

    # -----------------------
    # Determine site dir
    # -----------------------
    path_site_directory = path_project_directory / Path(project_config['site_directory'])
    if not path_site_directory.is_dir():
        FAULTS.error(f'Site directory {path_site_directory} does not exist.')
        return
    CONFIG['path_project_directory'] = path_project_directory
    CONFIG['path_site_directory'] = path_site_directory

    # -----------------------
    # Determine template dir
    # -----------------------
    path_theme_directory = PATH_THEMES / project_config['theme']
    if not path_theme_directory.is_dir():
        FAULTS.error(f'Theme directory {path_theme_directory} does not exist.')
        return
    CONFIG['path_theme_directory'] = path_theme_directory

    # print(f'build_project() {path_project_directory=}')
    # pprint(CONFIG)

    # -----------------------
    # Build menu and pase markdown
    # -----------------------
    builder = Builder()
    for navigation_descriptor in project_config['navigation']:
        builder.add_navigation_descriptor(navigation_descriptor)
    builder.render_site()

    if CONFIG['output_format'] == 'single_file':
        pack_site(path_site_directory)


def pack_site(path_site_directory):
    print("Packing...")

    path_file = path_site_directory / Path('index.raw.html')
    path_file_stage1 = path_site_directory / Path("index.packed.stage1.html")
    path_file_stage2 = path_site_directory / Path("index.packed.stage2.html")
    path_file_stage3 = path_site_directory / Path("index.packed.stage3.html")
    path_file_packed = path_site_directory / Path("index.html")
    print(f'packing {path_file.name} to {path_file_packed}...')
    simplepack(path_file, path_file_stage1, uglify=False)
    print(f'Inlining images to: {path_file_stage2}...')
    make_html_images_inline(str(path_file_stage1), str(path_file_stage2))
    print(f'Inlining html fonts to: {path_file_stage3}...')
    make_html_fonts_inline(path_file_stage2, path_file_stage3)
    print(f'Inlining html icons to: {path_file_packed}...')
    make_html_icons_inline(path_file_stage3, path_file_packed)
    print("Packing complete.")


class Builder:
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

    def add_navigation_descriptor(self, navigation_descriptor):
        if section_name := navigation_descriptor.get('section'):
            self.menu_html += f'<div class="section-heading">{section_name}</div>\n'
        self.menu_html += '<ul class="menu-tree">\n'
        documents_dict = navigation_descriptor.get('documents')
        if documents_dict is None:
            FAULTS.error(f'No "documents" key in navigation descriptor {navigation_descriptor}.')
            return
        for menu_name, md_document_name in documents_dict.items():
            document_name = md_document_name_to_document_name(md_document_name)
            path_markdown = CONFIG['path_project_directory'] / Path('docs') / Path(md_document_name)
            self.build_document(path_markdown, document_name)
            subheading_menu_html = self.build_subheading_menu_html(document_name)
            if subheading_menu_html:
                # caret_str = '<i class="caret fa fa-angle-right"></i>'
                caret_visible = True
            else:
                # classes += ' no_child'
                # caret_str = ''
                # caret_str = '<i class="caret-placeholder"></i>'
                caret_visible = False
            # self.menu_html += (
            #     f'<li><span class="{classes}" id="menuitem_doc_{document_name}" data-document-name="{document_name}">'
            #     + f'{caret_str}'
            #     # f'<a href="{document_name}.html">{menu_name}</a>'
            #     # f'{menu_name}'
            #     + format_menu_heading(menu_name)
            #     + f'</span>{subheading_menu_html}</li>\n'
            # )
            self.menu_html += (
                f'<li><span id="menuitem_doc_{document_name}" data-document-name="{document_name}">'
            )
            # pieces = split_if_numbered(menu_name)
            # if pieces:
            #     heading_number, heading_text = pieces
            #     self.menu_html += (
            #         '<div class="menu-text-container">'
            #         f'<div class="menu-left-piece wide">{caret_str}{heading_number}</div>'
            #         f'<div class="menu-right-piece">{heading_text}</div>'
            #         '</div>'
            #     )
            # else:
            # menu_heading_html = format_menu_heading(menu_name, caret_html=caret_str)
            # if CONFIG['output_format'] == 'static_site':
            #     menu_heading_html = f'<a href="{document_name}.html">{menu_heading_html}</a>'
            # self.menu_html += menu_heading_html
            link_url = f"{document_name}.html" if CONFIG['output_format'] == 'static_site' else None
            self.menu_html += format_menu_heading(
                menu_name, include_caret_space=True, caret_visible=caret_visible, link_url=link_url
            )
            self.menu_html += f'</span>{subheading_menu_html}</li>\n'

        self.menu_html += '</ul>\n'

    def build_subheading_menu_html(self, document_name):
        document_html = self.site_documents[document_name]['html']
        soup = BeautifulSoup(document_html, 'lxml')
        menu_html = ''
        for heading in soup.find_all('h2'):
            if not menu_html:
                menu_html = '    <ul class="nested">\n'
            heading_text = heading.text.strip()
            heading_id = heading_text_to_id(heading_text)
            link_url = (
                f'{document_name}.html#{heading_id}'
                if CONFIG['output_format'] == 'static_site'
                else None
            )
            menu_heading_html = format_menu_heading(heading_text, on_nbsp=True, link_url=link_url)
            # if CONFIG['output_format'] == 'static_site':
            #     menu_heading_html = f'<a href="{heading_url}">{menu_heading_html}</a>'
            menu_html += (
                f'        <li><span class="no_child menu-node" data-document-name="{document_name}" data-target-heading-id="{heading_id}">\n'
                + menu_heading_html
                + '</span></li>\n'
            )
        if menu_html:
            menu_html = '\n' + menu_html + '    </ul>\n'
        return menu_html

    def build_document(self, path_markdown, document_name):
        if not path_markdown.exists():
            FAULTS.error(f'Source markdown document {path_markdown} does not exist.')
            return
        with open(path_markdown, 'r') as file:
            document_markdown = file.read()

        # --------------------
        # Pre-parser
        # --------------------
        document_markdown = self.pre_parse_markdown(document_markdown)

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

        # Add copyright text
        if copyright_text := CONFIG['project_config'].get('copyright'):
            document_html += '<hr />\n' f'<div class="copyright">{copyright_text}</div>'

        # ---------------------
        # Search
        # ---------------------
        if path_markdown.name not in CONFIG['project_config'].get('exclude_from_search', []):
            self.add_document_to_search(document_name, document_html)

        self.site_documents[document_name] = {'path_markdown': path_markdown, 'html': document_html}

    def pre_parse_markdown(self, markdown):
        notice_start_types = re.findall(r'{{% notice (\S*) %}}', markdown)
        logger.debug(f'{notice_start_types=}')
        for notice_start_type in notice_start_types:
            index = str(len(self.substitutions))
            html_temporary = f'<div class="PRE-PARSER-SUBSTITUTION-{index}"></div>'
            substitution = build_snippet_notice_start(notice_start_type)
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
            FAULTS.warning(
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

    def _render_document(self, template_html, path_document_out, content_html, document_name):
        print(f'Building {path_document_out}...')

        project_config = CONFIG['project_config']

        document_html = template_html.replace(r'{{% content %}}', content_html)

        # Items replaced here will also target user content, since user content has been merged by
        # now.
        document_html = document_html.replace(r'{{% site_name %}}', project_config['site_name'])
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

    def render_site(self):
        path_site_directory = CONFIG['path_site_directory']
        path_theme_directory = CONFIG['path_theme_directory']
        path_project_directory = CONFIG['path_project_directory']

        # -------------------
        # Copy CSS
        # -------------------
        path_css_source = path_theme_directory / Path('style.css')
        path_css_destination = path_site_directory / Path('style.css')
        shutil.copy(path_css_source, path_css_destination)
        path_css_source = path_theme_directory / Path('fonts.css')
        path_css_destination = path_site_directory / Path('fonts.css')
        shutil.copy(path_css_source, path_css_destination)
        path_css_source = path_theme_directory / Path('fontawesome.css')
        path_css_destination = path_site_directory / Path('fontawesome.css')
        shutil.copy(path_css_source, path_css_destination)

        # -------------------
        # Copy CSS overlay
        # -------------------
        path_css_destination = path_site_directory / Path('overlay.css')
        path_css_source = path_project_directory / Path('docs') / Path('overlay.css')
        if not path_css_source.exists():
            path_css_source = path_theme_directory / Path('overlay.css')
        shutil.copy(path_css_source, path_css_destination)

        # -------------------
        # Copy Images
        # -------------------
        path_image_source_dir = path_project_directory / Path('docs') / Path('img')
        path_image_destination_dir = path_site_directory / Path('img')
        copy_directory_contents(path_image_source_dir, path_image_destination_dir)

        # -------------------
        # Copy Fonts
        # -------------------
        path_fonts_source_dir = path_theme_directory / Path('fonts')
        path_image_destination_dir = path_site_directory / Path('fonts')
        copy_directory_contents(path_fonts_source_dir, path_image_destination_dir)

        # -------------------
        # Copy Attachments
        # -------------------
        path_attachment_source_dir = path_project_directory / Path('docs') / Path('attachments')
        path_attachments_destination_dir = path_site_directory / Path('attachments')
        copy_directory_contents(path_attachment_source_dir, path_attachments_destination_dir)

        # -------------------
        # Copy Favicon
        # -------------------
        path_favicon_source = path_theme_directory / Path('favicon.ico')
        path_favicon_destination = path_site_directory / Path('img') / Path('favicon.ico')
        shutil.copy(path_favicon_source, path_favicon_destination)

        # -------------------
        # Copy js
        # -------------------
        path_js_resource_source_dir = PATH_JS
        path_js_destination_dir = path_site_directory / Path('js')
        copy_directory_contents(path_js_resource_source_dir, path_js_destination_dir)

        path_js_theme_source_dir = path_theme_directory / Path('js')
        copy_directory_contents(path_js_theme_source_dir, path_js_destination_dir)

        # -------------------
        # Build search results doc
        # -------------------
        self.build_document(PATH_SEARCH_RESULTS_MD, 'search')

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
            path_site_document = path_site_directory / Path('index.raw.html')
            self._render_document(template_html, path_site_document, pages_html, "Document")
        else:
            # -------------------
            # Static site
            # -------------------
            for document_name, document_info in self.site_documents.items():
                path_site_document = path_site_directory / Path(f'{document_name}.html')
                self._render_document(
                    template_html, path_site_document, document_info['html'], document_name
                )

        # # TODO: cleanup
        # path_site_document = path_site_directory / Path(f'index.raw.html')
        # # path_site_document = path_site_directory / Path(f'{document_name}.html')
        # document_html = template_html.replace(r'{{% site_name %}}', project_config['site_name'])
        # document_html = document_html.replace(r'{{% sidebar_menu %}}', self.menu_html)
        # # TODO: cleanup
        # # document_html = document_html.replace(r'{{% current_document_name %}}', document_name)
        # document_html = document_html.replace(r'{{% current_document_name %}}', "Document")
        # document_html = document_html.replace(
        #     r'{{% single_file %}}', to_json_bool(project_config.get("pack", False))
        # )
        # document_html = document_html.replace(
        #     r'{{% watermark %}}', project_config.get("watermark", "")
        # )

        # pages_html = ""
        # style_html = ""
        # for document_name, document_info in self.site_documents.items():
        #     pages_html += f'<div class="content-page" id="{document_name}" {style_html}>'
        #     # pages_html += f'(content from {document_name})<br><br>'
        #     pages_html += document_info['html']
        #     pages_html += '</div>'
        #     # Start all subsequent pages as hidden
        #     style_html = 'style="display: none"'

        # document_html = document_html.replace(r'{{% content %}}', pages_html)

        # # Items replaced here will also target user content, since user content has been merged by
        # # now.
        # document_html = document_html.replace(r'{{% electro_version %}}', CONFIG['version'])
        # document_html = document_html.replace(
        #     r'{{% timestamp %}}', datetime.now().astimezone().replace(microsecond=0).isoformat()
        # )
        # document_html = document_html.replace(
        #     r'{{% year %}}', str(date.today().year)
        # )

        # print(f'Building {path_site_document}...')
        # with open(path_site_document, 'w') as file:
        #     file.write(document_html)

        # -------------------
        # Save search index
        # -------------------
        path_search_directory = path_site_directory / Path('search')
        path_search_directory.mkdir(parents=True, exist_ok=True)
        path_search_index = path_search_directory / Path('search_index.js')
        # TODO: Remove this JSON indent after everything is working?
        search_js = "App.searchData = " + json.dumps(self.search_index, indent=4)
        with open(path_search_index, 'w') as file:
            file.write(search_js)


def md_document_name_to_document_name(md_document_name):
    return Path(md_document_name).stem


def to_json_bool(python_bool):
    """Return a json compliant boolean string (given a python bool)."""
    return 'true' if python_bool else 'false'


def format_menu_heading(
    text, on_nbsp=False, include_caret_space=False, caret_visible=False, link_url=None
):
    """Given a menu heading, split it into two divs if it has a numeric prefix.

    For headings that start with a section number (e.g. "1.5 Study Results") we
    will split the heading into two pieces and wrap them each in a div.
    """
    if include_caret_space:
        caret_html = (
            '<i class="caret fa fa-angle-right"></i>'
            if caret_visible
            else '<i class="caret-placeholder"></i>'
        )
    else:
        caret_html = ''

    html = ""

    NBSP = "\xa0"
    if (NBSP in text and on_nbsp) or (" " in text and not on_nbsp):
        temp_text = text.replace(NBSP, " ") if on_nbsp else text
        pieces = split_if_numbered(temp_text)
        if pieces:
            left_content, right_content = pieces

        if link_url:
            right_content = f'<a href="{link_url}">{right_content}</a>'
        html = (
            '<div class="menu-text-container">\n'
            f'<div class="menu-left-piece">{caret_html}{left_content}</div>\n'
            f'<div class="menu-right-piece">{right_content}</div>\n'
            '</div>\n'
        )

    if not html:
        content_html = text
        if link_url:
            content_html = f'<a href="{link_url}">{content_html}</a>'
        html = caret_html + content_html

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
    logger.debug(f'{heading_number=} {text=}')
    if re.match(r'^[\d\.]+$', heading_number):
        remaining = ' '.join(pieces[1:])
        return (heading_number, remaining)
    return None


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
