"""Root file path definition."""

# Standard library
from pathlib import Path

PATH_PROJECT_ROOT = (Path(__file__).parent / Path('../')).resolve()
PATH_APP = PATH_PROJECT_ROOT / Path('app/')
