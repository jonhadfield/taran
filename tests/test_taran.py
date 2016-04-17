# coding: utf-8
"""Test workflow creation, start and termination basics"""
from __future__ import (absolute_import, print_function, unicode_literals)

import pytest
from moto import mock_swf

import tests.config as config
from taran.foreman import Foreman
from taran.helpers.aws.swf import get_activity_version
from taran.starter import Starter


# def test_import_error():
#     with mock.patch.dict('sys.modules', {'botocore.exceptions.ClientError': None}):
#         from taran.starter import Starter
#         starter = Starter()


@mock_swf
def test_starter_instanciation():
    """Test instanciation of Starter with example config"""
    starter = Starter(configuration=config)
    assert starter


@mock_swf
def test_foreman_instanciation():
    """Test foreman instanciation"""
    foreman = Foreman(configuration=config)
    assert foreman


@mock_swf
def test_starter_domain_registration():
    """Test domain registration when domain does not exist"""
    starter = Starter(configuration=config)
    assert starter.ensure_domain_exists(domain_name=starter.domain_name)
    # Test re-registration
    assert starter.ensure_domain_exists(domain_name=starter.domain_name)


@mock_swf
def test_activity_type_registration():
    """Test activity type registration"""
    starter = Starter(configuration=config)
    starter.ensure_domain_exists(domain_name=starter.domain_name)
    first_activity = starter.activity_list[0]
    assert starter.ensure_activity_type_exists(activity_name=first_activity['name'],
                                               activity_version=first_activity['version'],
                                               activity_task_list=first_activity['task_list'])
    # Test re-registration
    assert starter.ensure_activity_type_exists(activity_name=first_activity['name'],
                                               activity_version=first_activity['version'],
                                               activity_task_list=first_activity['task_list'])
    assert get_activity_version(activity_type=first_activity['name'],
                                activity_list=starter.activity_list) == first_activity['version']
    with pytest.raises(AttributeError):
        activity_list = [1, 2, 3]
        get_activity_version(activity_type='test', activity_list=activity_list)


@mock_swf
def test_starter_workflow_registration():
    """Test workflow registration"""
    starter = Starter(configuration=config)
    starter.domain_name = starter.domain_name
    starter.workflow_name = starter.workflow_name
    starter.workflow_version = starter.workflow_version
    starter.ensure_domain_exists(domain_name=starter.domain_name)
    assert starter.ensure_workflow_type_exists(workflow_name=starter.workflow_name,
                                               workflow_version=starter.workflow_name)
    # Test re-registration
    assert starter.ensure_workflow_type_exists(workflow_name=starter.workflow_name,
                                               workflow_version=starter.workflow_name)


@mock_swf
def test_workflow_starter():
    """Test termination of a workflow"""
    starter = Starter(configuration=config)
    starter.domain_name = starter.domain_name
    starter.default_task_list = starter.task_list
    starter.workflow_name = starter.workflow_name
    starter.workflow_version = starter.workflow_version
    starter.ensure_domain_exists(domain_name=starter.domain_name)
    starter.task_list = starter.task_list
    starter.activity_list = starter.activity_list
    assert starter.start_workflow()


@mock_swf
def test_workflow_termination():
    """Test termination of a workflow"""
    starter = Starter(configuration=config)
    starter.domain_name = starter.domain_name
    starter.default_task_list = starter.task_list
    starter.workflow_name = starter.workflow_name
    starter.workflow_version = starter.workflow_version
    starter.ensure_domain_exists(domain_name=starter.domain_name)
    starter.task_list = starter.task_list
    starter.activity_list = starter.activity_list
    starter.start_workflow()
    assert starter.terminate_workflow(reason='Test Reason', details='Test Details')


@mock_swf
def test_schedule_activity_task():
    """Test scheduling of an activity task"""
    starter = Starter(configuration=config)
    print(starter.foreman_task_list)
    starter.default_task_list = starter.task_list
    starter.workflow_name = starter.workflow_name
    starter.workflow_version = starter.workflow_version
    starter.ensure_domain_exists(domain_name=starter.domain_name)
    starter.task_list = starter.task_list
    starter.activity_list = starter.activity_list
    starter.start_workflow()
    foreman = Foreman(configuration=config)
    foreman.identity = foreman.hostname
    assert foreman.poll_for_decision_task()

# RAW BOTO3 EXAMPLE - FOR TEST COMPARISON
# @mock_swf
# def test_start_workflow_raw():
#     session = Session()
#     swf_client = session.create_client('swf')
#     swf_client.register_domain(
#         name='test',
#         description='test_desc',
#         workflowExecutionRetentionPeriodInDays='10'
#     )
#     swf_client.register_workflow_type(
#         domain='test',
#         name='test_wf_type',
#         version='1',
#         description='desc',
#         defaultTaskStartToCloseTimeout='5',
#         defaultExecutionStartToCloseTimeout='5',
#         defaultTaskList={
#             'name': 'task_list'
#         },
#         defaultTaskPriority='1',
#         defaultChildPolicy='TERMINATE')
#     swf_client.start_workflow_execution(domain='test', workflowId='123',
#                                         workflowType={'name': 'test_wf_type', 'version': '1'})


# @mock_swf
# def test_get_workflow_history():
#     starter = Starter()
#     starter.ensure_domain_exists(domain_name=config.DOMAIN_NAME)
#     starter.domain_name = config.DOMAIN_NAME
#     starter.default_task_list = config.DECIDER_TASK_LIST
#     starter.workflow_name = config.WORKFLOW_NAME
#     starter.workflow_version = config.WORKFLOW_VERSION
#     starter.domain_name = config.DOMAIN_NAME
#     starter.ensure_domain_exists(domain_name=config.DOMAIN_NAME)
#     starter.task_list = config.DECIDER_TASK_LIST
#     starter.activity_list = [{'task_list': config.DECIDER_TASK_LIST, 'name': 'name', 'version': '1'}]
#     start_response = starter.start_workflow()
#     foreman = Foreman()
#     foreman.domain_name = config.DOMAIN_NAME
#     foreman.workflow_id = start_response.get('workflow_id')
#     foreman.run_id = start_response.get('run_id')
#     assert foreman.get_workflow_history()
