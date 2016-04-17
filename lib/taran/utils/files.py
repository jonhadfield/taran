#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""File related utilities"""
from __future__ import (absolute_import, print_function, unicode_literals)

import hashlib
import io
from os import path
from six import text_type
from contracts import contract


@contract(file_path='unicode')
def write_md5(file_path=None):
    """Write a file containing a file's MD5 checksum"""
    with io.open('{0}'.format(file_path), "r", encoding='utf-8') as index_file:
        data = text_type(index_file.read())
        md5 = hashlib.md5(data.encode('utf-8')).hexdigest()
        print(md5.__class__.__name__)
    with io.open('{0}.md5'.format(file_path), "w", encoding='utf-8') as md5_file:
        md5_file.write(text_type(md5))


@contract(file_path='unicode')
def get_local_md5(file_path=None):
    """Get the MD5 value of a file's content."""
    if not path.exists('{0}.md5'.format(file_path)):
        return None
    with open('{0}.md5'.format(file_path, "r")) as existing_md5_file:
        return existing_md5_file.read()
