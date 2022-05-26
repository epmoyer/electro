"""Electro App, entry point"""

# Standard library

# library
import click
from loguru import logger

# Local
from electro.app_config import CONFIG
from electro.build import build_project
from electro.faults import FAULTS

# Rich console
print = CONFIG['console_print']

@click.group()
@click.option('-d', '--debug', 'enable_debug_logging', is_flag=True, help='Enable debug logging')
@click.option('-n', '--nobreak', 'disable_nl2br', is_flag=True, help='Disable newline-to-break injection')
def cli(enable_debug_logging, disable_nl2br):
    CONFIG['enable_debug_logging'] = enable_debug_logging
    CONFIG['disable_nl2br'] = disable_nl2br

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
    build_project(project_directory)
    FAULTS.render()

if __name__ == '__main__':
    cli(prog_name=CONFIG['app_name'])
