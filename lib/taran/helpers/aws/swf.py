#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""This module provides utilities to simplify working with Amazon SWF."""
from __future__ import (absolute_import, print_function, unicode_literals)

from taran.errors import TaranError


def get_activity_version(activity_type=None,
                         activity_list=None):
    """Get the version of the activity.

    Args:
        activity_type (unicode): the type of the activity to retrieve the version of.
        activity_list (list): the list of activities
    Returns:
        the version number of the matching activity.
    """
    try:
        for activity in activity_list:
            if activity.get('name') == activity_type:
                return activity.get('version')
    except AttributeError:
        raise AttributeError('Activity is not a dict')


def get_activity_history(workflow_history=None, scheduled_ids=None, activity_type=None):
    """Get workflow history for a specific activity.

    Args:
        workflow_history (dict): the current workflow history.
        scheduled_ids (list): a list of scheduled event ids.
        activity_type (unicode): the type of activity to restrict the history to.

    Returns:
        a list of events from the workflow history that match the specified activity type.
    """
    if not any((scheduled_ids, activity_type)):
        raise TaranError('scheduled_ids and/or activity_type required.')
    if not scheduled_ids:
        scheduled_id_list = list()
        for event in workflow_history.get('events'):
            if (event.get('eventType') == 'ActivityTaskScheduled' and
                        event['activityTaskScheduledEventAttributes']['activityType']['name'] == activity_type):
                scheduled_id_list.append(event.get('eventId'))
        if scheduled_id_list:
            return get_activity_history(workflow_history=workflow_history,
                                        scheduled_ids=scheduled_id_list,
                                        activity_type=activity_type)
    elif scheduled_ids:
        statuses_list = list()
        for event in workflow_history.get('events'):
            event_type = event.get('eventType')
            if (event_type == 'ActivityTaskCompleted' and
                        event['activityTaskCompletedEventAttributes']['scheduledEventId'] in scheduled_ids):
                statuses_list.append(
                    {
                        'status': 'completed',
                        'event_id': event.get('eventId'),
                        'result': event['activityTaskCompletedEventAttributes']['result'],
                        'scheduled_event_id': event['activityTaskCompletedEventAttributes']['scheduledEventId']})
            elif (event_type == 'ActivityTaskScheduled' and
                          event['activityTaskScheduledEventAttributes']['activityType']['name'] == activity_type):
                statuses_list.append({'status': 'scheduled',
                                      'event_id': event.get('eventId'),
                                      'task_list': event['activityTaskScheduledEventAttributes']['taskList']['name']})
            elif (event_type == 'ActivityTaskStarted' and
                          event['activityTaskStartedEventAttributes']['scheduledEventId'] in scheduled_ids):
                statuses_list.append({'status': 'started',
                                      'event_id': event.get('eventId'),
                                      'scheduled_event_id': event['activityTaskStartedEventAttributes'][
                                          'scheduledEventId'],
                                      'identity': event['activityTaskStartedEventAttributes']['identity']})
            elif (event_type == 'ActivityTaskFailed' and
                          event['activityTaskFailedEventAttributes']['scheduledEventId'] in scheduled_ids):
                statuses_list.append({'status': 'failed',
                                      'event_id': event.get('eventId')})
            elif (event_type == 'ActivityTaskTimedOut' and
                          event['activityTaskTimedOutEventAttributes']['scheduledEventId'] in scheduled_ids):
                statuses_list.append({'status': 'timed_out',
                                      'event_id': event.get('eventId')})
            elif (event_type == 'ActivityTaskCanceled' and
                          event['activityTaskCanceledEventAttributes']['scheduledEventId'] in scheduled_ids):
                statuses_list.append({'status': 'cancelled',
                                      'event_id': event.get('eventId')})
            elif (event_type == 'ActivityTaskCancelRequested' and
                          event['activityTaskCancelRequestedEventAttributes']['scheduledEventId'] in scheduled_ids):
                statuses_list.append({'status': 'cancel_requested',
                                      'event_id': event.get('eventId')})
            elif event_type.startswith('Decision'):
                pass
        return statuses_list
