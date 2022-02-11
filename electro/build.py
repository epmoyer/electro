# Standard library
from pathlib import Path
import json

# Library
from prettyprinter import pformat

# Local
from loguru import logger
from electro.app_config import CONFIG
from electro.faults import FAULTS
from electro.paths import PATH_THEMES

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
    path_template_directory = Path()

    # print(f'build_project() {path_project_directory=}')
    # pprint(CONFIG)

    # -----------------------
    # Build menu and pase markdown
    # -----------------------
    builder = Builder()
    for navigation_descriptor in project_config['navigation']:
        builder.add_navigation_descriptor(navigation_descriptor)
    builder.render_site()


class Builder:
    def __init__(self):
        self.menu_html = ''
        self.menu_section_html = ''
        self.site_documents = {}

    def add_navigation_descriptor(self, navigation_descriptor):
        if section_name := navigation_descriptor.get('section'):
            self.menu_section_html += f'<div class="section-heading">{section_name}</div>'
        self.menu_section_html += '<ul class="menu-tree">'
        documents_dict = navigation_descriptor.get('documents')
        if documents_dict is None:
            FAULTS.error(f'No "documents" key in navigation descriptor {navigation_descriptor}.')
            return
        for menu_name, md_document_name in documents_dict.items():
            self.menu_section_html += f'<li><span class="no_child">{menu_name}</span></li>'
            self.build_document(md_document_name)
        self.menu_section_html += '/<ul>'

    def build_document(self, md_document_name):
        path_markdown = CONFIG['path_project_directory'] / Path('docs') / Path(md_document_name)
        if not path_markdown.exists():
            FAULTS.error(f'Source markdown document {path_markdown} does not exist.')
            return
        document_name = path_markdown.stem
        self.site_documents[document_name] = {
            'path_markdown': path_markdown
        }

    def render_site(self):
        for document_name, document_info in self.site_documents.items():
            path_site_document = CONFIG['path_site_directory'] / Path(
                f'{document_name}.html'
            )

            print(path_site_document)
