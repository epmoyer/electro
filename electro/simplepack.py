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
from loguru import logger

# Local


__version__ = '1.2.0'
qlog = None  # Will assign on start

BUILD_PATH = Path('project_build/')
EVENT_DATA_PATH = BUILD_PATH / Path('app/js/event_data.js')
TEMP_PATH = BUILD_PATH / Path('temp/')


REPLACEMENT_MAP = (
    # fmt:off
    # Content type Regex search string                                                      Replacement tag  # noqa:E501
    # ------------ ---------------------------------------------------------------------    -------          # noqa:E501
    ('javascript', r'^\s*<script\s*type="text/javascript"\s*src="(.*)">\s*</script>\s$',    'script'),       # noqa:E501, E241
    ('CSS',        r'^\s*<link\s*rel="stylesheet"\s*type="text/css"\s*href="(.*)"\s*>\s*$', 'style'),
    # noqa:E501, E241
    # fmt:on
)

def simplepack(path_file_in, path_file_out, debug=False, uglify=True):
    """Module entry point."""

    logger.info(f'simplepack version {__version__}')
    logger.info(f'Args: {path_file_in=} {path_file_out=} {debug=} {uglify=}')

    with open(path_file_in) as file:
        lines = file.readlines()
    logger.debug(f'Read {len(lines)} lines.')

    new_lines = []
    for line in lines:
        expansion_lines = expand_line(line, path_file_in.parent, uglify)
        if expansion_lines:
            new_lines += expansion_lines
        else:
            new_lines.append(line)

    print(f'Writing: {path_file_out}')
    logger.info(f'Writing: {path_file_out}')

    with open(path_file_out, "w") as file:
        file.writelines(new_lines)

def expand_line(line, path_working_dir, uglify):
    for content_type, regex, tag_name in REPLACEMENT_MAP:
        find_result = re.findall(regex, line)
        if find_result:
            find_text = find_result[0]
            logger.debug(f'Found {content_type} import:\n   line:{line}   find_text: {find_text}')
            if 'http' in find_text:
                logger.debug('Ignoring (not local file)')
                continue
            filename = path_working_dir / Path(find_text)
            logger.info(f'Merging {content_type}: {filename}')
            return get_file_lines(filename, tag_name, uglify)
    return None

def get_file_lines(filename, tag_name, uglify):
    if (
        uglify
        and tag_name == 'script'
    ):
        # ----------------------
        # Make local build folder if it does not exist
        # ----------------------
        BUILD_PATH.mkdir(parents=True, exist_ok=True)
        TEMP_PATH.mkdir(parents=True, exist_ok=True)

        # Minimize js using node package UglifyJS
        assert filename.startswith('project_build/app/')
        temp_filemane = TEMP_PATH / Path(filename).name
        command = f'uglifyjs {filename} -o {temp_filemane} -c'
        logger.info(f'Compressing {filename}...')
        logger.debug(f'Running: {command}')
        subprocess.run(command, shell=True)
        filename = temp_filemane
    new_lines = []
    with open(filename) as file:
        script_lines = file.readlines()
    new_lines.append(f'   <{tag_name}>\n')
    if '\n' not in script_lines[-1]:
        script_lines[-1] = script_lines[-1] + ('\n')
    for script_line in script_lines:
        new_lines.append('      ' + script_line)
    new_lines.append(f'   </{tag_name}>\n')
    return new_lines
