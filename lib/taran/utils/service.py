#!/u#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Service related utilities"""
from __future__ import (print_function, unicode_literals)

from taran.utils import execute_command


def service(name=None, action=None, daemonize=True):
    """Handle calls to affect installed daemons."""
    # TODO: Handle different service types
    execute_command(cmd='service {0} {1}'.format(name, action), daemonize=daemonize)
    return True
