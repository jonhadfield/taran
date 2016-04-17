#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""This module provides utilities to simplify working with Amazon S3."""
from __future__ import (print_function, unicode_literals)

from taran.helpers.aws.clients import get_s3_client
from taran.utils.files import write_md5


def get_s3_md5(s3_client=None, bucket=None, s3_path=None):
    """Get the MD5 (S3 ETag) sum of an S3 object."""
    if s3_path.startswith('/'):
        s3_path = s3_path[1:]
    etag = s3_client.head_object(
        Bucket=bucket,
        Key=s3_path,
    ).get('ETag')
    return etag[1:-1]


def s3_download(items=None):
    """Download an object from S3."""
    s3_client = get_s3_client()
    for item in items:
        s3_path = item.get('s3_path')[1:] if item.get('s3_path').startswith("/") else item.get('s3_path')
        s3_client.download_file(item.get('bucket'), s3_path, item.get('local_path'))
        write_md5(file_path=item.get('local_path'))
    return True


def s3_upload(items=None):
    """Upload an object to S3."""
    s3_client = get_s3_client()
    for item in items:
        s3_path = item.get('s3_path')[1:] if item.get('s3_path').startswith("/") else item.get('s3_path')
        s3_client.upload_file(item.get('local_path'), item.get('bucket'), s3_path)
    return True
