"""Electro App, entry point"""

# Standard library
from pathlib import Path

# library
import click
from loguru import logger

# Local
from electro.app_config import CONFIG
from electro.faults import Faults
from electro.build import build_project

# Rich console
print = CONFIG['console_print']

@click.group()
@click.option('-d', '--debug', 'enable_debug_logging', is_flag=True, help='Enable debug logging')
def cli(enable_debug_logging):
    CONFIG['enable_debug_logging'] = enable_debug_logging

    CONFIG['fault_handler'] = Faults()

    # --------------------
    # Initialize logging
    # --------------------
    logger.remove()
    logger.add(
        f'logs/{CONFIG["app_name"]}.log',
        rotation="1 MB",
        retention=3,
        level="DEBUG" if enable_debug_logging else "INFO",
        format="{time:YYYY-MM-DD HH:mm:ss.SSS} | {level:<8} | {message}",
    )
    logger.info('------------------------------ BEGIN ------------------------------')
    logger.info(f'{CONFIG["app_name"]}, version {CONFIG["version"]}')


@cli.command()
@click.argument('project_directory', default='./')
def build(project_directory):
    print('build()')
    FAULTS = CONFIG['fault_handler']
    path_project_directory = Path(project_directory)
    if not path_project_directory.is_dir():
        FAULTS.error(f'project_directory is not a directory: {project_directory}')
    else:
        build_project(path_project_directory)
    FAULTS.render()

if __name__ == '__main__':
    cli(prog_name=CONFIG['app_name'])
