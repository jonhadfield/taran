#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Utilities relating to online services"""
from __future__ import (print_function, unicode_literals)

import time

import requests


def url_check(url=None, method='get', timeout=None, response_timeout=2, expected_codes=200, interval=2,
              healthy_threshold=1, ssl_verify=False):
    """Check if a request for a URL gives an expected response"""
    if not isinstance(expected_codes, list):
        expected_codes = expected_codes,
    response = None
    wait_timeout = time.time() + timeout
    healthy_count = 0
    while wait_timeout > time.time():
        try:
            if method.lower() == 'get':
                response = requests.get(url=url, timeout=response_timeout, verify=ssl_verify)
            elif method.lower() == 'head':
                response = requests.head(url=url, timeout=response_timeout, verify=ssl_verify)
            if response.status_code in expected_codes:
                healthy_count += 1
            if healthy_count == healthy_threshold:
                return True
        # TODO: Remove broad exception - some scenarios should be a failure
        except BaseException:
            healthy_count = 0
            time.sleep(interval)
