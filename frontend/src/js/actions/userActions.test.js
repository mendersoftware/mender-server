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
import { act } from '@testing-library/react';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';
import Cookies from 'universal-cookie';

import { inventoryDevice } from '../../../tests/__mocks__/deviceHandlers';
import { accessTokens, defaultPassword, defaultState, receivedPermissionSets, receivedRoles, testSsoId, token, userId } from '../../../tests/mockData';
import { HELPTOOLTIPS } from '../components/helptips/helptooltips';
import {
  SET_ANNOUNCEMENT,
  SET_ENVIRONMENT_DATA,
  SET_FEATURES,
  SET_FIRST_LOGIN_AFTER_SIGNUP,
  SET_OFFLINE_THRESHOLD,
  SET_SNACKBAR,
  SET_VERSION_INFORMATION,
  SORTING_OPTIONS
} from '../constants/appConstants';
import {
  RECEIVE_DEPLOYMENTS,
  RECEIVE_FINISHED_DEPLOYMENTS,
  RECEIVE_INPROGRESS_DEPLOYMENTS,
  SELECT_INPROGRESS_DEPLOYMENTS
} from '../constants/deploymentConstants';
import {
  ADD_DYNAMIC_GROUP,
  DEVICE_LIST_DEFAULTS,
  DEVICE_STATES,
  EXTERNAL_PROVIDER,
  RECEIVE_DEVICES,
  RECEIVE_DYNAMIC_GROUPS,
  RECEIVE_GROUPS,
  SET_ACCEPTED_DEVICES,
  SET_DEVICE_LIMIT,
  SET_DEVICE_LIST_STATE,
  SET_FILTER_ATTRIBUTES,
  SET_PENDING_DEVICES,
  SET_PREAUTHORIZED_DEVICES,
  SET_REJECTED_DEVICES,
  UNGROUPED_GROUP
} from '../constants/deviceConstants';
import { SET_DEMO_ARTIFACT_PORT, SET_ONBOARDING_COMPLETE } from '../constants/onboardingConstants';
import { RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS, getSamlStartUrl } from '../constants/organizationConstants';
import { RECEIVE_RELEASES, SET_RELEASES_LIST_STATE } from '../constants/releaseConstants';
import {
  CREATED_ROLE,
  CREATED_USER,
  RECEIVED_ACTIVATION_CODE,
  RECEIVED_PERMISSION_SETS,
  RECEIVED_QR_CODE,
  RECEIVED_ROLES,
  RECEIVED_USER,
  RECEIVED_USER_LIST,
  REMOVED_ROLE,
  REMOVED_USER,
  SET_CUSTOM_COLUMNS,
  SET_GLOBAL_SETTINGS,
  SET_SHOW_CONNECT_DEVICE,
  SET_TOOLTIPS_STATE,
  SET_TOOLTIP_STATE,
  SET_USER_SETTINGS,
  SUCCESSFULLY_LOGGED_IN,
  UPDATED_ROLE,
  UPDATED_USER,
  USER_LOGOUT,
  emptyRole,
  uiPermissionsById
} from '../constants/userConstants';
import { attributeReducer, receivedInventoryDevice } from './appActions.test';
import { expectedOnboardingActions } from './onboardingActions.test';
import {
  addUserToCurrentTenant,
  createRole,
  createUser,
  disableUser2fa,
  editRole,
  editUser,
  enableUser2fa,
  generateToken,
  get2FAQRCode,
  getRoles,
  getTokens,
  getUser,
  getUserList,
  initializeSelf,
  loginUser,
  logoutUser,
  passwordResetComplete,
  passwordResetStart,
  removeRole,
  removeUser,
  revokeToken,
  saveGlobalSettings,
  saveUserSettings,
  setAccountActivationCode,
  setAllTooltipsReadState,
  setHideAnnouncement,
  setShowConnectingDialog,
  setTooltipReadState,
  switchUserOrganization,
  updateUserColumnSettings,
  verify2FA,
  verifyEmailComplete,
  verifyEmailStart
} from './userActions';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

const defaultRole = { ...emptyRole, name: 'test', description: 'test description' };
const settings = { test: true };

// eslint-disable-next-line no-unused-vars
const { attributes, ...expectedDevice } = defaultState.devices.byId.a1;

const offlineThreshold = { type: SET_OFFLINE_THRESHOLD, value: '2019-01-12T13:00:06.900Z' };
const appInitActions = [
  { type: RECEIVED_USER, user: defaultState.users.byId[userId] },
  { type: SET_CUSTOM_COLUMNS, value: [] },
  {
    type: SUCCESSFULLY_LOGGED_IN,
    value: {
      token: undefined
    }
  },
  { type: SET_ONBOARDING_COMPLETE, complete: false },
  { type: SET_DEMO_ARTIFACT_PORT, value: 85 },
  { type: SET_FEATURES, value: { ...defaultState.app.features, hasMultitenancy: true } },
  {
    type: SET_VERSION_INFORMATION,
    docsVersion: '',
    value: {
      Deployments: '1.2.3',
      Deviceauth: null,
      GUI: undefined,
      Integration: 'master',
      Inventory: null,
      'Mender-Artifact': undefined,
      'Mender-Client': 'next',
      'Meta-Mender': 'saas-123.34'
    }
  },
  { type: SET_ENVIRONMENT_DATA, value: { hostAddress: null, hostedAnnouncement: '', recaptchaSiteKey: '', stripeAPIKey: '', trackerCode: '' } },
  { type: SET_FIRST_LOGIN_AFTER_SIGNUP, firstLoginAfterSignup: false },
  { type: RECEIVE_DEPLOYMENTS, deployments: defaultState.deployments.byId },
  {
    type: RECEIVE_FINISHED_DEPLOYMENTS,
    deploymentIds: Object.keys(defaultState.deployments.byId),
    status: 'finished',
    total: Object.keys(defaultState.deployments.byId).length
  },
  { type: RECEIVE_DEPLOYMENTS, deployments: defaultState.deployments.byId },
  {
    type: RECEIVE_INPROGRESS_DEPLOYMENTS,
    deploymentIds: Object.keys(defaultState.deployments.byId),
    status: 'inprogress',
    total: Object.keys(defaultState.deployments.byId).length
  },
  {
    type: SELECT_INPROGRESS_DEPLOYMENTS,
    deploymentIds: Object.keys(defaultState.deployments.byId),
    status: 'inprogress'
  },
  {
    type: SET_FILTER_ATTRIBUTES,
    attributes: {
      identityAttributes: ['status', 'mac'],
      inventoryAttributes: [
        'artifact_name',
        'cpu_model',
        'device_type',
        'hostname',
        'ipv4_wlan0',
        'ipv6_wlan0',
        'kernel',
        'mac_eth0',
        'mac_wlan0',
        'mem_total_kB',
        'mender_bootloader_integration',
        'mender_client_version',
        'network_interfaces',
        'os',
        'rootfs_type'
      ],
      systemAttributes: ['created_ts', 'updated_ts', 'group'],
      tagAttributes: []
    }
  },
  { type: SET_DEVICE_LIMIT, limit: 500 },
  {
    type: RECEIVE_GROUPS,
    groups: {
      testGroup: defaultState.devices.groups.byId.testGroup,
      testGroupDynamic: {
        filters: [{ key: 'group', operator: '$eq', scope: 'system', value: 'things' }],
        id: 'filter1'
      }
    }
  },
  {
    type: RECEIVE_DYNAMIC_GROUPS,
    groups: {
      testGroup: defaultState.devices.groups.byId.testGroup,
      testGroupDynamic: {
        deviceIds: [],
        filters: [
          { key: 'id', operator: '$in', scope: 'identity', value: [defaultState.devices.byId.a1.id] },
          { key: 'mac', operator: '$nexists', scope: 'identity', value: false },
          { key: 'kernel', operator: '$exists', scope: 'identity', value: true }
        ],
        id: 'filter1',
        total: 0
      }
    }
  },
  {
    type: RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS,
    value: [
      { connection_string: 'something_else', id: 1, provider: EXTERNAL_PROVIDER['iot-hub'].provider },
      { id: 2, provider: EXTERNAL_PROVIDER['iot-core'].provider, something: 'new' }
    ]
  },
  { type: RECEIVE_RELEASES, releases: defaultState.releases.byId },
  {
    type: SET_RELEASES_LIST_STATE,
    value: {
      ...defaultState.releases.releasesList,
      releaseIds: [
        'release-999',
        'release-998',
        'release-997',
        'release-996',
        'release-995',
        'release-994',
        'release-993',
        'release-992',
        'release-991',
        'release-990',
        'release-99',
        'release-989',
        'release-988',
        'release-987',
        'release-986',
        'release-985',
        'release-984',
        'release-983',
        'release-982',
        'release-981'
      ],
      page: 1,
      total: 5000
    }
  },
  { type: SET_GLOBAL_SETTINGS, settings: { '2fa': 'enabled', previousFilters: [] } },
  offlineThreshold,
  { type: SET_USER_SETTINGS, settings: { ...defaultState.users.userSettings } },
  { type: RECEIVED_PERMISSION_SETS, value: receivedPermissionSets },
  { type: RECEIVED_ROLES, value: receivedRoles },
  {
    type: RECEIVE_DEVICES,
    devicesById: { [expectedDevice.id]: { ...defaultState.devices.byId.a1, isOffline: true, monitor: {}, tags: {} } }
  },
  {
    type: RECEIVE_DEVICES,
    devicesById: {
      a1: {
        ...defaultState.devices.byId.a1,
        attributes: inventoryDevice.attributes.reduce(attributeReducer, {}),
        group: 'test',
        identity_data: { ...defaultState.devices.byId.a1.identity_data, status: DEVICE_STATES.accepted },
        isOffline: true,
        monitor: {},
        tags: {},
        updated_ts: inventoryDevice.updated_ts
      }
    }
  },
  {
    type: SET_ACCEPTED_DEVICES,
    deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id],
    status: DEVICE_STATES.accepted,
    total: defaultState.devices.byStatus.accepted.deviceIds.length
  },
  {
    type: RECEIVE_DEVICES,
    devicesById: {
      a1: {
        ...defaultState.devices.byId.a1,
        attributes: inventoryDevice.attributes.reduce(attributeReducer, {}),
        group: 'test',
        identity_data: { ...defaultState.devices.byId.a1.identity_data, status: DEVICE_STATES.accepted },
        isOffline: true,
        monitor: {},
        status: 'pending',
        tags: {},
        updated_ts: inventoryDevice.updated_ts
      }
    }
  },
  {
    type: SET_PENDING_DEVICES,
    deviceIds: Array.from({ length: defaultState.devices.byStatus.pending.total }, () => defaultState.devices.byId.a1.id),
    status: 'pending',
    total: defaultState.devices.byStatus.pending.deviceIds.length
  },
  { type: RECEIVE_DEVICES, devicesById: {} },
  { type: SET_PREAUTHORIZED_DEVICES, deviceIds: [], status: 'preauthorized', total: 0 },
  { type: RECEIVE_DEVICES, devicesById: {} },
  { type: SET_REJECTED_DEVICES, deviceIds: [], status: 'rejected', total: 0 },
  {
    type: RECEIVE_DEVICES,
    devicesById: { [expectedDevice.id]: { ...defaultState.devices.byId.a1, group: undefined, isOffline: true, monitor: {}, tags: {} } }
  },
  {
    type: RECEIVE_DEVICES,
    devicesById: {
      [expectedDevice.id]: {
        ...receivedInventoryDevice,
        attributes: defaultState.devices.byId.a1.attributes,
        identity_data: defaultState.devices.byId.a1.identity_data
      }
    }
  },
  {
    type: RECEIVE_DEVICES,
    devicesById: { [expectedDevice.id]: { ...defaultState.devices.byId.a1, group: undefined, isOffline: true, monitor: {}, tags: {} } }
  },
  {
    type: ADD_DYNAMIC_GROUP,
    groupName: UNGROUPED_GROUP.id,
    group: { deviceIds: [], total: 0, filters: [{ key: 'group', value: ['testGroup'], operator: '$nin', scope: 'system' }] }
  },
  {
    type: SET_DEVICE_LIST_STATE,
    state: {
      ...DEVICE_LIST_DEFAULTS,
      deviceIds: [],
      isLoading: true,
      selectedAttributes: [],
      selectedIssues: [],
      selection: [],
      setOnly: false,
      sort: { direction: SORTING_OPTIONS.desc },
      state: DEVICE_STATES.accepted,
      total: 0
    }
  },
  { type: SET_TOOLTIPS_STATE, value: {} },
  { type: SET_GLOBAL_SETTINGS, settings: { '2fa': 'enabled', previousFilters: [] } },
  offlineThreshold,
  { type: SET_USER_SETTINGS, settings: { ...defaultState.users.userSettings } },
  { type: SET_GLOBAL_SETTINGS, settings: { '2fa': 'enabled', previousFilters: [] } },
  { type: SET_USER_SETTINGS, settings: { ...defaultState.users.userSettings } },
  { type: RECEIVE_DEVICES, devicesById: { [expectedDevice.id]: { ...receivedInventoryDevice, group: 'test' } } },
  {
    type: SET_ACCEPTED_DEVICES,
    deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id],
    status: DEVICE_STATES.accepted,
    total: defaultState.devices.byStatus.accepted.total
  },
  {
    type: RECEIVE_DEVICES,
    devicesById: { [expectedDevice.id]: { ...defaultState.devices.byId.a1, group: undefined, isOffline: true, monitor: {}, tags: {} } }
  },
  {
    type: SET_DEVICE_LIST_STATE,
    state: {
      ...DEVICE_LIST_DEFAULTS,
      deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id],
      isLoading: false,
      selectedAttributes: [],
      selectedIssues: [],
      selection: [],
      sort: { direction: SORTING_OPTIONS.desc },
      state: DEVICE_STATES.accepted,
      total: 2
    }
  },
  ...expectedOnboardingActions
];

/* eslint-disable sonarjs/no-identical-functions */
describe('user actions', () => {
  it('should forward connecting dialog visibility', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      {
        type: SET_SHOW_CONNECT_DEVICE,
        show: true
      }
    ];
    await store.dispatch(setShowConnectingDialog(true));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow retrieving 2fa qr codes', async () => {
    jest.clearAllMocks();
    const expectedActions = [{ type: RECEIVED_QR_CODE, value: btoa('test') }];
    const store = mockStore({ ...defaultState });
    await store.dispatch(get2FAQRCode(true));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should verify 2fa codes during 2fa setup', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: RECEIVED_USER, user: defaultState.users.byId[userId] },
      { type: SET_CUSTOM_COLUMNS, value: [] }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(verify2FA({ token2fa: '123456' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow enabling 2fa during 2fa setup', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: RECEIVED_USER, user: defaultState.users.byId.a1 },
      { type: SET_CUSTOM_COLUMNS, value: [] }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(enableUser2fa(defaultState.users.byId.a1.id));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow disabling 2fa during 2fa setup', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: RECEIVED_USER, user: defaultState.users.byId.a1 },
      { type: SET_CUSTOM_COLUMNS, value: [] }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(disableUser2fa(defaultState.users.byId.a1.id));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow beginning email verification', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: RECEIVED_USER, user: defaultState.users.byId[userId] },
      { type: SET_CUSTOM_COLUMNS, value: [] }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(verifyEmailStart());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow processing email verification codes', async () => {
    jest.clearAllMocks();
    const expectedActions = [{ type: RECEIVED_ACTIVATION_CODE, code: 'code' }];
    const store = mockStore({ ...defaultState });
    await store.dispatch(setAccountActivationCode('code'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow completing email verification', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: RECEIVED_USER, user: defaultState.users.byId[userId] },
      { type: SET_CUSTOM_COLUMNS, value: [] }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(verifyEmailComplete('superSecret'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    const result = store.dispatch(verifyEmailComplete('ohNo'));
    expect(result).rejects.toBeTruthy();
  });
  it('should allow logging in', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: RECEIVED_USER, user: defaultState.users.byId[userId] },
      { type: SET_CUSTOM_COLUMNS, value: [] },
      { type: SUCCESSFULLY_LOGGED_IN, value: { token } }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(loginUser({ email: 'test@example.com', password: defaultPassword }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should redirect on SSO login', async () => {
    const store = mockStore({ ...defaultState });
    const replaceSpy = jest.spyOn(window.location, 'replace');
    await store.dispatch(loginUser({ email: 'test@example.com' }));
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
    expect(replaceSpy).toHaveBeenCalledWith(getSamlStartUrl(testSsoId));
  });
  it('should prevent logging in with a limited user', async () => {
    jest.clearAllMocks();
    window.localStorage.getItem.mockReturnValue(JSON.stringify({ token: 'limitedToken' }));
    const expectedActions = [{ type: SET_SNACKBAR, snackbar: { message: 'forbidden by role-based access control' } }];
    const store = mockStore({ ...defaultState });
    try {
      await store.dispatch(loginUser({ email: 'test-limited@example.com', password: defaultPassword }));
    } catch (error) {
      expect(error).toMatchObject(expectedActions[0]);
    }
    expect(window.localStorage.removeItem).toHaveBeenCalledWith('JWT');
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    window.localStorage.getItem.mockReset();
  });
  it('should allow logging out', async () => {
    jest.clearAllMocks();
    const expectedActions = [{ type: USER_LOGOUT }];
    const store = mockStore({ ...defaultState });
    await store.dispatch(logoutUser());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should not allow logging out with an active upload', async () => {
    jest.clearAllMocks();
    const store = mockStore({ ...defaultState, releases: { ...defaultState.releases, uploadProgress: 42 } });
    await store.dispatch(logoutUser()).catch(() => expect(true).toEqual(true));
  });
  it('should allow switching users', async () => {
    jest.clearAllMocks();
    const reloadSpy = jest.spyOn(window.location, 'reload');
    const store = mockStore({ ...defaultState, releases: { ...defaultState.releases, uploadProgress: 42 } });
    await store.dispatch(switchUserOrganization('a1'));
    expect(localStorage.getItem).toHaveBeenCalledWith('JWT');
    expect(localStorage.setItem).toHaveBeenCalledWith('JWT', JSON.stringify({ token: 'differentToken' }));
    expect(reloadSpy).toHaveBeenCalled();
  });
  it('should not allow switching users during uploads', async () => {
    jest.clearAllMocks();
    const store = mockStore({ ...defaultState, releases: { ...defaultState.releases, uploadProgress: 42 } });
    await store.dispatch(switchUserOrganization('a1')).catch(() => expect(true).toEqual(true));
  });
  it('should allow single user retrieval', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: RECEIVED_USER, user: defaultState.users.byId.a1 },
      { type: SET_CUSTOM_COLUMNS } //, value: [] } <= we can't check for the correct value here as the localstorage is (ab)used by msw to track state during req/res cycles, thus the localStorage expectation further down
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(getUser('a1'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expect(window.localStorage.getItem).toHaveBeenCalledWith(`a1-column-widths`);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow current user initialization', async () => {
    jest.clearAllMocks();
    const expectedActions = appInitActions;
    const store = mockStore({ ...defaultState });
    await store.dispatch(initializeSelf());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow user list retrieval', async () => {
    jest.clearAllMocks();
    const expectedActions = [{ type: RECEIVED_USER_LIST, users: defaultState.users.byId }];
    const store = mockStore({ ...defaultState });
    await store.dispatch(getUserList());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow single user creation', async () => {
    jest.clearAllMocks();
    const createdUser = { email: 'a@b.com', password: defaultPassword };
    const expectedActions = [
      { type: CREATED_USER, user: createdUser },
      { type: SET_SNACKBAR, snackbar: { message: 'The user was created successfully.' } },
      { type: RECEIVED_USER_LIST, users: defaultState.users.byId }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(createUser(createdUser));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow single user edits', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: UPDATED_USER, userId: 'a1', user: { password: defaultPassword } },
      { type: SET_SNACKBAR, snackbar: { message: 'The user has been updated.' } }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(editUser('a1', { email: defaultState.users.byId.a1.email, password: defaultPassword }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should not allow current user edits without proper password', async () => {
    jest.clearAllMocks();
    const store = mockStore({ ...defaultState });
    const result = store.dispatch(editUser('a1', { email: 'a@evil.com', password: 'mySecretPasswordNot' }));
    expect(result).rejects.toBeTruthy();
  });
  it('should allow single user removal', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: REMOVED_USER, userId: 'a1' },
      { type: SET_SNACKBAR, snackbar: { message: 'The user was removed from the system.' } },
      { type: RECEIVED_USER_LIST, users: defaultState.users.byId }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(removeUser('a1'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow single user removal', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: SET_SNACKBAR, snackbar: { message: 'The user was added successfully.' } },
      { type: RECEIVED_USER_LIST, users: defaultState.users.byId }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(addUserToCurrentTenant('a1'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should allow role list retrieval', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: RECEIVED_PERMISSION_SETS, value: receivedPermissionSets },
      { type: RECEIVED_ROLES, value: receivedRoles }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(getRoles());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow role creation', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      { type: CREATED_ROLE, role: defaultRole, roleId: defaultRole.name },
      { type: SET_SNACKBAR, snackbar: { message: 'The role was created successfully.' } },
      { type: RECEIVED_PERMISSION_SETS, value: receivedPermissionSets },
      { type: RECEIVED_ROLES, value: receivedRoles }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(createRole({ ...defaultRole, uiPermissions: { groups: [{ item: 'testGroup', uiPermissions: [uiPermissionsById.manage.value] }] } }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow role edits', async () => {
    jest.clearAllMocks();
    const expectedActions = [
      {
        type: UPDATED_ROLE,
        roleId: defaultRole.name,
        role: {
          ...defaultRole,
          uiPermissions: {
            ...defaultRole.uiPermissions,
            groups: { ...defaultRole.uiPermissions.groups, testGroup: [uiPermissionsById.manage.value] }
          }
        }
      },
      { type: SET_SNACKBAR, snackbar: { message: 'The role has been updated.' } },
      { type: RECEIVED_PERMISSION_SETS, value: receivedPermissionSets },
      { type: RECEIVED_ROLES, value: receivedRoles }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(
      editRole({ name: defaultRole.name, uiPermissions: { groups: [{ item: 'testGroup', uiPermissions: [uiPermissionsById.manage.value] }] } })
    );
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow role removal', async () => {
    jest.clearAllMocks();
    // eslint-disable-next-line no-unused-vars
    const { test, ...remainder } = defaultState.users.rolesById;
    const expectedActions = [
      { type: REMOVED_ROLE, value: remainder },
      { type: SET_SNACKBAR, snackbar: { message: 'The role was deleted successfully.' } },
      { type: RECEIVED_PERMISSION_SETS, value: receivedPermissionSets },
      { type: RECEIVED_ROLES, value: receivedRoles }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(removeRole('test'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow password reset - pt. 1', async () => {
    const store = mockStore({ ...defaultState });
    store.dispatch(passwordResetStart(defaultState.users.byId.a1.email)).then(() => expect(true).toEqual(true));
  });
  it('should allow password reset - pt. 2', async () => {
    const store = mockStore({ ...defaultState });
    store.dispatch(passwordResetComplete('secretHash', 'newPassword')).then(() => expect(true).toEqual(true));
  });
  it('should allow storing global settings without deletion', async () => {
    jest.clearAllMocks();
    // eslint-disable-next-line no-unused-vars
    const { id_attribute, ...retrievedSettings } = defaultState.users.globalSettings;
    const expectedActions = [
      { type: SET_GLOBAL_SETTINGS, settings: { ...retrievedSettings } },
      offlineThreshold,
      { type: SET_GLOBAL_SETTINGS, settings: { ...defaultState.users.globalSettings, ...settings } }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(saveGlobalSettings(settings));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow storing global settings without deletion and with notification', async () => {
    jest.clearAllMocks();
    // eslint-disable-next-line no-unused-vars
    const { id_attribute, ...retrievedSettings } = defaultState.users.globalSettings;
    const expectedActions = [
      { type: SET_GLOBAL_SETTINGS, settings: { ...retrievedSettings } },
      offlineThreshold,
      { type: SET_GLOBAL_SETTINGS, settings: { ...defaultState.users.globalSettings, ...settings } },
      { type: SET_SNACKBAR, snackbar: { message: 'Settings saved successfully' } }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(saveGlobalSettings(settings, false, true));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow storing user scoped settings', async () => {
    jest.clearAllMocks();
    // eslint-disable-next-line no-unused-vars
    const { ...settings } = defaultState.users.userSettings;
    const expectedActions = [
      { type: SET_USER_SETTINGS, settings },
      {
        type: SET_USER_SETTINGS,
        settings: { ...settings, extra: 'this' }
      }
    ];
    const store = mockStore({ ...defaultState });
    await store.dispatch(saveUserSettings({ extra: 'this' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should store the visibility of the announcement shown in the header in a cookie on dismissal', async () => {
    jest.clearAllMocks();
    const cookies = new Cookies();
    const expectedActions = [{ type: SET_ANNOUNCEMENT, announcement: undefined }];
    const store = mockStore({ ...defaultState, app: { ...defaultState.app, hostedAnnouncement: 'something' } });
    await store.dispatch(setHideAnnouncement(true));
    const storeActions = store.getActions();
    expect(cookies.get).toHaveBeenCalledTimes(1);
    expect(cookies.set).toHaveBeenCalledTimes(1);
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should store the sizes of columns in local storage', async () => {
    jest.clearAllMocks();
    const expectedActions = [{ type: SET_CUSTOM_COLUMNS, value: [{ asd: 'asd' }] }];
    const store = mockStore({ ...defaultState, users: { ...defaultState.users, customColumns: [{ asd: 'asd' }] } });
    await store.dispatch(updateUserColumnSettings([{ asd: 'asd' }]));
    const storeActions = store.getActions();
    expect(localStorage.getItem).not.toHaveBeenCalled();
    expect(localStorage.setItem).toHaveBeenCalled();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));

    jest.clearAllMocks();
    await store.dispatch(updateUserColumnSettings());
    expect(localStorage.getItem).toHaveBeenCalledTimes(1);
    expect(localStorage.setItem).toHaveBeenCalledTimes(1);
  });

  it('should allow token list retrieval', async () => {
    jest.clearAllMocks();
    const expectedActions = [{ type: UPDATED_USER, userId: 'a1', user: { tokens: accessTokens } }];
    const store = mockStore({ ...defaultState });
    await store.dispatch(getTokens());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow token generation', async () => {
    jest.clearAllMocks();
    const expectedActions = [{ type: UPDATED_USER, userId: 'a1', user: { tokens: accessTokens } }];
    const store = mockStore({ ...defaultState });
    const result = await store.dispatch(generateToken({ name: 'name' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expect(result[result.length - 1]).toEqual('aNewToken');
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow token removal', async () => {
    jest.clearAllMocks();
    // eslint-disable-next-line no-unused-vars
    const expectedActions = [{ type: UPDATED_USER, userId: 'a1', user: { tokens: accessTokens } }];
    const store = mockStore({ ...defaultState });
    await store.dispatch(revokeToken({ id: 'some-id-1' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should handle setting single tooltip read state', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: SET_TOOLTIP_STATE, id: 'foo', value: { readState: 'testRead' } },
      { type: SET_USER_SETTINGS, settings: { ...defaultState.users.userSettings } },
      { type: SET_USER_SETTINGS, settings: { ...defaultState.users.userSettings } }
    ];
    await store.dispatch(setTooltipReadState('foo', 'testRead', true));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should handle setting tooltip read state for all tips', async () => {
    const store = mockStore({ ...defaultState });

    const expectedActions = [
      { type: SET_TOOLTIPS_STATE, value: { ...Object.values(HELPTOOLTIPS).reduce((accu, { id }) => ({ ...accu, [id]: { readState: 'testRead' } }), {}) } },
      { type: SET_USER_SETTINGS, settings: { ...defaultState.users.userSettings } },
      { type: SET_USER_SETTINGS, settings: { ...defaultState.users.userSettings } }
    ];
    await store.dispatch(setAllTooltipsReadState('testRead'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});
