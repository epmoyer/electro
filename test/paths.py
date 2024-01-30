# Standard Library
from pathlib import Path

PATH_TEST = Path(__file__).parent.resolve()

PATH_TEST_DATA = PATH_TEST / 'data'
PATH_TEST_DATA_RAW = PATH_TEST_DATA / 'raw'
PATH_TEST_DATA_PROCESSED = PATH_TEST_DATA / 'processed'

PATH_DATA_RAW_TEST_CASES = PATH_TEST_DATA_RAW / 'test_cases'
PATH_DATA_PROCESSED_TEST_CASES = PATH_TEST_DATA_PROCESSED / 'test_cases'

PATH_LOGS = PATH_TEST / 'logs'
PATH_LOG_FILE = PATH_LOGS / 'test.log'
