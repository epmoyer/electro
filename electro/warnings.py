"""Fault connection, reporting."""

# Library
from loguru import logger

# Local
from electro.console import CONSOLE, wrap_tag

# Rich console
print = CONSOLE.print


class Warnings:
    def __init__(self):
        self.warnings = []

    def warning(self, message):
        logger.warning(message)
        print(f'Warning: {wrap_tag("warning", message)}')
        self.warnings.append(message)

    def render(self):
        if not self.warnings:
            return
        print('\nWarnings:')
        for message in self.warnings:
            print(f'- {wrap_tag("warning", message)}')


WARNINGS = Warnings()
