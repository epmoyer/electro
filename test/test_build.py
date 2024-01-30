# Standard Library
import shutil
import pytest
import json

# Library
from loguru import logger
from result import Err

# Local
from test.log_manager import log_initialize
from electro.build import build_project

# Start default logger (imports below will use it)
log_initialize()
logger.info('test_build.py')

from test.util_files import make_workspace_dir, copy_items
from test.paths import (
    PATH_DATA_RAW_TEST_CASES,
    PATH_DATA_PROCESSED_TEST_CASES,
)

def test_clear_output_dir():
    path_results_data = PATH_DATA_PROCESSED_TEST_CASES
    if path_results_data.exists():
        shutil.rmtree(path_results_data)
    path_results_data.mkdir()


def log_test_separator():
    logger.info('-' * 30 + ' TEST ' + '-' * 30)

# fmt:off
@pytest.mark.parametrize(
    "test_case_name",
    [
        'single_file',
        'single_file_legacy',
        'static',
    ],
)
# fmt:on
def test_build(test_case_name):
    logger.separator('TEST')
    logger.info(f'{test_case_name=}')

    path_source_data = PATH_DATA_RAW_TEST_CASES / test_case_name
    path_workspace_dir = PATH_DATA_PROCESSED_TEST_CASES / test_case_name
    path_workspace_incoming_dir, path_workspace_build_dir, path_workspace_results_dir = make_workspace_dir(path_workspace_dir)

    copy_items(
        path_source_data,
        ['electro.json', 'docs/'],
        path_workspace_incoming_dir,
    )

    with open(path_source_data / 'electro.json', 'r') as file:
        electro_config = json.load(file)
    is_single_file = electro_config['output_format'] == 'single_file'

    result = build_project(path_workspace_incoming_dir)

    # sourcery skip: no-conditionals-in-tests
    if isinstance(result, Err):
        raise RuntimeError(f'Electro Build Error: {result.err_value}')

    # sourcery skip: no-conditionals-in-tests
    if is_single_file:
        # Copy the output file to the results dir, renamed with the test test case name.
        path_output_file = path_workspace_build_dir / 'index.html'
        path_result_file = path_workspace_results_dir / f'{test_case_name}.html'
        shutil.copy(path_output_file, path_result_file)
