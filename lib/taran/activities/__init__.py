#!/usr/bin/env python
# -*- coding: utf-8 -*-
from __future__ import (absolute_import, print_function, unicode_literals)

from taran.worker import Worker


class BaseActivity(Worker):
    def __init__(self, configuration=None):
        super(BaseActivity, self).__init__(configuration=configuration)
