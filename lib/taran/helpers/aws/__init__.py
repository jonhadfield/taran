#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""This module provides common AWS functionality."""
from __future__ import (absolute_import, print_function, unicode_literals)

import json

from botocore.exceptions import ClientError, NoCredentialsError
from requests import get
from requests.exceptions import RequestException

from taran.errors import TaranAWSCredentialsError, TaranAWSPermissionsError, TaranError
from taran.helpers.aws.clients import get_iam_client


def get_account_id():
    """Return the AWS Account id number associated with the current Boto3 session.
    If not available, try and get it from the instance meta-data (assuming this is an EC2 instance.)"""
    # Attempt to retrieve account number via API credentials
    try:
        iam_client = get_iam_client()
        users = iam_client.list_users(MaxItems=1)
        if users.get('Users'):
            return users.get('Users')[0]['Arn'].split(':')[4]
    except ClientError as ce:
        if ce.response['Error']['Code'] == 'AccessDenied':
            raise TaranAWSPermissionsError(
                "Taran requires iam:ListUsers read access to check the "
                "account id matches the one in the configuration.")
        else:
            raise TaranError("Unexpected error: {0}".format(ce))
    except NoCredentialsError:
        raise TaranAWSCredentialsError('AWS credentials could not be found when retrieving AWS account id.')
    except:
        raise

    # Attempt to retrieve account number via local meta-data (from an EC2 instance)
    try:
        response = get('http://169.254.169.254/latest/meta-data/iam/info/', timeout=1)
        json_output = json.loads(response.content)
        arn = json_output.get('InstanceProfileArn')
        return arn.split(':')[4]
    except RequestException:
        exit('Unable to obtain AWS account id.')
    except:
        raise
