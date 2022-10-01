"""Global app config."""

# library
from rich.console import Console
from rich.theme import Theme

# --------------------
# Rich output console
# --------------------
# fmt: off
THEME = Theme({
    "error": "#FF1020",
    "warning": "#FD971F",
})
# fmt: on

CONSOLE = Console(highlight=False, color_system='256', theme=THEME)
CONSOLE_PPRINT = Console(highlight=True, color_system='256', theme=THEME)
OUTPUT_FORMATS = ['static_site', 'single_file']
