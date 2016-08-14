#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""This module provides an abstract base class, containing common methods and
attributes that all types of workers and foreman instances may use."""

from __future__ import (absolute_import, print_function, unicode_literals)

import inspect
import json
import logging
import os
import time
from abc import ABCMeta, abstractmethod

from contracts import contract

from taran.helpers.aws import get_account_id
from taran.helpers.aws.clients import get_swf_client
from taran.helpers.aws.swf import get_activity_history
from taran.utils.host import get_hostname

__title__ = 'taran'
__version__ = '0.0.1'
__author__ = 'Jon Hadfield'
__license__ = 'MIT'
__copyright__ = 'Copyright 2016 Jon Hadfield'

try:
    from botocore.exceptions import ClientError
except ImportError:
    ClientError = None
    exit('boto3 is required but not installed.')

try:
    from collections import Counter
except ImportError:
    from backport_collections import Counter


class Taran(object):
    """The base class for all processors that interact with SWF (Simple WorkFlow).

    Attributes:
        processor (unicode): The type of processor that is processing decisions or tasks.
        hostname (unicode): The name of the host the processor is running on.
        configuration (module): The workflow configuration.
        domain_name: (unicode): The domain the workflow exists in.
        swf_client: (SWF): An instance of the SWF client.
        activity_task (unicode): Name of the activity task.
    """

    __metaclass__ = ABCMeta

    @contract(configuration='*')
    @abstractmethod
    def __init__(self, configuration=None):
        """Initialise state that applies to all processors.

        Args:
            configuration (module): The workflow configuration.
        """
        self.aws_region = configuration.AWS_REGION if hasattr(configuration, 'AWS_REGION') else None
        self.swf_client = get_swf_client(region=self.aws_region)
        self.configuration = configuration
        if hasattr(configuration, 'AWS_ACCOUNT_ID'):
            if configuration.AWS_ACCOUNT_ID != get_account_id():
                exit('Your configuration prevents this running outside of AWS account: {0}'.format(
                    configuration.AWS_ACCOUNT_ID))
        self.processor = None
        self.hostname = get_hostname()
        self.workflow_history = None
        self.task_list = None
        self.identity = None
        self.workflow_id = None
        self.run_id = None
        self.activity_type_name = None
        self.domain_name = configuration.DOMAIN_NAME if hasattr(configuration, 'DOMAIN_NAME') else None
        self.workflow_name = configuration.WORKFLOW_NAME if hasattr(configuration, 'WORKFLOW_NAME') else '-'
        self.workflow_version = configuration.WORKFLOW_VERSION if hasattr(configuration, 'WORKFLOW_VERSION') else '-'
        self.activity_task = None
        self.activity_list = configuration.ACTIVITY_LIST if hasattr(configuration, 'ACTIVITY_LIST') else None
        self.decision_task = None
        self.task_token = None
        self.log_level = self.get_log_level()
        self.logger = self.get_logger()

    def get_log_level(self):
        """Return a log level depending on specified LOG_LEVEL in configuration or default to INFO"""
        levels = {'NOTSET': 0, 'DEBUG': 10, 'INFO': 20, 'WARNING': 30, 'ERROR': 40, 'CRITICAL': 50}
        config_specified = self.configuration.LOG_LEVEL if hasattr(self.configuration, 'LOG_LEVEL') else None
        if isinstance(config_specified, int) and config_specified in levels.values():
            return config_specified
        elif config_specified and levels.get('config_specified'):
            return levels.get(config_specified)
        return 10

    @staticmethod
    @contract(workflow_history='dict[>0]', scheduled_ids='list|None', activity_type='unicode|None')
    def get_activity_history(workflow_history=None, scheduled_ids=None, activity_type=None):
        """Get a subset of the workflow history that relates to a specific activity type and (optionally) event ids.

        Args:
            workflow_history (Dict): The entire history of the workflow.
            scheduled_ids (Optional[List]): A list of scheduled event ids to reduce the event search by.
            activity_type (unicode): The type of activity to retrieve events for.

        Returns:
            task (dict): Workflow events matching the activity type and (oprtionally) event ids.
        """
        return get_activity_history(workflow_history=workflow_history, scheduled_ids=scheduled_ids,
                                    activity_type=activity_type)

    @contract(reason='unicode', domain_name='unicode|None', details='unicode',
              child_policy='unicode|None')
    def terminate_workflow(self, reason=None, domain_name=None,
                           details=None, child_policy='TERMINATE'):
        """End the workflow as the result of completion or failure.

        Args:
            reason (unicode): The type of activity to retrieve events for.
            domain_name (unicode): The name of domain the workflow to terminate is a member of
            details (unicode): Details of the termination.
            child_policy (Optional[unicode]): How to treat child workflows. Terminate by default.
        """
        try:
            if not domain_name:
                domain_name = self.domain_name
            self.swf_client.terminate_workflow_execution(
                domain=domain_name,
                workflowId=self.workflow_id,
                runId=self.run_id,
                reason=reason,
                details=details,
                childPolicy=child_policy
            )
            self.msg(message='Termination details: {0}'.format(details))
            return True
        except ClientError as ce:
            if 'Unknown execution' in ce.response['Error']['Code']:
                self.msg(message='Workflow already terminated')
        except Exception:
            raise

    def get_workflow_input(self):
        """Get the input specified when the workflow execution was started"""
        for event in self.workflow_history.get('events'):
            if event.get('eventType') == 'WorkflowExecutionStarted':
                return json.loads(event['workflowExecutionStartedEventAttributes']['input'])

    def get_activity_status(self, activity=None):
        """Retrieve the activity history and a count of each status recorded"""
        activity_history = self.get_activity_history(workflow_history=self.workflow_history,
                                                     activity_type=activity)

        activity_list = list()
        if activity_history:
            for activity_event in activity_history:
                if activity_event.get('status') in (
                        'failed', 'timed_out', 'cancelled', 'cancel_requested', 'scheduled', 'started'):
                    activity_list.append({'status': activity_event.get('status')})
                elif activity_event.get('status') == 'completed':
                    activity_list.append(
                        {'status': activity_event.get('status'), 'result': activity_event.get('result')})
        status_counts = Counter(token['status'] for token in activity_list)
        return dict(events=activity_list, counts=status_counts)

    @contract(message='unicode|None', level='unicode|None')
    def msg(self, message='-', level='info'):
        """Accept specific attributes to help produce an output message.

         Args:
             message (unicode): The message to output.
             level (unicode): The log level to output.
         """
        workflow_id = self.workflow_id if self.workflow_id else '-'
        run_id = self.run_id if self.run_id else '-'
        output = '\"{0}\" \"{1}\" \"{2}\" {3} {4} {5}'.format(self.domain_name, self.workflow_name,
                                                              self.workflow_version, workflow_id,
                                                              run_id, message)
        if level == 'debug':
            self.logger.debug(msg=output)
        elif level == 'info':
            self.logger.info(msg=output)
        elif level == 'warning':
            self.logger.warning(msg=output)
        elif level == 'error':
            self.logger.error(msg=output)
        elif level == 'critical':
            self.logger.critical(msg=output)

    def get_logger(self):
        """Return a logger we can use to write messages to files."""
        root_module_name = inspect.stack()[-1][1].split('/')[-1].split('.')[0]
        logger = logging.getLogger(root_module_name)
        if not len(logger.handlers):
            logger.setLevel(self.log_level)
            ch = logging.StreamHandler()
            ch.setLevel(self.log_level)
            log_dir = self.configuration.LOG_DIR if hasattr(self.configuration,
                                                            'LOG_DIR') else os.getcwd() + os.path.sep
            fh = logging.FileHandler('{0}{1}.log'.format(log_dir, root_module_name))
            fh.setLevel(logging.INFO)
            formatter = logging.Formatter(fmt='%(asctime)s %(levelname)s %(message)s')
            formatter.converter = time.gmtime
            ch.setFormatter(formatter)
            fh.setFormatter(formatter)
            # Suppress verbose boto messages
            logging.getLogger("botocore.vendored.requests.packages.urllib3.connectionpool").setLevel(logging.WARN)
            logger.addHandler(ch)
            logger.addHandler(fh)
            logger.propagate = False
        return logger
