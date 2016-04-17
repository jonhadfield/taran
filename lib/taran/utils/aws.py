#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""AWS Utils"""
from __future__ import (absolute_import, print_function, unicode_literals)

from contracts import contract


@contract(instance_id='unicode')
def is_ec2_instance_id(instance_id=None):
    """Return True if instance id is a valid format."""
    if instance_id.startswith('i-') and len(instance_id) in (10, 18):
        if instance_id.split('i-')[1].isalnum:
            return True
    return False
