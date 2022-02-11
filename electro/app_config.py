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

CONFIG = {
    'version': '0.1.0',
    'app_name': 'electro',
    'enable_debug_logging': True,
    'project_config': None,  # Loaded at runtime
    # 'fault_handler':  Faults(),  # Loaded at runtime
    'console_print': CONSOLE.print,
    'console_pprint': CONSOLE_PPRINT.print,
    'project_filename': 'electro.json'
}
