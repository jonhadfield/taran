# coding: utf-8
"""Test utils"""
from __future__ import (absolute_import, print_function, unicode_literals)

import responses
from taran.utils.web import url_check


@responses.activate
def test_web_url_check():
    """Test web check"""
    responses.add(responses.GET, 'http://example.com/test',
                  body='{}', status=200,
                  content_type='text/html')
    assert url_check(url='http://example.com/test', timeout=1)
