"""Fault connection, reporting."""

# Standard Library
from enum import Enum

# Library
from loguru import logger

# Local
from electro.console import CONSOLE, wrap_tag

# Rich console
print = CONSOLE.print

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
    
    def has_warnings(self):
        return any(fault.fault_type == FaultType.WARNING for fault in self.faults)

    def error(self, message, cluster=None):
        self.faults.append(Fault(FaultType.ERROR, message, cluster))

    def warning(self, message, cluster=None):
        logger.warning(message)
        print(f'Warning: {wrap_tag("warning", message)}')
        self.faults.append(Fault(FaultType.WARNING, message, cluster))

    def render(self):
        if not self.faults:
            return
        if self.has_errors():
            print('\nErrors:')
            self._render_faults(FaultType.ERROR)
        if self.has_warnings():
            print('\nWarnings:')
            self._render_faults(FaultType.WARNING)
    
    def _render_faults(self, fault_type):
        for fault in self.faults:
            if fault.fault_type != fault_type:
                continue
            tag = fault.fault_type.name.lower()
            print(f'- {wrap_tag(tag, fault.message)}')
            if fault.cluster:
                print(f'Line: {fault.cluster.line_number + 1}')
                print('\n'.join(fault.cluster.lines))

FAULTS = Faults()