#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""This module provides a class that abstracts the configuration and the SWF 'start_workflow_execution' operation.
"""
from __future__ import (print_function, unicode_literals)

import uuid
from datetime import datetime

from botocore.exceptions import ClientError
from contracts import contract
from six import text_type

from taran import Taran


class Starter(Taran):
    """Class that defines instances of starter that are used to perform checks and then execute a workflow."""

    @contract(configuration='*')
    def __init__(self, configuration=None):
        super(Starter, self).__init__(configuration=configuration)
        self.workflow_input = '-'
        self.foreman_task_list = configuration.FOREMAN_TASK_LIST if hasattr(configuration,
                                                                            'FOREMAN_TASK_LIST') else 'default'
        self.allow_parallel_exec = configuration.ALLOW_PARALLEL_EXEC if hasattr(configuration,
                                                                                'ALLOW_PARALLEL_EXEC') else False

    def ensure_domain_exists(self, domain_name=None):
        """Return true if specified domain exists, otherwise create it."""
        prefix = 'Registering domain \'{0}\':'.format(domain_name)
        try:
            self.swf_client.register_domain(
                name=domain_name,
                description='{0} domain'.format(domain_name),
                workflowExecutionRetentionPeriodInDays='10'
            )
            self.msg(message='{0} [REGISTERED]'.format(prefix))
            return True
        except ClientError as ce:
            if 'DomainAlreadyExistsFault' in ce.response['Error']['Code']:
                self.msg(message='{0} [OK]'.format(prefix))
                return True
            else:
                raise
        except:
            raise

    def start_workflow(self):
        """Start the workflow."""
        self.msg(message='Request received to start \'{0}\' workflow'.format(self.workflow_name))
        open_workflow_executions = self.swf_client.list_open_workflow_executions(
            domain=self.domain_name,
            startTimeFilter={
                'oldestDate': datetime(2016, 1, 1)
            },
            typeFilter={
                'name': self.workflow_name,
                'version': self.workflow_version
            },

        )
        if not self.allow_parallel_exec and open_workflow_executions.get('executionInfos'):
            self.msg(message='Cannot start workflow as an existing instance is already running', level='warning')
            exit(1)
        self.msg(message='Checking registration of workflow and activity types:')
        self.ensure_domain_exists(domain_name=self.domain_name)
        self.ensure_workflow_type_exists(
            workflow_name=self.workflow_name, workflow_version=self.workflow_version)
        for activity in self.activity_list:
            self.ensure_activity_type_exists(
                activity_name=activity.get('name'),
                activity_version=activity.get('version'),
                activity_task_list=activity.get('task_list'))

        self.workflow_id = text_type(uuid.uuid1())[:13]
        self.msg(message='Starting workflow execution')

        start_result = self.swf_client.start_workflow_execution(domain=self.domain_name,
                                                                workflowType={'name': self.workflow_name,
                                                                              'version': self.workflow_version},
                                                                executionStartToCloseTimeout='3600',
                                                                input=self.workflow_input,
                                                                taskStartToCloseTimeout='10',
                                                                workflowId=self.workflow_id,
                                                                taskList={
                                                                    'name': self.foreman_task_list})
        self.run_id = start_result['runId']
        self.msg(message='Started workflow execution')
        return {'run_id': self.run_id, 'workflow_id': self.workflow_id}

    # TODO: Determine how default task list should be set - defined in config?
    @contract(workflow_name='unicode', workflow_version='unicode')
    def ensure_workflow_type_exists(self,
                                    workflow_name=None, workflow_version=None):
        """Check the workflow type exists and create it if it doesn't."""
        prefix = 'Registering workflow type \'{0}\':'.format(workflow_name)
        try:
            self.swf_client.register_workflow_type(
                domain=self.domain_name,
                name=workflow_name,
                version=workflow_version,
                defaultTaskList={
                    'name': 'default'
                },
                defaultTaskStartToCloseTimeout='500',
                defaultExecutionStartToCloseTimeout='500',
                defaultChildPolicy='TERMINATE',
            )
            self.msg(message='{0} [REGISTERED]'.format(prefix))
            return True
        except ClientError as ce:
            if 'TypeAlreadyExistsFault' in ce.response['Error']['Code']:
                self.msg(message='{0} [OK]'.format(prefix))
                return True
            else:
                raise

    @contract(activity_name='unicode', activity_version='unicode', activity_task_list='unicode')
    def ensure_activity_type_exists(self,
                                    activity_name=None, activity_version=None, activity_task_list=None):
        """Check the activity type exists and create it if it doesn't."""
        prefix = 'Registering activity type \'{0}\':'.format(activity_name)
        try:
            self.swf_client.register_activity_type(
                domain=self.domain_name,
                name=activity_name,
                version=activity_version,
                defaultTaskStartToCloseTimeout='60',
                defaultTaskHeartbeatTimeout='600',
                defaultTaskList={
                    'name': activity_task_list
                },
                defaultTaskScheduleToStartTimeout='60',
                defaultTaskScheduleToCloseTimeout='60'
            )
            self.msg(message='{0} [REGISTERED]'.format(prefix))
            return True
        except ClientError as ce:
            if 'TypeAlreadyExistsFault' in ce.response['Error']['Code']:
                self.msg(message='{0} [OK]'.format(prefix))
                return True
            else:
                raise
