#!/usr/bin/env python
# -*- coding: utf-8 -*-
""" The Foreman class - An abstraction of the AWS SWF Decider operations """
from __future__ import (absolute_import, print_function, unicode_literals)

import json
import uuid
from collections import namedtuple

from botocore.exceptions import ClientError
from contracts import contract

from taran import Taran
from taran.helpers.aws.swf import (get_activity_version)

Decision = namedtuple('Decision', ['name', 'type', 'schedule_to_start_timeout', 'start_to_close_timeout',
                                   'schedule_to_close_timeout', 'task_list', 'input'])


class Foreman(Taran):
    """A template for all decision processors.

    Attributes:
        configuration (module): The configuration a foreman needs in order to participate in the workflow.
    """

    @contract(configuration='*')
    def __init__(self, configuration=None):
        """Initialise state that applies to all foreman instances.

        Args:
            configuration (module): The configuration a foreman needs in order to participate in the workflow.
        """
        super(Foreman, self).__init__(configuration=configuration)
        self.task_list = configuration.FOREMAN_TASK_LIST if hasattr(configuration,
                                                                    'FOREMAN_TASK_LIST') else 'default'

    def poll_for_decision_task(self):
        """Poll for an decision task from SWF and return if a task token has been provided.

        Returns:
            task (dict): Details of the assigned task.
        """
        try:
            task = self.swf_client.poll_for_decision_task(domain=self.domain_name,
                                                          identity=self.identity,
                                                          taskList={'name': self.task_list})
            if task and 'taskToken' in task:
                self.decision_task = task
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
                self.workflow_history = dict(events=task['events'], next_page_token=task.get('nextPageToken'),
                                             previous_started_event_id=task.get('previousStartedEventId'))
            return True
        except ClientError as ce:
            if 'AccessDeniedException' in ce.response['Error']['Code']:
                self.msg(message='Insufficient privileges to poll for decision', level='error')
                exit()
        except:
            raise

    def get_workflow_history(self):
        """Get entire workflow history.

        Returns:
            a dict containing the entire workflow execution history
        """
        events = list()
        next_page_token = None
        while True:
            if not next_page_token:
                if not self.workflow_history:
                    self.workflow_history = self.swf_client.get_workflow_execution_history(
                        domain=self.domain_name,
                        execution={
                            'workflowId': self.workflow_id,
                            'runId': self.run_id
                        },
                        maximumPageSize=1000,
                        reverseOrder=True
                    )
                else:
                    events.extend(self.workflow_history.get('events'))
                if self.workflow_history.get('nextPageToken'):
                    next_page_token = self.workflow_history.get('nextPageToken')
                else:
                    self.workflow_history['events'] = events
                    break
            else:
                workflow_history = self.swf_client.get_workflow_execution_history(
                    domain=self.domain_name,
                    execution={
                        'workflowId': self.workflow_id,
                        'runId': self.run_id
                    },
                    nextPageToken=next_page_token,
                    maximumPageSize=1000,
                    reverseOrder=True
                )
                if workflow_history:
                    events.extend(workflow_history.get('events'))
                else:
                    exit('No response')
                if workflow_history.get('nextPageToken'):
                    next_page_token = workflow_history.get('nextPageToken')
                else:
                    self.workflow_history['events'] = events
                    break

    @contract(decisions='list')
    def schedule_activity_tasks(self, decisions=None):
        """Retrieve the workflow history.

        Args:
            decisions (List): A list of dictionaries containing details of the activities to schedule.
        """
        decisions_to_schedule = list()
        for decision_to_schedule in decisions:
            decisions_to_schedule.append(
                {'decisionType': 'ScheduleActivityTask',
                 'scheduleActivityTaskDecisionAttributes': {
                     'activityType': {
                         'name': decision_to_schedule.name,
                         'version': get_activity_version(activity_type=decision_to_schedule.type,
                                                         activity_list=self.activity_list),
                     },
                     'scheduleToStartTimeout': decision_to_schedule.schedule_to_start_timeout,
                     'startToCloseTimeout': decision_to_schedule.start_to_close_timeout,
                     'scheduleToCloseTimeout': decision_to_schedule.schedule_to_close_timeout,
                     'activityId': str(uuid.uuid1()),
                     'taskList': {'name': decision_to_schedule.task_list},
                     'input': decision_to_schedule.input
                 }
                 }
            )
        try:
            self.swf_client.respond_decision_task_completed(taskToken=self.task_token,
                                                            decisions=decisions_to_schedule)
            return True
        except ClientError:
            raise
        except:
            raise

    def get_activity_results(self, activity=None):
        """Get the result returned when the activity became completed."""
        activity_history = self.get_activity_history(workflow_history=self.workflow_history,
                                                     activity_type=activity)
        results_list = list()
        if activity_history:
            for activity_event in activity_history:
                if activity_event.get('status') == 'completed':
                    results_list.append(json.loads(activity_event.get('result')))
        return results_list
