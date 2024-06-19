// Copyright 2019 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
import React from 'react';

import { mdiAws as AWS, mdiMicrosoftAzure as Azure } from '@mdi/js';

const credentialTypes = {
  aws: 'aws',
  http: 'http',
  sas: 'sas',
  x509: 'x509'
};
export const timeUnits = {
  days: 'days',
  minutes: 'minutes',
  hours: 'hours'
};

export const DEVICE_FILTERING_OPTIONS = {
  $eq: { key: '$eq', title: 'equals', shortform: '=' },
  $ne: { key: '$ne', title: 'not equal', shortform: '!=' },
  $gt: {
    key: '$gt',
    title: '>',
    shortform: '>',
    help: 'The "greater than" operator can work both on numbers and strings. In the latter case, the operator applies the lexicographical order to the value strings.'
  },
  $gte: {
    title: '>=',
    shortform: '>=',
    help: 'The "greater than or equal" operator can work both on numbers and strings. In the latter case, the operator applies the lexicographical order to the value strings.'
  },
  $lt: {
    key: '$lt',
    title: '<',
    shortform: '<',
    help: 'The "lesser than" operator can work both on numbers and strings. In the latter case, the operator applies the lexicographical order to the value strings.'
  },
  $lte: {
    title: '<=',
    shortform: '<=',
    help: 'The "lesser than or equal" operator can work both on numbers and strings. In the latter case, the operator applies the lexicographical order to the value strings.'
  },
  $ltne: {
    key: '$ltne',
    title: '$ltne',
    shortform: 'ltne',
    help: 'The "lesser than or does not exist" operator can work both on numbers and strings. In the latter case, the operator applies the lexicographical order to the value strings.'
  },
  $in: {
    key: '$in',
    title: 'in',
    shortform: 'in',
    help: 'The "in" operator accepts a list of comma-separated values. It matches if the selected field is equal to one of the specified values.'
  },
  $nin: {
    key: '$nin',
    title: 'not in',
    shortform: 'not in',
    help: `The "not in" operator accepts a list of comma-separated values. It matches if the selected field's value is not equal to any of the specified options.`
  },
  $exists: {
    key: '$exists',
    title: 'exists',
    shortform: 'exists',
    value: true,
    help: `The "exists" operator matches if the selected field's value has a value. No value needs to be provided for this operator.`
  },
  $nexists: {
    key: '$nexists',
    title: `doesn't exist`,
    shortform: `doesn't exist`,
    value: true,
    help: `The "doesn't exist" operator matches if the selected field's value has no value. No value needs to be provided for this operator.`
  },
  $regex: {
    key: '$regex',
    title: `matches regular expression`,
    shortform: `matches`,
    help: `The "regular expression" operator matches the selected field's value with a Perl compatible regular expression (PCRE), automatically anchored by ^. If the regular expression is not valid, the filter will produce no results. If you need to specify options and flags, you can provide the full regex in the format of /regex/flags, for example.`
  }
};
export const emptyFilter = { key: '', value: '', operator: DEVICE_FILTERING_OPTIONS.$eq.key, scope: 'inventory' };

export const SELECT_GROUP = 'SELECT_GROUP';
export const ADD_TO_GROUP = 'ADD_TO_GROUP';
export const ADD_DYNAMIC_GROUP = 'ADD_DYNAMIC_GROUP';
export const ADD_STATIC_GROUP = 'ADD_STATIC_GROUP';
export const REMOVE_DYNAMIC_GROUP = 'REMOVE_DYNAMIC_GROUP';
export const REMOVE_STATIC_GROUP = 'REMOVE_STATIC_GROUP';
export const REMOVE_FROM_GROUP = 'REMOVE_FROM_GROUP';
export const RECEIVE_GROUPS = 'RECEIVE_GROUPS';
export const RECEIVE_DYNAMIC_GROUPS = 'RECEIVE_DYNAMIC_GROUPS';
export const RECEIVE_DEVICE = 'RECEIVE_DEVICE';
export const RECEIVE_DEVICES = 'RECEIVE_DEVICES';
export const RECEIVE_DEVICE_CONFIG = 'RECEIVE_DEVICE_CONFIG';
export const RECEIVE_DEVICE_CONNECT = 'RECEIVE_DEVICE_CONNECT';
export const RECEIVE_GROUP_DEVICES = 'RECEIVE_GROUP_DEVICES';
export const SET_TOTAL_DEVICES = 'SET_TOTAL_DEVICES';
export const SET_ACCEPTED_DEVICES_COUNT = 'SET_ACCEPTED_DEVICES_COUNT';
export const SET_PENDING_DEVICES_COUNT = 'SET_PENDING_DEVICES_COUNT';
export const SET_REJECTED_DEVICES_COUNT = 'SET_REJECTED_DEVICES_COUNT';
export const SET_PREAUTHORIZED_DEVICES_COUNT = 'SET_PREAUTHORIZED_DEVICES_COUNT';
export const SET_FILTER_ATTRIBUTES = 'SET_FILTER_ATTRIBUTES';
export const SET_FILTERABLES_CONFIG = 'SET_FILTERABLES_CONFIG';
export const SET_DEVICE_FILTERS = 'SET_DEVICE_FILTERS';

export const SET_ACCEPTED_DEVICES = 'SET_ACCEPTED_DEVICES';
export const SET_PENDING_DEVICES = 'SET_PENDING_DEVICES';
export const SET_REJECTED_DEVICES = 'SET_REJECTED_DEVICES';
export const SET_PREAUTHORIZED_DEVICES = 'SET_PREAUTHORIZED_DEVICES';

export const SET_INACTIVE_DEVICES = 'SET_INACTIVE_DEVICES';
export const SET_DEVICE_LIST_STATE = 'SET_DEVICE_LIST_STATE';

export const SET_DEVICE_LIMIT = 'SET_DEVICE_LIMIT';

export const SET_DEVICE_REPORTS = 'SET_DEVICE_REPORTS';

export const EXTERNAL_PROVIDER = {
  'iot-core': {
    credentialsType: credentialTypes.aws,
    icon: AWS,
    title: 'AWS IoT Core',
    twinTitle: 'Device Shadow',
    provider: 'iot-core',
    enabled: true,
    deviceTwin: true,
    configHint: <>For help finding your AWS IoT Core connection string, check the AWS IoT documentation.</>
  },
  'iot-hub': {
    credentialsType: credentialTypes.sas,
    icon: Azure,
    title: 'Azure IoT Hub',
    twinTitle: 'Device Twin',
    provider: 'iot-hub',
    enabled: true,
    deviceTwin: true,
    configHint: (
      <span>
        For help finding your Azure IoT Hub connection string, look under &apos;Shared access policies&apos; in the Microsoft Azure UI as described{' '}
        {
          <a
            href="https://devblogs.microsoft.com/iotdev/understand-different-connection-strings-in-azure-iot-hub/#iothubconn"
            target="_blank"
            rel="noopener noreferrer"
          >
            here
          </a>
        }
        .
      </span>
    )
  },
  webhook: {
    credentialsType: credentialTypes.http,
    deviceTwin: false,
    // disable the webhook provider here, since it is treated different than other integrations, with a custom configuration & management view, etc.
    enabled: false,
    provider: 'webhook'
  }
};

// see https://github.com/mendersoftware/go-lib-micro/tree/master/ws
//     for the description of proto_header and the consts
// *Note*: this needs to be aligned with mender-connect and deviceconnect.
export const DEVICE_MESSAGE_PROTOCOLS = {
  Shell: 1
};
export const DEVICE_MESSAGE_TYPES = {
  Delay: 'delay',
  New: 'new',
  Ping: 'ping',
  Pong: 'pong',
  Resize: 'resize',
  Shell: 'shell',
  Stop: 'stop'
};
export const DEVICE_LIST_DEFAULTS = {
  page: 1,
  perPage: 20
};
export const DEVICE_LIST_MAXIMUM_LENGTH = 50;
export const DEVICE_ISSUE_OPTIONS = {
  issues: {
    isCategory: true,
    key: 'issues',
    title: 'Devices with issues',
    filterRule: {}
  },
  offline: {
    issueCategory: 'issues',
    key: 'offline',
    needsFullFiltering: true,
    needsMonitor: false,
    needsReporting: false,
    filterRule: {
      scope: 'system',
      key: 'check_in_time',
      operator: DEVICE_FILTERING_OPTIONS.$ltne.key,
      value: ({ offlineThreshold }) => offlineThreshold
    },
    title: 'Offline devices'
  },
  failedLastUpdate: {
    issueCategory: 'issues',
    key: 'failedLastUpdate',
    needsFullFiltering: false,
    needsMonitor: false,
    needsReporting: true,
    filterRule: { scope: 'monitor', key: 'failed_last_update', operator: DEVICE_FILTERING_OPTIONS.$eq.key, value: true },
    title: 'Deployment failed'
  },
  monitoring: {
    issueCategory: 'issues',
    key: 'monitoring',
    needsFullFiltering: false,
    needsMonitor: true,
    needsReporting: false,
    filterRule: { scope: 'monitor', key: 'alerts', operator: DEVICE_FILTERING_OPTIONS.$eq.key, value: true },
    title: 'Monitoring alert'
  },
  authRequests: {
    key: 'authRequests',
    needsFullFiltering: false,
    needsMonitor: false,
    needsReporting: true,
    filterRule: { scope: 'monitor', key: 'auth_requests', operator: DEVICE_FILTERING_OPTIONS.$gt.key, value: 1 },
    title: 'Devices with new authentication requests'
  },
  gatewayDevices: {
    key: 'gatewayDevices',
    needsFullFiltering: false,
    needsMonitor: false,
    needsReporting: true,
    filterRule: { scope: 'inventory', key: 'mender_is_gateway', operator: DEVICE_FILTERING_OPTIONS.$eq.key, value: 'true' },
    title: 'Gateway devices'
  }
};
// we can't include the dismiss state with the rest since this would include dismissed devices in several queries
export const DEVICE_DISMISSAL_STATE = 'dismiss';
export const DEVICE_STATES = {
  accepted: 'accepted',
  pending: 'pending',
  preauth: 'preauthorized',
  rejected: 'rejected'
};
export const DEVICE_CONNECT_STATES = {
  connected: 'connected',
  disconnected: 'disconnected',
  unknown: 'unknown'
};
export const DEVICE_ONLINE_CUTOFF = { interval: 1, intervalName: timeUnits.days };
export const ATTRIBUTE_SCOPES = {
  inventory: 'inventory',
  identity: 'identity',
  monitor: 'monitor',
  system: 'system',
  tags: 'tags'
};
export const ALL_DEVICES = 'All devices';
export const UNGROUPED_GROUP = { id: '*|=ungrouped=|*', name: 'Unassigned' };
