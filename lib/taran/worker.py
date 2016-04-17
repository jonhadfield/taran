#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""This module provides worker/actor specific methods to all child classes.
"""
from __future__ import (absolute_import, print_function, unicode_literals)

import json

from botocore.exceptions import ClientError
from contracts import contract

from taran import Taran


class Worker(Taran):
    """A template for all decision processors.

    Attributes:
        configuration (module): The configuration a worker needs in order to participate in the workflow.
    """

    @contract(configuration='*')
    def __init__(self, configuration=None):
        """Initialise state that applies to all workers.

        Args:
            configuration (module): The configuration a worker needs in order to participate in the workflow.
        """
        super(Worker, self).__init__(configuration=configuration)

    def poll_for_activity_task(self):
        """Poll for an activity task from SWF and return if a task token has been provided.

        Returns:
            task (dict): Details of the assigned task.
        """
        self.msg(message='Polling for task routed to: ({0})...'.format(self.task_list))
        try:
            task = self.swf_client.poll_for_activity_task(domain=self.domain_name,
                                                          taskList={'name': self.task_list},
                                                          identity=self.identity)
            if task and 'taskToken' in task:
                self.activity_task = task
                self.activity_type_name = task['activityType']['name']
                self.workflow_id = task['workflowExecution']['workflowId']
                self.run_id = task['workflowExecution']['runId']
                self.task_token = task['taskToken']
                # SET WORKFLOW NAME AND VERSION
                workflow_execution = self.swf_client.describe_workflow_execution(domain=self.domain_name,
                                                                                 execution={
                                                                                     'workflowId': self.workflow_id,
                                                                                     'runId': self.run_id
                                                                                 }
                                                                                 )
                self.workflow_name = workflow_execution['executionInfo']['workflowType']['name']
                self.workflow_version = workflow_execution['executionInfo']['workflowType']['version']
        except ClientError as ce:
            if 'AccessDeniedException' in ce.response['Error']['Code']:
                self.msg(message='Insufficient privileges to poll for task', level='error')
                exit()
        except:
            raise

    def get_activity_results(self, activity=None):
        """Get a list of all results (when activity completed)"""
        activity_history = self.get_activity_history(workflow_history=self.workflow_history,
                                                     activity_type=activity)
        results_list = list()
        if activity_history:
            for activity_event in activity_history:
                if activity_event.get('status') == 'completed':
                    results_list.append(json.loads(activity_event.get('result')))
        return results_list

    def complete_activity_task(self, result='Undefined'):
        """Signal activity task as complete."""
        try:
            self.swf_client.respond_activity_task_completed(taskToken=self.task_token, result=result)
        except ClientError as ce:
            if 'UnknownResourceFault' in ce.response['Error']['Code']:
                self.msg(message='Unable to complete activity task as Workflow'
                                 ' execution does not exist (already terminated?)')

    def activity_task_failed(self, reason=None, details=None):
        """Signal that activity task failed."""
        try:
            self.swf_client.respond_activity_task_failed(
                taskToken=self.task_token,
                reason=reason,
                details=details
            )
        except ClientError as ce:
            print(str(ce))
