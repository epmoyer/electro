import shutil
from pathlib import Path
import pytest

# Libraries
from result import Err

# Local
from loguru import logger
from test.log_manager import log_initialize

# Start default logger (imports below will use it)
log_initialize()
logger.info('test_build.py')

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
        (
            'singlefile',
        ),
        (
            'singlefile_legacy',
        ),
    ],
)
# fmt:on
def test_build(test_case_name):
    logger.separator('TEST')
    logger.info(f'{test_case_name=}')