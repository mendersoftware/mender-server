// Copyright 2020 Northern.tech AS
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
import { maxSessionAge } from '../src/js/auth';
import { SORTING_OPTIONS } from '../src/js/constants/appConstants';
import * as DeviceConstants from '../src/js/constants/deviceConstants';
import { ALL_RELEASES } from '../src/js/constants/releaseConstants';
import {
  defaultPermissionSets,
  emptyRole,
  emptyUiPermissions,
  rolesById,
  rolesByName,
  scopedPermissionAreas,
  twoFAStates,
  uiPermissionsById
} from '../src/js/constants/userConstants';
import { initialState as initialAppState } from '../src/js/reducers/appReducer';
import { initialState as initialDeploymentsState } from '../src/js/reducers/deploymentReducer';
import { initialState as initialDevicesState } from '../src/js/reducers/deviceReducer';
import { initialState as initialMonitorState } from '../src/js/reducers/monitorReducer';
import { initialState as initialOnboardingState } from '../src/js/reducers/onboardingReducer';
import { initialState as initialOrganizationState } from '../src/js/reducers/organizationReducer';
import { initialState as initialReleasesState } from '../src/js/reducers/releaseReducer';
import { initialState as initialUsersState } from '../src/js/reducers/userReducer';

export const undefineds = /undefined|\[object Object\]/;
export const menderEnvironment = {
  features: {
    hasMultitenancy: true
  },
  integrationVersion: 'saas-123.34',
  menderVersion: 'next',
  metaMenderVersion: 'saas-123.34',
  services: {
    deploymentsVersion: '1.2.3',
    deviceauthVersion: null,
    inventoryVersion: null
  }
};
export const TEST_SESSION_DATETIME = '2019-01-13T13:00:00.000Z';
const testDate = new Date(TEST_SESSION_DATETIME);
export const mockDate = new Date(testDate.setMilliseconds(testDate.getMilliseconds() + maxSessionAge));

export const defaultPassword = 'mysecretpassword!123';
export const defaultCreationDate = '2019-01-13T06:25:00.000Z';
export const defaultMacAddress = 'dc:a6:32:12:ad:bf';

export const testSsoId = 'sso_id';

const deviceTypes = { qemu: 'qemux86-64' };
const permissionSetObjectTypes = {
  any: 'Any',
  artifacts: 'Artifacts',
  empty: '',
  groups: 'Device groups',
  releases: 'Releases',
  userManagement: 'User management'
};

const commonEndpoints = {
  artifacts: '^/api/management/v1/deployments/artifacts',
  artifactDetails: '^/api/management/v1/deployments/artifacts/[^/]+',
  deviceManagement: '^/api/management/(v[1-9])/(devauth|inventory)/'
};

export const token =
  'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiJjZTNkMGY4Yy1hZWRlLTQwMzAtYjM5MS03ZDUwMjBlYjg3M2UiLCJzdWIiOiJhMzBhNzgwYi1iODQzLTUzNDQtODBlMy0wZmQ5NWE0ZjZmYzMiLCJleHAiOjE2MDY4MTUzNjksImlhdCI6MTYwNjIxMDU2OSwibWVuZGVyLnRlbmFudCI6IjVmODVjMTdiY2U2MmI3ZmE3ZjVmNzA0MCIsIm1lbmRlci51c2VyIjp0cnVlLCJpc3MiOiJNZW5kZXIgVXNlcnMiLCJzY3AiOiJtZW5kZXIuKiIsIm1lbmRlci5wbGFuIjoicHJvZmVzc2lvbmFsIiwibmJmIjoxNjA2MjEwNTY5fQ.qVgYdCzLTf8OdK9uUctqqaY_HWkIiwpekuGvuGQAXCEgOv4bRNDlZRN_ZRSbxQoARG3pquhScbQrjBV9tcF4irTUPlTn3yrsXNO17DpcbTVeKRkb88RDtIKiRw3orVZ_GlIb-ckTQ5dS-Nqlyyf3Fmrhca-gwt6m_xv2UrmJK6eYYTMfggdRRWb-4u7mEkBI_pHPMTQrT8kJ2BeX-vHgazH9AoH0k85LHtFZQXD7pXHlDZRnLxJXukncwMGDmF17374gavYAIyDIzcC8sEBMDnVXgpikeA1sauzirqix6mAVs6XmxdQO7aF0wfXO1_PTYUA3Nk1oQfMYNlEI3U9uLRJRZIq2L8fmrrBryhstKd4y0KlBbGAQrx8NtRkgajjd1ljMfPBUEZrb7uSerVjneiO-aIBO76CuH0zdklphIjpGJeogkBhe8pAYNggp1XsZHgpZfl7IE5faKaDkMGnutaea--Czor6bhqUNCuY4tR0cpQJbNwy6LS9o1CFy4Log';

export const accessTokens = [
  { id: 'some-id-1', expiration_date: '2022-06-02T11:11:21.725Z', name: 'some-name-1' },
  { id: 'some-id-2', expiration_date: '2022-06-02T11:11:21.725Z', last_used: '2022-06-02T11:05:21.725Z', name: 'some-name-2' }
];

export const webhookEvents = [
  {
    id: '1',
    type: 'device-status-changed',
    data: { id: '1', status: 'accepted' },
    time: '2020-09-01T12:00:00.000Z',
    delivery_statuses: [{ integration_id: '1', success: true, status_code: 200 }]
  },
  {
    id: '2',
    type: 'device-status-changed',
    data: { id: '3', status: 'accepted' },
    time: '2020-09-01T12:00:05.000Z',
    delivery_statuses: [{ integration_id: '1', success: true, status_code: 200 }]
  }
];

export const adminUserCapabilities = {
  canAuditlog: true,
  canConfigure: true,
  canDeploy: true,
  canManageDevices: true,
  canManageReleases: true,
  canManageUsers: true,
  canReadDeployments: true,
  canReadDevices: true,
  canReadReleases: true,
  canReadUsers: true,
  canTroubleshoot: true,
  canUploadReleases: true,
  canWriteDevices: true
};

let checkInTimeRounded = '2019-01-01T00:00:00.000Z';
let checkInTimeExact = '2019-01-01T10:25:00.000Z';

export const userId = 'a30a780b-b843-5344-80e3-0fd95a4f6fc3';
export const defaultState = {
  app: {
    ...initialAppState,
    searchState: {
      deviceIds: [],
      searchTerm: '',
      searchTotal: 0,
      sort: {}
    },
    snackbar: {},
    uploadsById: {},
    versionInformation: {}
  },
  deployments: {
    ...initialDeploymentsState,
    byId: {
      d1: {
        id: 'd1',
        name: 'test deployment',
        artifact_name: 'r1',
        artifacts: ['123'],
        created: '2019-01-01T12:30:00.000Z',
        device_count: 1,
        devices: {
          a1: {
            attributes: {},
            id: 'a1',
            image: { size: 123 },
            status: 'installing'
          }
        },
        statistics: {
          status: {
            downloading: 0,
            decommissioned: 0,
            failure: 0,
            installing: 1,
            noartifact: 0,
            pending: 0,
            rebooting: 0,
            success: 0,
            'already-installed': 0
          },
          total_size: 1234
        }
      },
      d2: {
        id: 'd2',
        name: 'test deployment 2',
        artifact_name: 'r1',
        artifacts: ['123'],
        created: '2019-01-01T12:25:00.000Z',
        device_count: 1,
        devices: {
          b1: {
            attributes: {},
            id: 'b1',
            status: 'pending'
          }
        },
        statistics: {
          status: {
            downloading: 0,
            decommissioned: 0,
            failure: 0,
            installing: 0,
            noartifact: 0,
            pending: 1,
            rebooting: 0,
            success: 0,
            'already-installed': 0
          }
        }
      }
    },
    byStatus: {
      finished: { deploymentIds: ['d1'], total: 1 },
      inprogress: { deploymentIds: ['d1'], total: 1 },
      pending: { deploymentIds: ['d2'], total: 1 },
      scheduled: { deploymentIds: ['d2'], total: 1 }
    },
    deploymentDeviceLimit: 500,
    selectedDeviceIds: [],
    selectionState: {
      finished: {
        ...DeviceConstants.DEVICE_LIST_DEFAULTS,
        selection: ['d1'],
        endDate: undefined,
        search: '',
        total: 1,
        type: ''
      },
      inprogress: { ...DeviceConstants.DEVICE_LIST_DEFAULTS, selection: ['d1'], total: 1 },
      pending: { ...DeviceConstants.DEVICE_LIST_DEFAULTS, selection: ['d2'], total: 1 },
      scheduled: { ...DeviceConstants.DEVICE_LIST_DEFAULTS, selection: ['d2'], total: 1 },
      general: {
        state: 'active',
        showCreationDialog: false,
        showReportDialog: false,
        reportType: null
      },
      selectedId: 'd1'
    }
  },
  devices: {
    ...initialDevicesState,
    byId: {
      a1: {
        id: 'a1',
        attributes: {
          device_type: ['raspberrypi4'],
          ipv4_wlan0: '192.168.10.141/24'
        },
        system: {
          check_in_time: checkInTimeRounded
        },
        check_in_time: checkInTimeExact,
        check_in_time_exact: checkInTimeExact,
        check_in_time_rounded: checkInTimeRounded,
        identity_data: { mac: defaultMacAddress },
        status: 'accepted',
        decommissioning: false,
        created_ts: defaultCreationDate,
        updated_ts: '2019-01-01T09:25:00.000Z',
        auth_sets: [
          {
            id: 'auth1',
            identity_data: { mac: defaultMacAddress },
            pubkey: '-----BEGIN PUBLIC KEY-----\nMIIBojWELzgJ62hcXIhAfqfoNiaB1326XZByZwcnHr5BuSPAgMBAAE=\n-----END PUBLIC KEY-----\n',
            ts: defaultCreationDate,
            status: 'accepted'
          }
        ]
      },
      b1: {
        id: 'b1',
        attributes: {
          ipv4_wlan0: '192.168.10.141/24',
          device_type: [deviceTypes.qemu]
        },
        system: {
          check_in_time: checkInTimeRounded
        },
        check_in_time: checkInTimeExact,
        check_in_time_exact: checkInTimeExact,
        check_in_time_rounded: checkInTimeRounded,
        identity_data: { mac: defaultMacAddress },
        status: 'accepted',
        decommissioning: false,
        created_ts: defaultCreationDate,
        updated_ts: '2019-01-01T09:25:00.000Z',
        auth_sets: [
          {
            id: 'auth1',
            identity_data: { mac: defaultMacAddress },
            pubkey: '-----BEGIN PUBLIC KEY-----\nMIIBojWELzgJ62hcXIhAfqfoNiaB1326XZByZwcnHr5BuSPAgMBAAE=\n-----END PUBLIC KEY-----\n',
            ts: defaultCreationDate,
            status: 'accepted'
          }
        ]
      },
      c1: {
        id: 'c1',
        auth_sets: [],
        attributes: {
          device_type: ['qemux86-128']
        }
      }
    },
    byStatus: {
      accepted: { deviceIds: ['a1', 'b1'], total: 2 },
      active: { deviceIds: [], total: 0 },
      inactive: { deviceIds: [], total: 0 },
      pending: { deviceIds: ['c1'], total: 1 },
      preauthorized: { deviceIds: [], total: 0 },
      rejected: { deviceIds: [], total: 0 }
    },
    deviceList: {
      deviceIds: [],
      isLoading: false,
      page: 1,
      perPage: 20,
      selectedAttributes: [],
      selectedIssues: [],
      selection: [],
      sort: {
        direction: SORTING_OPTIONS.desc
        // key: null,
        // scope: null
      },
      state: DeviceConstants.DEVICE_STATES.accepted,
      total: 0
    },
    filteringAttributes: {
      identityAttributes: ['mac'],
      inventoryAttributes: ['artifact_name'],
      systemAttributes: [],
      tagAttributes: []
    },
    filteringAttributesLimit: 10,
    filters: [],
    groups: {
      byId: {
        testGroup: {
          deviceIds: ['a1', 'b1'],
          filters: [],
          total: 2
        },
        testGroupDynamic: {
          id: 'filter1',
          filters: [{ scope: 'system', key: 'group', operator: '$eq', value: 'things' }]
        }
      },
      selectedGroup: undefined
    },
    limit: 500
  },
  onboarding: {
    ...initialOnboardingState,
    progress: undefined,
    complete: false,
    demoArtifactPort: 85,
    showConnectDeviceDialog: false,
    showTipsDialog: false
  },
  monitor: {
    ...initialMonitorState,
    alerts: {
      alertList: { ...DeviceConstants.DEVICE_LIST_DEFAULTS, total: 0 },
      byDeviceId: {
        a1: {
          alerts: [
            {
              description: 'something',
              id: '31346239-3839-6262-2d63-3365622d3437',
              name: 'SSH Daemon is not running',
              device_id: 'a1',
              level: 'CRITICAL',
              subject: {
                name: 'sshd',
                type: 'systemd',
                status: 'not-running',
                details: { description: 'Jul 22 10:40:56 raspberrypi sshd[32031]: pam_unix(sshd:session): session closed for user root' }
              },
              timestamp: '2021-07-23T12:22:36Z'
            }
          ],
          latest: []
        }
      }
    },
    issueCounts: {
      byType: Object.values(DeviceConstants.DEVICE_ISSUE_OPTIONS).reduce(
        (accu, { isCategory, key }) => {
          if (isCategory) {
            return accu;
          }
          const current = accu[key] ?? {};
          accu[key] = { filtered: 0, total: 0, ...current };
          return accu;
        },
        {
          [DeviceConstants.DEVICE_ISSUE_OPTIONS.authRequests.key]: { filtered: 0, total: 0 },
          [DeviceConstants.DEVICE_ISSUE_OPTIONS.monitoring.key]: { filtered: 3, total: 0 },
          [DeviceConstants.DEVICE_ISSUE_OPTIONS.offline.key]: { filtered: 0, total: 0 }
        }
      )
    },
    settings: {
      global: {
        channels: { email: { enabled: true } }
      }
    }
  },
  organization: {
    ...initialOrganizationState,
    card: {
      brand: 'testCorp',
      last4: '7890',
      expiration: { month: 1, year: 2024 }
    },
    auditlog: {
      events: [
        {
          actor: {
            id: 'string',
            type: 'user',
            email: 'string@example.com'
          },
          time: '2019-01-01T12:10:22.667Z',
          action: 'create',
          object: {
            id: 'string',
            type: 'user',
            user: {
              email: 'user@acme.com'
            }
          },
          change: 'change1'
        },
        {
          actor: {
            id: 'string',
            type: 'user',
            email: 'string',
            identity_data: 'string'
          },
          time: '2019-01-01T12:16:22.667Z',
          action: 'create',
          object: {
            id: 'string',
            type: 'deployment',
            deployment: {
              name: 'production',
              artifact_name: 'Application 0.0.1'
            }
          },
          change: 'change2'
        },
        {
          actor: {
            id: 'string',
            type: 'user',
            email: 'string@example.com'
          },
          time: '2019-01-01T12:10:22.669Z',
          action: 'open_terminal',
          meta: {
            session_id: ['abd313a8-ee88-48ab-9c99-fbcd80048e6e']
          },
          object: {
            id: 'a1',
            type: 'device'
          },
          change: 'change3'
        }
      ],
      selectionState: {
        ...DeviceConstants.DEVICE_LIST_DEFAULTS,
        sort: {},
        total: 3
      }
    },
    intentId: 'testIntent',
    organization: {
      addons: [],
      id: 1,
      name: 'test',
      plan: 'os',
      trial: false
    }
  },
  releases: {
    ...initialReleasesState,
    byId: {
      r1: {
        name: 'r1',
        artifacts: [
          {
            id: 'art1',
            description: 'test description',
            device_types_compatible: [deviceTypes.qemu],
            modified: '2020-09-10T12:16:22.667Z',
            updates: [{ type_info: 'testtype' }],
            artifact_depends: {
              device_type: [deviceTypes.qemu]
            },
            artifact_provides: {
              artifact_name: 'myapp',
              'data-partition.myapp.version': 'v2020.10',
              list_of_fancy: [deviceTypes.qemu, 'x172']
            },
            clears_artifact_provides: ['data-partition.myapp.*']
          }
        ],
        device_types_compatible: [deviceTypes.qemu],
        modified: '2020-09-10T12:16:22.667Z',
        metaData: {}
      }
    },
    releasesList: {
      ...DeviceConstants.DEVICE_LIST_DEFAULTS,
      searchedIds: [],
      isLoading: false,
      releaseIds: ['r1'],
      selection: [],
      sort: {
        direction: SORTING_OPTIONS.desc,
        key: 'name'
      },
      searchTerm: '',
      searchTotal: 0,
      total: 1
    }
  },
  users: {
    ...initialUsersState,
    byId: {
      a1: { email: 'a@b.com', id: 'a1', created_ts: '2019-01-01T10:30:00.000Z', roles: [rolesByName.admin], verified: true },
      [userId]: { email: 'a2@b.com', id: userId, created_ts: '2019-01-01T12:30:00.000Z', roles: [rolesByName.admin], tfa_status: twoFAStates.enabled }
    },
    currentUser: 'a1',
    globalSettings: { '2fa': 'enabled', id_attribute: undefined, previousFilters: [] },
    rolesById: {
      ...rolesById,
      test: {
        ...emptyRole,
        name: 'test',
        description: 'test description',
        editable: true,
        uiPermissions: {
          ...emptyUiPermissions,
          groups: { testGroup: [uiPermissionsById.read.value] },
          releases: { bar: [uiPermissionsById.read.value] }
        }
      }
    },
    settingsInitialized: true,
    userSettings: { columnSelection: [], onboarding: { something: 'here' }, tooltips: {} }
  }
};

export const releasesList = Array.from({ length: 5000 }, (x, i) => ({
  ...defaultState.releases.byId.r1,
  name: `release-${i + 1}`,
  modified: i
}));

export const permissionSets = [
  {
    ...defaultPermissionSets.Basic,
    object: permissionSetObjectTypes.empty,
    description: 'Set containing basic permissions.',
    permissions: [
      { action: 'http', object: { type: 'any', value: '^/api/management/v1/useradm/settings$' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/v1/useradm/users/me$' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/1.0/auth/verify$' } },
      { action: 'http', object: { type: 'PUT', value: '^/api/management/v1/useradm/2faverify$' } },
      { action: 'http', object: { type: 'POST', value: '^/api/management/v1/useradm/users/me/2fa/(enable|disable)$' } },
      { action: 'http', object: { type: 'GET', value: '^/api/management/(v[1-9])/useradm/roles' } },
      { action: 'http', object: { type: 'GET', value: '^/api/management/(v[1-9])/tenantadm/user/tenant' } }
    ]
  },
  {
    ...defaultPermissionSets.ManageReleases,
    action: uiPermissionsById.manage.title,
    object: permissionSetObjectTypes.releases,
    description: 'Set of permissions which allows user to manage releases',
    permissions: [
      { action: 'http', object: { type: 'GET', value: commonEndpoints.artifacts } },
      { action: 'http', object: { type: 'GET', value: commonEndpoints.artifactDetails } },
      { action: 'http', object: { type: 'PUT', value: commonEndpoints.artifactDetails } },
      { action: 'http', object: { type: 'DELETE', value: commonEndpoints.artifactDetails } },
      { action: 'http', object: { type: 'GET', value: '^/api/management/v1/deployments/artifacts/[^/]+/download' } }
    ]
  },
  {
    ...defaultPermissionSets.ReadUsers,
    action: uiPermissionsById.read.title,
    object: permissionSetObjectTypes.userManagement,
    description: 'Set of permissions which allows user to view other users',
    permissions: [{ action: 'http', object: { type: 'GET', value: '^/api/management/(v[1-9])/useradm/' } }]
  },
  {
    ...defaultPermissionSets.ManageUsers,
    action: uiPermissionsById.manage.title,
    object: permissionSetObjectTypes.userManagement,
    description: 'Set of permissions which allows user manage other user accounts',
    permissions: [{ action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/useradm/' } }]
  },
  {
    ...defaultPermissionSets.ReadAuditLogs,
    action: uiPermissionsById.read.title,
    object: 'System audit log',
    description: 'Set of permissions which allows user to view system audit log',
    permissions: [{ action: 'http', object: { type: 'GET', value: '^/api/management/(v[1-9]|0.1.0)/auditlogs/logs' } }]
  },
  {
    ...defaultPermissionSets.DeployToDevices,
    action: 'Deploy',
    object: permissionSetObjectTypes.groups,
    description: 'Set of permissions which allows user to deploy to devices',
    permissions: [{ action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/(deployments|deviceconfig)/' } }],
    supported_scope_types: [scopedPermissionAreas.groups.scopeType]
  },
  {
    ...defaultPermissionSets.ConfigureDevices,
    action: 'Configure',
    object: permissionSetObjectTypes.groups,
    description: 'Set of permissions which allows user to manage configuration of the devices',
    permissions: [{ action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/deviceconfig/' } }],
    supported_scope_types: [scopedPermissionAreas.groups.scopeType]
  },
  {
    ...defaultPermissionSets.ConnectToDevices,
    action: 'Connect',
    object: permissionSetObjectTypes.groups,
    description: 'Set of permissions which allows user to use remote terminal and file transfer',
    permissions: [
      { action: 'http', object: { type: 'GET', value: '^/api/management/(v[1-9]|0.1.0)/deviceconnect/devices/[^/]+/connect$' } },
      { action: 'http', object: { type: 'GET', value: '^/api/management/(v[1-9]|0.1.0)/deviceconnect/devices/[^/]+/download\\?path=[^\u0026]+$' } },
      { action: 'http', object: { type: 'PUT', value: '^/api/management/(v[1-9]|0.1.0)/deviceconnect/devices/[^/]+/upload$' } }
    ],
    supported_scope_types: [scopedPermissionAreas.groups.scopeType]
  },
  {
    ...defaultPermissionSets.SuperUser,
    action: 'Any',
    object: permissionSetObjectTypes.any,
    description: 'Set of permissions which allows user to do anything',
    permissions: [{ action: 'any', object: { type: 'any', value: 'any' } }]
  },
  {
    ...defaultPermissionSets.UploadArtifacts,
    action: 'Upload',
    object: permissionSetObjectTypes.artifacts,
    description: 'Set of permissions which allows user to upload artifacts',
    permissions: [
      { action: 'http', object: { type: 'POST', value: commonEndpoints.artifacts } },
      { action: 'http', object: { type: 'POST', value: '^/api/management/v1/deployments/artifacts/generate' } }
    ]
  },
  {
    ...defaultPermissionSets.ReadDevices,
    action: uiPermissionsById.read.title,
    object: permissionSetObjectTypes.groups,
    description: 'Set of permissions which allows user to view devices',
    permissions: [
      { action: 'http', object: { type: 'POST', value: '^/api/management/v2/inventory/filters/search' } },
      { action: 'http', object: { type: 'POST', value: '^/api/management/v1/reporting/devices/search' } },
      { action: 'http', object: { type: 'GET', value: '^/api/management/(v[1-9])/(deployments|devauth|inventory|deviceconfig|devicemonitor)/' } },
      { action: 'http', object: { type: 'GET', value: '^/api/management/(v[1-9]|0.1.0)/deviceconnect/devices/[^/]+$' } }
    ],
    supported_scope_types: [scopedPermissionAreas.groups.scopeType]
  },
  {
    ...defaultPermissionSets.ManageDevices,
    action: uiPermissionsById.manage.title,
    object: permissionSetObjectTypes.groups,
    description: 'Set of permissions which allows user to manage devices',
    permissions: [
      { action: 'http', object: { type: 'POST', value: commonEndpoints.deviceManagement } },
      { action: 'http', object: { type: 'PUT', value: commonEndpoints.deviceManagement } },
      { action: 'http', object: { type: 'DELETE', value: commonEndpoints.deviceManagement } }
    ],
    supported_scope_types: [scopedPermissionAreas.groups.scopeType]
  },
  {
    ...defaultPermissionSets.ReadReleases,
    action: uiPermissionsById.read.title,
    object: permissionSetObjectTypes.releases,
    description: 'Set of permissions which allows user to view releases',
    permissions: [
      { action: 'http', object: { type: 'GET', value: commonEndpoints.artifacts } },
      { action: 'http', object: { type: 'GET', value: commonEndpoints.artifactDetails } },
      { action: 'http', object: { type: 'GET', value: '^/api/management/v1/deployments/artifacts/[^/]+/download' } }
    ]
  },
  {
    name: 'almostAdmin',
    object: permissionSetObjectTypes.empty,
    description: 'Set containing all the permissions.',
    permissions: [
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/auditlogs/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/deployments/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/deployments/config/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/deployments/deployments/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/deployments/deployments/releases/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/devauth/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/deviceconfig/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/deviceconnect/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/deviceconnect/devices' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/inventory/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/iot-manager/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/monitor/.*' } },
      { action: 'http', object: { type: 'any', value: '^/api/management/(v[1-9])/useradm/.*' } }
    ]
  }
];

export const rbacRoles = [
  {
    name: 'dyn',
    description: '',
    permissions: [
      { action: 'CREATE_DEPLOYMENT', object: { type: 'DEVICE_GROUP', value: 'dyn' } },
      { action: 'VIEW_DEVICE', object: { type: 'DEVICE_GROUP', value: 'dyn' } }
    ]
  },
  { name: 'asdasd', description: '123', permissions: [{ action: 'http', object: { type: 'any', value: '/api/management/v1/useradm/.*' } }] },
  {
    name: '141sasd',
    description: '1313adg',
    permission_sets_with_scope: [
      { ...defaultPermissionSets.ReadDevices, scope: { type: scopedPermissionAreas.groups.scopeType, value: ['bestgroup'] } },
      { ...defaultPermissionSets.ConnectToDevices, scope: { type: scopedPermissionAreas.groups.scopeType, value: ['bestgroup'] } },
      { ...defaultPermissionSets.ManageUsers }
    ]
  },
  {
    name: 'kljlkk',
    description: 'lkl',
    permission_sets_with_scope: [{ ...defaultPermissionSets.ConnectToDevices, scope: { type: scopedPermissionAreas.groups.scopeType, value: ['bestgroup'] } }]
  },
  {
    name: 'yyyyy',
    description: 'asd',
    permission_sets_with_scope: [
      { ...defaultPermissionSets.ManageDevices, scope: { type: scopedPermissionAreas.groups.scopeType, value: ['dockerclient'] } },
      { ...defaultPermissionSets.ManageReleases }
    ]
  },
  {
    name: 'RBAC_ROLE_DEPLOYMENTS_MANAGER',
    description: 'Intended for users responsible for managing deployments, this role can create and abort deployments',
    permission_sets_with_scope: [{ ...defaultPermissionSets.DeployToDevices }]
  },
  {
    name: 'RBAC_ROLE_REMOTE_TERMINAL',
    description: `Intended for tech support accounts, this role can access the devices' Remote Terminal.`,
    permission_sets_with_scope: [{ ...defaultPermissionSets.ConnectToDevices }]
  },
  { name: 'RBAC_ROLE_PERMIT_ALL', description: '', permission_sets_with_scope: [{ ...defaultPermissionSets.SuperUser }] },
  {
    name: 'RBAC_ROLE_OBSERVER',
    description:
      'Intended for team leaders or limited tech support accounts, this role can see all Devices, Artifacts and Deployment reports but not make any changes.',
    permission_sets_with_scope: [{ ...defaultPermissionSets.ReadReleases }, { ...defaultPermissionSets.ReadDevices }]
  },
  {
    name: 'RBAC_ROLE_CI',
    description:
      'Intended for automation accounts building software (e.g. CI/CD systems), this role can only manage Artifacts, including upload new Artifacts and delete Artifacts. It does not have access to Devices or Deployments.',
    permission_sets_with_scope: [
      { ...defaultPermissionSets.ReadReleases },
      { ...defaultPermissionSets.ManageReleases },
      { ...defaultPermissionSets.UploadArtifacts }
    ]
  },
  {
    name: 'almostAdmin',
    description: 'almost admin rights',
    permissions: permissionSets.find(item => item.name === 'almostAdmin')?.permissions
  },
  {
    name: 'almostAdminNew',
    description: 'almost admin rights',
    permission_sets_with_scope: Object.values(defaultPermissionSets)
  }
];

const expectedParsedRoles = {
  '141sasd': {
    editable: true,
    isCustom: undefined,
    uiPermissions: {
      ...emptyUiPermissions,
      groups: { bestgroup: [uiPermissionsById.read.value, uiPermissionsById.connect.value] },
      userManagement: [uiPermissionsById.manage.value]
    }
  },
  asdasd: {
    editable: false,
    isCustom: true,
    uiPermissions: { ...emptyUiPermissions, userManagement: [uiPermissionsById.read.value, uiPermissionsById.manage.value] }
  },
  dyn: {
    editable: false,
    isCustom: true,
    uiPermissions: {
      ...emptyUiPermissions,
      deployments: [uiPermissionsById.deploy.value],
      groups: { dyn: [uiPermissionsById.read.value, uiPermissionsById.deploy.value] }
    }
  },
  kljlkk: {
    editable: true,
    isCustom: false,
    uiPermissions: { ...emptyUiPermissions, groups: { bestgroup: [uiPermissionsById.read.value, uiPermissionsById.connect.value] } }
  },
  yyyyy: {
    editable: true,
    isCustom: undefined,
    uiPermissions: {
      ...emptyUiPermissions,
      groups: { dockerclient: [uiPermissionsById.read.value, uiPermissionsById.manage.value] },
      releases: { [ALL_RELEASES]: [uiPermissionsById.manage.value] }
    }
  },
  almostAdmin: {
    editable: false,
    isCustom: true,
    uiPermissions: {
      ...emptyUiPermissions,
      auditlog: [uiPermissionsById.read.value],
      deployments: [uiPermissionsById.manage.value, uiPermissionsById.deploy.value, uiPermissionsById.read.value],
      groups: {
        [DeviceConstants.ALL_DEVICES]: [
          uiPermissionsById.read.value,
          // we can't assign deployment permissions to devices here, since the old path based rbac controls don't allow the scoped permissions
          // uiPermissionsById.deploy.value,
          uiPermissionsById.manage.value,
          uiPermissionsById.connect.value,
          uiPermissionsById.configure.value
        ]
      },
      releases: {
        [ALL_RELEASES]: [uiPermissionsById.read.value, uiPermissionsById.upload.value, uiPermissionsById.manage.value]
      },
      userManagement: [uiPermissionsById.read.value, uiPermissionsById.manage.value]
    }
  },
  almostAdminNew: {
    editable: true,
    isCustom: undefined,
    uiPermissions: {
      ...emptyUiPermissions,
      auditlog: [uiPermissionsById.read.value],
      deployments: [uiPermissionsById.read.value, uiPermissionsById.deploy.value, uiPermissionsById.manage.value],
      groups: {
        [DeviceConstants.ALL_DEVICES]: [
          uiPermissionsById.read.value,
          uiPermissionsById.manage.value,
          uiPermissionsById.deploy.value,
          uiPermissionsById.connect.value,
          uiPermissionsById.configure.value
        ]
      },
      releases: {
        [ALL_RELEASES]: [uiPermissionsById.manage.value, uiPermissionsById.upload.value, uiPermissionsById.read.value]
      },
      userManagement: [uiPermissionsById.read.value, uiPermissionsById.manage.value]
    }
  }
};

export const receivedRoles = rbacRoles.reduce(
  (accu, role) => {
    const { name, description, ...roleRemainder } = role;
    if (name.startsWith('RBAC')) {
      accu[name] = {
        ...defaultState.users.rolesById[name],
        ...roleRemainder,
        editable: false,
        isCustom: false,
        description: defaultState.users.rolesById[name].description ? defaultState.users.rolesById[name].description : description
      };
    } else {
      const result = expectedParsedRoles[name] ?? {};
      accu[name] = {
        ...emptyRole,
        ...role,
        ...result,
        name
      };
    }
    return accu;
  },
  { ...defaultState.users.rolesById }
);

export const receivedPermissionSets = permissionSets.reduce((accu, set) => {
  const result = defaultPermissionSets[set.name]?.result ?? {};
  accu[set.name] = {
    ...set,
    result: {
      ...emptyUiPermissions,
      ...result
    }
  };
  return accu;
}, {});
