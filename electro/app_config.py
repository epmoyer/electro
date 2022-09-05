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

CONFIG = {
    'version': '1.0.1',
    'app_name': 'electro',
    'enable_debug_logging': True,
    'console_print': CONSOLE.print,
    'console_pprint': CONSOLE_PPRINT.print,
    'project_filename': 'electro.json',

    # ----------------------
    # Set at runtime
    # ----------------------
    'project_config': None,
    'path_project_directory': None,
    'path_site_directory': None,
    'path_theme_directory': None,
    'enable_newline_to_break': None,
    'output_format': None,
}
