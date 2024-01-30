# Standard Library
from pathlib import Path

# Library
from loguru import logger

# Local
from test.paths import PATH_LOG_FILE

LOG_STATUS = {
    'initialized': False,
    'enable_debug_logging': True,
}

def log_initialize():
    if LOG_STATUS['initialized']:
        return
    logger.remove()
    logger.add(
        PATH_LOG_FILE,
        rotation="1 MB",
        retention=6,
        level="DEBUG" if LOG_STATUS['enable_debug_logging'] else "INFO",
        format="{time:YYYY-MM-DD}T{time:HH:mm:ss.SSSZZ} | {level:<8} | {message}",
    )

    # Inject logger methods
    logger.separator = make_logger_separator(logger)
    logger.lprint = make_lprint(logger)

    logger.separator('BEGIN', marker="=")

    LOG_STATUS['initialized'] = True

def make_logger_separator(logger):
    """Build a function that logs a banner."""

    def separator(text, marker="-"):
        logger.info(marker * 28 + f' {text} ' + marker * 28)

    return separator

def make_lprint(logger):
    """Build a function that prints text and then logs it as info."""

    def lprint(text):
        logger.info(text)
        print(f'{text}')

    return lprint