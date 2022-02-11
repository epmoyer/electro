
# Standard library
from pathlib import Path
import json

# Local
from electro.app_config import CONFIG
from electro.faults import FAULTS


def build_project(project_directory):
    path_project_directory = Path(project_directory)
    if not path_project_directory.is_dir():
        FAULTS.error(f'project_directory is not a directory: {project_directory}')
        return
    path_project = path_project_directory / Path(CONFIG['project_filename'])
    if not path_project.exists():
        FAULTS.error(f'Project file {path_project} not found.')
        return

    print(f'build_project() {path_project_directory=}')
