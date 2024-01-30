"""Paths to files."""

# Standard library
from pathlib import Path

# Local
from electro.path_root import PATH_APP, PATH_PROJECT_ROOT

PATH_LOGS = PATH_PROJECT_ROOT / Path('logs')
PATH_THEMES = PATH_PROJECT_ROOT / Path('themes')
PATH_MIXINS = PATH_PROJECT_ROOT / Path('mixins')
PATH_JS = PATH_PROJECT_ROOT / Path('site_resources') / Path('js')
PATH_SEARCH_RESULTS_MD = PATH_PROJECT_ROOT / Path('site_resources') / Path('search.md')
