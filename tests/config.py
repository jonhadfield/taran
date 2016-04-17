# coding: utf-8
"""Test configuration"""
from __future__ import (absolute_import, print_function, unicode_literals)

import os

os.environ["AWS_DEFAULT_REGION"] = "eu-west-1"
FOREMAN_TASK_LIST = 'my_task_list'
ACTIVITY_NAME = 'activity_name'
DOMAIN_NAME = 'test_domain'
WORKFLOW_NAME = 'test_workflow'
WORKFLOW_VERSION = '1'
ACTIVITY_VERSION = '1'
ACTIVITY_LIST = [{'task_list': 'none', 'name': ACTIVITY_NAME, 'version': ACTIVITY_VERSION}]
