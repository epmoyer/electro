"""Electro App, entry point"""

# Standard library
from pathlib import Path

# library
import click
from loguru import logger
from result import Result, Ok, Err

# Local
from electro.app_config import CONFIG
from electro.console import CONSOLE, wrap_tag
from electro.build import build_project
from electro.warnings import WARNINGS

# Rich console
print = CONSOLE.print

@click.group()
@click.option('-d', '--debug', 'enable_debug_logging', is_flag=True, help='Enable debug logging')
def cli(enable_debug_logging):
    CONFIG['enable_debug_logging'] = enable_debug_logging

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
@click.argument('path_project_text', metavar='project_path', default='./')
def build(path_project_text):
    """Build the project.

    The user can pass either the project directory OR a path to the project file (i.e. the 
    project configuration JSON file).  If a directory is passed, then we will assume the
    configuration JSON file is in that directory and has the default name.
    """
    path_project = Path(path_project_text)
   
    result = build_project(path_project)
    if isinstance(result, Err):
        logger.error(result.value)
        print(f'Error: {wrap_tag("error", result.value)}')
    WARNINGS.render()

if __name__ == '__main__':
    cli(prog_name=CONFIG['app_name'])
