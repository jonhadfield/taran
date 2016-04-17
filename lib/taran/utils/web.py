#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Utilities relating to online services"""
from __future__ import (print_function, unicode_literals)

import requests


def url_check(url=None, method='get', timeout=None, expected_codes=200):
    """Check if a request for a URL gives an expected response"""
    if expected_codes:
        if not isinstance(expected_codes, list):
            expected_codes = expected_codes,
        response = None
        if method:
            if method.lower() == 'get':
                response = requests.get(url=url, timeout=timeout)
            if method.lower() == 'head':
                response = requests.head(url=url, timeout=timeout)
        if response.status_code in expected_codes:
            return True
