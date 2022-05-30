# Standard library
from pathlib import Path
import json
import shutil
import shutil
import re
from textwrap import indent

# Library
from prettyprinter import pformat
import markdown
from bs4 import BeautifulSoup

# Local
from loguru import logger
from pytest import mark
from electro.app_config import CONFIG
from electro.faults import FAULTS
from electro.paths import PATH_THEMES, PATH_JS, PATH_SEARCH_RESULTS_MD
from electro.html_snippets import build_snippet_notice_start, SNIPPET_NOTICE_END
from electro.simplepack import simplepack
from electro.inline_images import make_html_images_inline
from electro.inline_fonts import make_html_fonts_inline
from electro.inline_icons import make_html_icons_inline

pprint = CONFIG['console_pprint']


def build_project(project_directory):
    # -----------------------
    # Load project config file
    # -----------------------
    path_project_directory = Path(project_directory)
    if not path_project_directory.is_dir():
        FAULTS.error(f'project_directory is not a directory: {project_directory}')
        return
    path_project = path_project_directory / Path(CONFIG['project_filename'])
    if not path_project.exists():
        FAULTS.error(f'Project file {path_project} not found.')
        return
    with open(path_project, 'r') as file:
        project_config = json.load(file)
    CONFIG['project_config'] = project_config
    logger.info(f'Project Config:\n{pformat(project_config)}')

    # -----------------------
    # Determine site dir
    # -----------------------
    path_site_directory = Path(project_directory) / Path(project_config['site_directory'])
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

    if project_config.get('pack'):
        pack_site(path_site_directory)

def pack_site(path_site_directory):
    print("Packing...")
    for path_file in path_site_directory.glob('*.html'):
        if path_file.name == "search.html":
            continue
        if ".packed." in path_file.name:
            # Already packed
            continue
        if ".inlined." in path_file.name:
            # Already packed
            continue
        path_file_stage1 = path_site_directory / Path(f"{path_file.stem}.packed.stage1.html")
        path_file_stage2 = path_site_directory / Path(f"{path_file.stem}.packed.stage2.html")
        path_file_stage3 = path_site_directory / Path(f"{path_file.stem}.packed.stage3.html")
        path_file_packed = path_site_directory / Path(f"{path_file.stem}.packed.html")
        print(f'packing {path_file.name} to {path_file_packed}...')
        simplepack(path_file, path_file_stage1, uglify=False)
        make_html_images_inline(str(path_file_stage1), str(path_file_stage2))
        make_html_fonts_inline(path_file_stage2, path_file_stage3)
        make_html_icons_inline(path_file_stage3, path_file_packed)
        

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
            classes = "menu-node"
            if subheading_menu_html:
                subheading_menu_html = '\n' + subheading_menu_html
                caret_str = '<i class="caret fa fa-angle-right"></i>'
            else:
                classes += ' no_child'
                caret_str = ''
            self.menu_html += (
                f'<li><span class="{classes}" id="menuitem_doc_{document_name}">'
                f'{caret_str}'
                # f'<a href="{document_name}.html">{menu_name}</a>'
                f'{menu_name}'
                f'</span>{subheading_menu_html}</li>\n'
            )

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
            heading_url = f'{document_name}.html#{heading_id}'
            menu_html += (
                '        <li><span class="no_child">'
                + f'<a href="{heading_url}">{heading_text}</a>'
                + '</span></li>\n'
            )
        if menu_html:
            menu_html += '    </ul>\n'
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
        extensions=[
                'tables',
                'fenced_code',
                'electro.mdx_urlize:UrlizeExtension',
                'codehilite',
                'attr_list',
            ]
        if not CONFIG['disable_nl2br']:
            extensions.append('nl2br')
        document_html = markdown.markdown(
            document_markdown,
            extensions=extensions,
        )

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
            document_html = document_html.replace(img_tag, f'<div class="img-wrapper">{img_tag}</div>')

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
            markdown = markdown.replace(r'{{% notice ' + notice_start_type + r' %}}', html_temporary)

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

    def render_site(self):
        project_config = CONFIG['project_config']
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

        # TODO: cleanup
        path_site_document = path_site_directory / Path(f'index.html')
        # path_site_document = path_site_directory / Path(f'{document_name}.html')
        document_html = template_html.replace(r'{{% site_name %}}', project_config['site_name'])
        document_html = document_html.replace(r'{{% sidebar_menu %}}', self.menu_html)
        # TODO: cleanup
        # document_html = document_html.replace(r'{{% current_document_name %}}', document_name)
        document_html = document_html.replace(r'{{% current_document_name %}}', "Document")
        document_html = document_html.replace(r'{{% electro_version %}}', CONFIG['version'])

        pages_html = ""
        style_html = ""
        for document_name, document_info in self.site_documents.items():
            pages_html += f'<div class="content-page" id="{document_name}" {style_html}>'
            pages_html += f'(content from {document_name})<br><br>'
            pages_html += document_info['html']
            pages_html += '</div>'
            # Start all subsequent pages as hidden
            style_html = 'style="display: none"'

        document_html = document_html.replace(r'{{% content %}}', pages_html)
        print(f'   Building {path_site_document}...')
        with open(path_site_document, 'w') as file:
            file.write(document_html)

        # -------------------
        # Save search index
        # -------------------
        path_search_directory = path_site_directory / Path('search')
        path_search_directory.mkdir(parents=True, exist_ok=True)
        path_search_index = path_search_directory / Path('search_index.js')
        search_js = "App.searchData = " + json.dumps(self.search_index, indent=4)
        with open(path_search_index, 'w') as file:
            # TODO: Remove this indent after everything is working.
            file.write(search_js)
            # json.dump(self.search_index, file, indent=4)


def md_document_name_to_document_name(md_document_name):
    return Path(md_document_name).stem

def heading_text_to_id(text):
    _id = ''
    dash_appended = False
    text = text.replace('&nbsp;', '')
    for char in text.lower():
        if char == ' ':
            if not dash_appended:
                _id += '-'
                dash_appended = True
        elif re.match(r'[a-z0-9]', char):
            _id += char
            dash_appended = False
    return _id


def copy_directory_contents(source_directory, target_directory):
    logger.debug(f'Copying directory "{source_directory}" to "{target_directory}"')
    target_directory.mkdir(parents=True, exist_ok=True)
    paths_source_files = source_directory.glob('*')
    for path_source_file in sorted(list(paths_source_files)):
        logger.debug(f'   {path_source_file.name}')
        path_destination_file = target_directory / Path(path_source_file.name)
        shutil.copy(path_source_file, path_destination_file)
