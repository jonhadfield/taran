# -*- coding: utf-8 -*-
from __future__ import (absolute_import, division, print_function)


class TaranError(Exception):
    """Base class for all Taran errors."""

    def __init__(self, message=None):
        self.message = message

    def __str__(self):
        return self.message

    def __repr__(self):
        return self.message


class TaranAWSCredentialsError(TaranError):
    """Class for all AWS credential related errors."""

    def __init__(self, message=None):
        self.message = message

    def __str__(self):
        return self.message

    def __repr__(self):
        return self.message


class TaranAWSPermissionsError(TaranError):
    """Class for all AWS credential related errors."""

    def __init__(self, message=None):
        self.message = message

    def __str__(self):
        return self.message

    def __repr__(self):
        return self.message
