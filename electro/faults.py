"""Fault connection, reporting."""

# Standard Library
from enum import Enum

# Library
from rich.markup import escape

# Local
from electro.app_config import CONFIG

# Rich console
print = CONFIG['console_print']


def wrap_tag(tag, text, no_escape=False):
    """Wrap text with a rich tag"""
    return f'[{tag}]{text if no_escape else escape(text)}[/{tag}]'


class FaultType(Enum):
    ERROR = 1
    WARNING = 2


class Fault:
    def __init__(self, fault_type, message, cluster=None):
        self.fault_type = fault_type
        self.message = message
        self.cluster = cluster


class Faults:
    def __init__(self):
        self.faults = []

    def has_errors(self):
        return any(fault.fault_type == FaultType.ERROR for fault in self.faults)

    def error(self, message, cluster=None):
        self.faults.append(Fault(FaultType.ERROR, message, cluster))

    def warning(self, message, cluster=None):
        self.faults.append(Fault(FaultType.WARNING, message, cluster))

    def render(self):
        if not self.faults:
            return
        print('\nErrors:')
        for fault in self.faults:
            tag = fault.fault_type.name.lower()
            print(wrap_tag(tag, fault.message))
            if fault.cluster:
                print(f'Line: {fault.cluster.line_number + 1}')
                print('\n'.join(fault.cluster.lines))

