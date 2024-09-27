// Copyright 2024 Northern.tech AS
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
// @ts-nocheck
import React from 'react';

import { mdiAws as AWS, mdiMicrosoftAzure as Azure } from '@mdi/js';

export const DEVICE_LIST_DEFAULTS = {
  page: 1,
  perPage: 20
};

export const timeUnits = {
  days: 'days',
  minutes: 'minutes',
  hours: 'hours'
};

export const ALL_DEVICES = 'All devices';
export const UNGROUPED_GROUP = { id: '*|=ungrouped=|*', name: 'Unassigned' };

export const DEVICE_LIST_MAXIMUM_LENGTH = 50;

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
const oneSecond = 1000;
export const TIMEOUTS = {
  debounceDefault: 700,
  debounceShort: 300,
  halfASecond: 0.5 * oneSecond,
  oneSecond,
  twoSeconds: 2 * oneSecond,
  threeSeconds: 3 * oneSecond,
  fiveSeconds: 5 * oneSecond,
  refreshDefault: 10 * oneSecond,
  refreshLong: 60 * oneSecond
};

export const SORTING_OPTIONS = {
  asc: 'asc',
  desc: 'desc'
};

export const DEVICE_ONLINE_CUTOFF = { interval: 1, intervalName: timeUnits.days };

export const ATTRIBUTE_SCOPES = {
  inventory: 'inventory',
  identity: 'identity',
  monitor: 'monitor',
  system: 'system',
  tags: 'tags'
};

export const defaultIdAttribute = Object.freeze({ attribute: 'id', scope: ATTRIBUTE_SCOPES.identity });

const credentialTypes = {
  aws: 'aws',
  http: 'http',
  sas: 'sas',
  x509: 'x509'
};
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
export const MAX_PAGE_SIZE = 500;

export const ALL_RELEASES = 'All releases';

export const emptyUiPermissions = Object.freeze({
  auditlog: [],
  deployments: [],
  groups: Object.freeze({}),
  releases: Object.freeze({}),
  userManagement: []
});

export const emptyRole = Object.freeze({
  name: undefined,
  description: '',
  permissions: [],
  uiPermissions: Object.freeze({ ...emptyUiPermissions })
});
