#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Host related utilities."""
from __future__ import (absolute_import, print_function, unicode_literals)

import socket


def get_hostname():
    """Return the instance's defined hostname"""
    return socket.gethostname()
