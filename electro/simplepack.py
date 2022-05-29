#!/usr/bin/env python

# Standard Library
import logging
import sys
import re
import copy
import json
import subprocess
from pathlib import Path

# Library
import click

# Local
import quicklog

__version__ = '1.1.0'
qlog = None  # Will assign on start
OUT_FILENAME_DEFAULT = 'PROJECT.html'
OUT_FILENAME_AUTODETECT = '(autodetect)'

BUILD_PATH = Path('project_build/')
EVENT_DATA_PATH = BUILD_PATH / Path('app/js/event_data.js')
LOG_PATH = BUILD_PATH / Path('logs/')
TEMP_PATH = BUILD_PATH / Path('temp/')


@click.command()
@click.argument("in_filename", default='project_build/app/index.html')
@click.argument("out_filename", default=OUT_FILENAME_AUTODETECT)
@click.option('-d', '--debug', is_flag=True)
@click.option('-n', '--nopack', is_flag=True)
@click.version_option(version=__version__)
def cli(in_filename, out_filename, debug, nopack):
    """CLI entry point."""
    simplepack(in_filename, out_filename, debug, nopack)


def simplepack(in_filename, out_filename=OUT_FILENAME_AUTODETECT, debug=False, nopack=False):
    """Module entry point."""
    # ----------------------
    # Make local build folder if it does not exist
    # ----------------------
    BUILD_PATH.mkdir(parents=True, exist_ok=True)
    LOG_PATH.mkdir(parents=True, exist_ok=True)
    TEMP_PATH.mkdir(parents=True, exist_ok=True)

    # Setup logging
    global qlog
    logger_name = 'simplepack'
    try:
        qlog = quicklog.get_logger(logger_name)
    except ValueError:
        qlog = quicklog.Quicklog(
            log_filename=LOG_PATH / Path('simplepack.log'),
            logging_level=logging.DEBUG if debug else logging.INFO,
            maxBytes=5000000,
            backupCount=1,
            logger_name='simplepack',
        )
    qlog.begin(f'simplepack version {__version__}')
    qlog.info('Command line: ' + ' '.join(sys.argv))
    qlog.info(f'Args: {in_filename=} {out_filename=} {debug=} {nopack=}')

    with open(in_filename) as file:
        lines = file.readlines()
    qlog.debug(f'Read {len(lines)} lines.')

    replacement_map = (
        # fmt:off
        # Content type Regex search string                                                    Replacement tag  # noqa:E501
        # ------------ ---------------------------------------------------------------------  -------          # noqa:E501
        ('javascript', r'^\s*<script\s*src="(.*)">\s*</script>\s$',                           'script'),       # noqa:E501, E241
        ('CSS',        r'^\s*<link\s*rel="stylesheet"\s*type=text/css\s*href="(.*)"\s*>\s*$', 'style'),
        # noqa:E501, E241
        # fmt:on
    )
    new_lines = []
    for line in lines:
        for content_type, regex, tag_name in replacement_map:
            find_result = re.findall(regex, line)
            if find_result:
                find_text = find_result[0]
                qlog.debug(f'Found {content_type} import:\n   line:{line}   find_text: {find_text}')
                if 'http' in find_text:
                    qlog.debug(f'Ignoring (not local file)')
                    continue
                filename = 'project_build/app/' + find_text
                qlog.info(f'Merging {content_type}: {filename}')
                new_lines = merge_file(new_lines, filename, tag_name, nopack)
                continue

        new_lines.append(line)

    if out_filename == OUT_FILENAME_AUTODETECT:
        out_filename = extract_output_filename()
        if out_filename is None:
            out_filename = OUT_FILENAME_DEFAULT

    qlog.lprint(f'Writing: {out_filename}')
    with open(out_filename, "w") as file:
        file.writelines(new_lines)

    qlog.end()


def extract_output_filename():
    """Build output filename by extracting project title from event data in JS project."""
    out_filename = None
    with open(EVENT_DATA_PATH, 'r') as file:
        lines = file.readlines()
    for line in lines:
        if 'project_title' in line:
            project_title = line.split(':')[1].strip().strip('"').replace(' ', '_')
            out_filename = 'PROJECT_' + project_title + '.html'
            break
    return out_filename


def merge_file(lines, filename, tag_name, nopack):
    if (
        not nopack
        and tag_name == 'script'
        and 'search-query-parser' not in filename
        and 'event_data' not in filename
        and 'userPreferences' not in filename
    ):
        # Minimize js using node package UglifyJS
        assert filename.startswith('project_build/app/')
        temp_filemane = TEMP_PATH / Path(filename).name
        command = f'uglifyjs {filename} -o {temp_filemane} -c'
        qlog.info(f'Compressing {filename}...')
        qlog.debug(f'Running: {command}')
        subprocess.run(command, shell=True)
        filename = temp_filemane
    new_lines = copy.deepcopy(lines)
    with open(filename) as file:
        script_lines = file.readlines()
    new_lines.append(f'   <{tag_name}>\n')
    if '\n' not in script_lines[-1]:
        script_lines[-1] = script_lines[-1] + ('\n')
    for script_line in script_lines:
        new_lines.append('      ' + script_line)
    new_lines.append(f'   </{tag_name}>\n')
    return new_lines


if __name__ == '__main__':
    cli()
