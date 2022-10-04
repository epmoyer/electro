"""Rich text console."""

# library
from rich.console import Console
from rich.theme import Theme
from rich.markup import escape

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

def wrap_tag(tag, text, no_escape=False):
    """Wrap text with a rich tag"""
    return f'[{tag}]{text if no_escape else escape(text)}[/{tag}]'