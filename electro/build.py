# Standard library
from pathlib import Path
import json

# Library
from prettyprinter import pformat

# Local
from loguru import logger
from electro.app_config import CONFIG
from electro.faults import FAULTS

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

    print(f'build_project() {path_project_directory=}')
    pprint(CONFIG)
