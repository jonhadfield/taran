# coding: utf-8
"""Test workflow creation, start and termination basics"""
from __future__ import (absolute_import, print_function, unicode_literals)

from moto import mock_swf, mock_ec2, mock_elb, mock_autoscaling, mock_s3

from taran.helpers.aws.clients import (get_swf_client, get_ec2_client, get_elb_client,
                                       get_asg_client, get_s3_client)


@mock_swf
def test_get_swf_client():
    """Get a simple workflow client."""
    assert get_swf_client()
    assert get_swf_client(region='us-east-1')


@mock_ec2
def test_get_ec2_client():
    """Get an ec2 client."""
    assert get_ec2_client()
    assert get_ec2_client(region='us-east-1')


@mock_elb
def test_get_elb_client():
    """Get an elb client."""
    assert get_elb_client()
    assert get_elb_client(region='us-east-1')


@mock_autoscaling
def test_get_asg_client():
    """Get an asg client."""
    assert get_asg_client()
    assert get_asg_client(region='us-east-1')


@mock_s3
def test_get_s3_client():
    """Get an s3 client."""
    assert get_s3_client()
    assert get_s3_client(region='us-east-1')
    # with pytest.raises(ValueError, SystemExit):
    #     os.environ['AWS_DEFAULT_REGION'] = ''
    #     assert get_s3_client()
