# coding: utf-8
"""Test helpers"""
from __future__ import (absolute_import, print_function, unicode_literals)

import datetime

import pytest
from boto3 import Session
from dateutil.tz.tz import tzlocal
from moto import mock_ec2, mock_s3, mock_iam
from six import text_type

from taran.errors import TaranError
from taran.helpers.aws import get_account_id
from taran.helpers.aws.s3 import get_s3_md5, s3_download, s3_upload
from taran.helpers.aws.swf import get_activity_history, get_activity_version


# TODO - Implement once moto library supports list_users
@mock_iam
def test_get_account_id():
    """Test get_account_id from ec2 instance"""
    with pytest.raises(NotImplementedError):
        get_account_id()


@mock_ec2
def test_get_ec2_instance_id():
    """Return the id from an ec2 instance"""
    assert test_get_ec2_instance_id


@mock_s3
def test_get_s3_md5(tmpdir):
    """Return the md5/etag of an object in S3"""
    session = Session()
    s3_client = session.client('s3')
    s3_client.create_bucket(Bucket='test-bucket')
    local_file = tmpdir.mkdir("sub").join("test-file.txt")
    local_file.write("content")
    s3_client.upload_file(str(local_file), "test-bucket", "test-file.txt")
    assert get_s3_md5(s3_client=s3_client, bucket='test-bucket',
                      s3_path='test-file.txt') == '9a0364b9e99bb480dd25e1f0284c8555'
    assert get_s3_md5(s3_client=s3_client, bucket='test-bucket',
                      s3_path='/test-file.txt') == '9a0364b9e99bb480dd25e1f0284c8555'


@mock_s3
def test_s3_download(tmpdir):
    """Download a file from S3"""
    session = Session()
    s3_client = session.client('s3')
    s3_client.create_bucket(Bucket='test-bucket')
    local_file = tmpdir.mkdir("sub").join("test-file.txt")
    local_file.write("content")
    download_dir = tmpdir.mkdir("download_dir").join("downloaded.txt")
    s3_client.upload_file(str(local_file), "test-bucket", "test-file.txt")
    assert s3_download(
        items=[{'s3_path': 'test-file.txt', 'bucket': 'test-bucket', 'local_path': text_type(download_dir)}])


@mock_s3
def test_s3_upload(tmpdir):
    """Upload a file to S3"""
    session = Session()
    s3_client = session.client('s3')
    s3_client.create_bucket(Bucket='test-bucket')
    local_file = tmpdir.mkdir("sub").join("test-file.txt")
    local_file.write("content")
    s3_client.upload_file(str(local_file), "test-bucket", "test-file.txt")
    assert s3_upload(
        items=[
            {'s3_path': 'test-file.txt', 'bucket': 'test-bucket', 'local_path': text_type(local_file)}])


successful_workflow_history = {'previous_started_event_id': 9, 'next_page_token': None, 'events': [
    {'eventId': 1, 'eventType': 'WorkflowExecutionStarted',
     'workflowExecutionStartedEventAttributes': {'taskList': {'name': 'default'},
                                                 'parentInitiatedEventId': 0, 'taskStartToCloseTimeout': '10',
                                                 'childPolicy': 'TERMINATE',
                                                 'executionStartToCloseTimeout': '3600',
                                                 'input': '{"test": "test","task_list": "i-6fbd1de3"}',
                                                 'workflowType': {'version': '1', 'name': 'wftype'}},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 31, 949000, tzinfo=tzlocal())},
    {'eventId': 2, 'eventType': 'DecisionTaskScheduled',
     'decisionTaskScheduledEventAttributes': {'startToCloseTimeout': '10',
                                              'taskList': {'name': 'default'}},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 31, 949000, tzinfo=tzlocal())},
    {'eventId': 3, 'eventType': 'DecisionTaskStarted',
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 32, 9000, tzinfo=tzlocal()),
     'decisionTaskStartedEventAttributes': {'scheduledEventId': 2, 'identity': 'localhost'}},
    {'eventId': 4, 'eventType': 'DecisionTaskCompleted',
     'decisionTaskCompletedEventAttributes': {'startedEventId': 3, 'scheduledEventId': 2},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 32, 136000, tzinfo=tzlocal())},
    {'eventId': 5, 'eventType': 'ActivityTaskScheduled',
     'activityTaskScheduledEventAttributes': {'taskList': {'name': 'i-6fbd1de3'},
                                              'scheduleToCloseTimeout': '20',
                                              'activityType': {'version': '1', 'name': 'activity1'},
                                              'decisionTaskCompletedEventId': 4, 'heartbeatTimeout': '600',
                                              'activityId': '6167c0d7-04bb-11e6-8cab-3c15c2e45d3a',
                                              'scheduleToStartTimeout': '10', 'startToCloseTimeout': '5',
                                              'input': 'input'},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 32, 136000, tzinfo=tzlocal())},
    {'eventId': 6, 'eventType': 'ActivityTaskStarted',
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 32, 174000, tzinfo=tzlocal()),
     'activityTaskStartedEventAttributes': {'scheduledEventId': 5, 'identity': 'i-6fbd1de3'}},
    {'eventId': 7, 'eventType': 'ActivityTaskCompleted',
     'activityTaskCompletedEventAttributes': {'startedEventId': 6, 'scheduledEventId': 5,
                                              'result': '{"result": "the_result"}'},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 33, 517000, tzinfo=tzlocal())},
    {'eventId': 8, 'eventType': 'DecisionTaskScheduled',
     'decisionTaskScheduledEventAttributes': {'startToCloseTimeout': '10',
                                              'taskList': {'name': 'default'}},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 33, 517000, tzinfo=tzlocal())},
    {'eventId': 9, 'eventType': 'DecisionTaskStarted',
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 33, 552000, tzinfo=tzlocal()),
     'decisionTaskStartedEventAttributes': {'scheduledEventId': 8, 'identity': 'localhost'}},
    {'eventId': 10, 'eventType': 'DecisionTaskCompleted',
     'decisionTaskCompletedEventAttributes': {'startedEventId': 9, 'scheduledEventId': 8},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 33, 751000, tzinfo=tzlocal())},
    {'eventId': 11, 'eventType': 'ActivityTaskScheduled',
     'activityTaskScheduledEventAttributes': {'taskList': {'name': 'i-6fbd1de3'},
                                              'scheduleToCloseTimeout': '80', 'activityType': {'version': '1',
                                                                                               'name': 'activity2'},
                                              'decisionTaskCompletedEventId': 10, 'heartbeatTimeout': '600',
                                              'activityId': '625d765e-04bb-11e6-b6a9-3c15c2e45d3a',
                                              'scheduleToStartTimeout': '60', 'startToCloseTimeout': '20',
                                              'input': '{"test": "test"'},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 33, 751000, tzinfo=tzlocal())},
    {'eventId': 12, 'eventType': 'ActivityTaskStarted',
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 34, 323000, tzinfo=tzlocal()),
     'activityTaskStartedEventAttributes': {'scheduledEventId': 11, 'identity': 'i-6fbd1de3'}},
    {'eventId': 13, 'eventType': 'ActivityTaskCompleted',
     'activityTaskCompletedEventAttributes': {'startedEventId': 12, 'scheduledEventId': 11,
                                              'result': '{"instance_id": "i-6fbd1de3", "test": "test"}}'},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 36, 71000, tzinfo=tzlocal())},
    {'eventId': 14, 'eventType': 'DecisionTaskScheduled',
     'decisionTaskScheduledEventAttributes': {'startToCloseTimeout': '10',
                                              'taskList': {'name': 'default'}},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 36, 71000, tzinfo=tzlocal())},
    {'eventId': 15, 'eventType': 'DecisionTaskStarted',
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 42, 36, 104000, tzinfo=tzlocal()),
     'decisionTaskStartedEventAttributes': {'scheduledEventId': 14, 'identity': 'localhost'}}]}

timed_out_workflow_history = {'previous_started_event_id': 3, 'next_page_token': None, 'events': [
    {'eventId': 1, 'eventType': 'WorkflowExecutionStarted',
     'workflowExecutionStartedEventAttributes': {'taskList': {'name': 'task_list'},
                                                 'parentInitiatedEventId': 0, 'taskStartToCloseTimeout': '10',
                                                 'childPolicy': 'TERMINATE',
                                                 'executionStartToCloseTimeout': '3600',
                                                 'input': '{"test": "test", "task_list": "i-e4c7ba6c"}',
                                                 'workflowType': {'version': '1', 'name': 'activity1'}},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 34, 32, 234000, tzinfo=tzlocal())},
    {'eventId': 2, 'eventType': 'DecisionTaskScheduled',
     'decisionTaskScheduledEventAttributes': {'startToCloseTimeout': '10',
                                              'taskList': {'name': 'task_list'}},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 34, 32, 234000, tzinfo=tzlocal())},
    {'eventId': 3, 'eventType': 'DecisionTaskStarted',
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 34, 32, 298000, tzinfo=tzlocal()),
     'decisionTaskStartedEventAttributes': {'scheduledEventId': 2, 'identity': 'localhost'}},
    {'eventId': 4, 'eventType': 'DecisionTaskCompleted',
     'decisionTaskCompletedEventAttributes': {'startedEventId': 3, 'scheduledEventId': 2},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 34, 32, 420000, tzinfo=tzlocal())},
    {'eventId': 5, 'eventType': 'ActivityTaskScheduled',
     'activityTaskScheduledEventAttributes': {'taskList': {'name': 'i-e4c7ba6c'}, 'scheduleToCloseTimeout': '20',
                                              'activityType': {'version': '1', 'name': 'activity1'},
                                              'decisionTaskCompletedEventId': 4, 'heartbeatTimeout': '600',
                                              'activityId': '4378db30-04ba-11e6-859f-3c15c2e45d3a',
                                              'scheduleToStartTimeout': '10', 'startToCloseTimeout': '5',
                                              'input': ''},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 34, 32, 420000, tzinfo=tzlocal())},
    {'eventId': 6, 'eventType': 'ActivityTaskTimedOut',
     'activityTaskTimedOutEventAttributes': {'startedEventId': 0, 'timeoutType': 'SCHEDULE_TO_START',
                                             'scheduledEventId': 5},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 34, 42, 425000, tzinfo=tzlocal())},
    {'eventId': 7, 'eventType': 'DecisionTaskScheduled',
     'decisionTaskScheduledEventAttributes': {'startToCloseTimeout': '10',
                                              'taskList': {'name': 'task_list'}},
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 34, 42, 425000, tzinfo=tzlocal())},
    {'eventId': 8, 'eventType': 'DecisionTaskStarted',
     'eventTimestamp': datetime.datetime(2016, 4, 17, 17, 34, 42, 468000, tzinfo=tzlocal()),
     'decisionTaskStartedEventAttributes': {'scheduledEventId': 7, 'identity': 'localhost'}}]}


def test_get_activity_history():
    """Get history for a particular activity from a workflow history"""
    assert get_activity_history(workflow_history=successful_workflow_history, activity_type='activity1')


def test_get_activity_history_with_timed_out_activity():
    """Get history for a particular activity from a workflow history"""
    assert get_activity_history(workflow_history=timed_out_workflow_history, activity_type='activity1')


def test_get_activity_with_missing_input():
    """Test the retrieval of a version number from an activity list based on the name as input."""
    with pytest.raises(TaranError):
        assert get_activity_history(workflow_history=successful_workflow_history)


def test_get_activity_version():
    """Test the retrieval of a version number from an activity list based on the name as input."""
    activity_list = [{'name': 'test', 'version': '9'}]
    assert get_activity_version(activity_type='test', activity_list=activity_list)
