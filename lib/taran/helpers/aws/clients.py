#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""This module provides functions to return boto3 (Python AWS SDK) low-level clients."""
from __future__ import (absolute_import, print_function, unicode_literals)

from boto3.session import Session
from botocore.client import Config
from botocore.exceptions import NoRegionError, NoCredentialsError
from taran.errors import TaranAWSCredentialsError


def get_swf_client(region=None):
    """Get a simple workflow client.

    Args:
        region (unicode): the region to connect to.
    Returns:
        an swf client.
    """
    session_config = Config(connect_timeout=70, read_timeout=70)
    if region:
        session = Session(region_name=region)
    else:
        session = Session()
    try:
        return session.client('swf', config=session_config)
    except (ValueError, NoRegionError) as exc:
        if 'invalid endpoint' in exc.message.lower():
            raise TaranAWSCredentialsError('Invalid endpoint when creating SWF client. Missing/invalid AWS Region?')
    except NoCredentialsError:
        raise TaranAWSCredentialsError('Unable to find AWS credentials when creating SWF client.')


def get_iam_client(region=None):
    """Get an iam client.

    Args:
        region (unicode): the region to connect to.
    Returns:
        an iam client.
    """
    if region:
        session = Session(region_name=region)
    else:
        session = Session()
    try:
        return session.client('iam')
    except (ValueError, NoRegionError) as exc:
        if 'invalid endpoint' in exc.message.lower():
            raise TaranAWSCredentialsError('Invalid endpoint when creating IAM client. Missing/invalid AWS Region?')
    except NoCredentialsError:
        raise TaranAWSCredentialsError('Unable to find AWS credentials when creating IAM client.')


def get_ec2_client(region=None):
    """Get an ec2 client.

    Args:
        region (unicode): the region to connect to.
    Returns:
        an ec2 client.
    """
    if region:
        session = Session(region_name=region)
    else:
        session = Session()
    try:
        return session.client('ec2')
    except NoRegionError:
        print('AWS region could not be determined when creating ec2 client')
    except:
        raise


def get_elb_client(region=None):
    """Get an elb client.

    Args:
        region (unicode): the region to connect to.
    Returns:
        an elb client.
    """
    if region:
        session = Session(region_name=region)
    else:
        session = Session()
    try:
        return session.client('elb')
    except NoRegionError:
        print('AWS region could not be determined when creating elb client')
    except:
        raise


def get_asg_client(region=None):
    """Get an asg client.

    Args:
        region (unicode): the region to connect to.
    Returns:
        an asg client.
    """
    if region:
        session = Session(region_name=region)
    else:
        session = Session()
    try:
        return session.client('autoscaling')
    except NoRegionError:
        print('AWS region could not be determined when creating asg client')
    except:
        raise


def get_s3_client(region=None):
    """Get an s3 client.

    Args:
        region (unicode): the region to connect to.
    Returns:
        an s3 client.
    """
    if region:
        session = Session(region_name=region)
    else:
        session = Session()
    try:
        return session.client('s3')
    except (ValueError, NoRegionError):
        exit('No region specified')
    except:
        raise
