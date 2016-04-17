#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Utilities relating to online services"""
from __future__ import (print_function, unicode_literals)

import time

import requests


def url_check(url=None, method='get', timeout=None, expected_codes=200, interval=None, healthy_threshold=None):
    """Check if a request for a URL gives an expected response"""
    if expected_codes:
        if not isinstance(expected_codes, list):
            expected_codes = expected_codes,
        response = None
        wait_timeout = time.time() + timeout
        healthy_count = 0
        while wait_timeout > time.time():
            try:
                if method:
                    if method.lower() == 'get':
                        response = requests.get(url=url, timeout=10, verify=False)
                    if method.lower() == 'head':
                        response = requests.head(url=url, timeout=10, verify=False)
                if response.status_code in expected_codes:
                    healthy_count += 1
                if healthy_count == healthy_threshold:
                    return True
            except:
                healthy_count = 0
                time.sleep(5)
