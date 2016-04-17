#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""This module provides utilities to simplify working with Amazon EC2 services."""
from __future__ import (absolute_import, print_function, unicode_literals)

import time

from requests import get
from requests.exceptions import RequestException

from taran.helpers.aws.clients import get_elb_client


def get_ec2_instance_id():
    """Use EC2 meta data to retrieve the instance id."""
    try:
        response = get(url='http://169.254.169.254/latest/meta-data/instance-id', timeout=1)
        return response.content
    except RequestException:
        return None
    except:
        raise


def elb_deregister_instances(elb_name=None, instance_ids=None, region=None):
    """De-Register EC2 instances on ELB."""
    instances_input = [{'InstanceId': inst_id} for inst_id in instance_ids]
    elb_client = get_elb_client(region=region)
    elb_client.deregister_instances_from_load_balancer(
        LoadBalancerName=elb_name,
        Instances=instances_input
    )
    while True:
        elb_state = elb_client.describe_load_balancers(
            LoadBalancerNames=[
                elb_name,
            ],
        )
        instances = elb_state['LoadBalancerDescriptions'][0]['Instances']
        for instance in instances:
            if instance.get('InstanceId') in instance_ids:
                time.sleep(5)
                continue
        break
    return True


def elb_register_instances(elb_name=None, instance_ids=None, region=None):
    """Register EC2 instances on an ELB."""
    # TODO: Implement timeout
    instances_input = [{'InstanceId': inst_id} for inst_id in instance_ids]
    elb_client = get_elb_client(region=region)
    elb_client.register_instances_with_load_balancer(
        LoadBalancerName=elb_name,
        Instances=instances_input
    )
    while True:
        instances_health = elb_client.describe_instance_health(
            LoadBalancerName=elb_name,
            Instances=instances_input,
        )
        if instances_health and instances_health.get('InstanceStates'):
            for instance_state in instances_health.get('InstanceStates'):
                if not instance_state.get('State') == 'InService':
                    time.sleep(8)
                    break
            return True
