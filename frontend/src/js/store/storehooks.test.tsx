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
import { Provider } from 'react-redux';

import {
  getDeploymentsByStatus,
  getDeviceAttributes,
  getDeviceLimit,
  getDevicesByStatus,
  getDevicesWithAuth,
  getDynamicGroups,
  getGroups,
  getIntegrations,
  getReleases,
  getUserOrganization,
  tenantDataDivergedMessage
} from '@northern.tech/store/thunks';
import { act, renderHook } from '@testing-library/react';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { inventoryDevice } from '../../../tests/__mocks__/deviceHandlers';
import { defaultState, receivedPermissionSets, receivedRoles, userId } from '../../../tests/mockData';
import { actions as appActions } from './appSlice';
import { getLatestReleaseInfo, setOfflineThreshold } from './appSlice/thunks';
import { latestSaasReleaseTag } from './appSlice/thunks.test';
import { getSessionInfo } from './auth';
import { EXTERNAL_PROVIDER, TIMEOUTS, UNGROUPED_GROUP, timeUnits } from './commonConstants';
import { DEVICE_STATES } from './constants';
import { actions as deploymentsActions } from './deploymentsSlice';
import { actions as deviceActions } from './devicesSlice';
import { actions as onboardingActions } from './onboardingSlice';
import { defaultOnboardingState, expectedOnboardingActions } from './onboardingSlice/thunks.test';
import { actions as organizationActions } from './organizationSlice';
import { actions as releasesActions } from './releasesSlice';
import { useAppInit } from './storehooks';
import { actions as userActions } from './usersSlice';
import { getGlobalSettings, getPermissionSets, getRoles, getUserSettings, saveUserSettings } from './usersSlice/thunks';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

const attributeReducer = (accu, item) => {
  if (item.scope === 'inventory') {
    accu[item.name] = item.value;
    if (item.name === 'device_type') {
      accu[item.name] = [].concat(item.value);
    }
  }
  return accu;
};

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const { attributes, ...expectedDevice } = defaultState.devices.byId.a1;
export const receivedInventoryDevice = {
  ...defaultState.devices.byId.a1,
  attributes: inventoryDevice.attributes.reduce(attributeReducer, {}),
  identity_data: { ...defaultState.devices.byId.a1.identity_data, status: DEVICE_STATES.accepted },
  isNew: false,
  isOffline: true,
  monitor: {},
  tags: {},
  updated_ts: inventoryDevice.updated_ts
};

const appInitActions = [
  { type: userActions.successfullyLoggedIn.type }, //, payload: { token }
  { type: onboardingActions.setOnboardingComplete.type, payload: false },
  { type: onboardingActions.setDemoArtifactPort.type, payload: 85 },
  { type: appActions.setFeatures.type, payload: { ...defaultState.app.features, hasMultitenancy: true } },
  {
    type: appActions.setVersionInformation.type,
    payload: {
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
    }
  },
  {
    type: appActions.setEnvironmentData.type,
    payload: { feedbackProbability: 0.3, hostAddress: null, hostedAnnouncement: '', recaptchaSiteKey: '', stripeAPIKey: '', trackerCode: '' }
  },
  { type: getLatestReleaseInfo.pending.type },
  { type: getUserSettings.pending.type },
  { type: getGlobalSettings.pending.type },
  { type: getDeviceAttributes.pending.type },
  { type: getDeploymentsByStatus.pending.type },
  { type: getDeploymentsByStatus.pending.type },
  { type: getDevicesByStatus.pending.type },
  { type: getDevicesByStatus.pending.type },
  { type: getDevicesByStatus.pending.type },
  { type: getDevicesByStatus.pending.type },
  { type: getDynamicGroups.pending.type },
  { type: getGroups.pending.type },
  { type: getIntegrations.pending.type },
  { type: getReleases.pending.type },
  { type: getDeviceLimit.pending.type },
  { type: getRoles.pending.type },
  { type: getPermissionSets.pending.type },
  { type: appActions.setFirstLoginAfterSignup.type, payload: false },
  { type: getUserOrganization.pending.type },
  { type: deploymentsActions.receivedDeployments.type, payload: defaultState.deployments.byId },
  {
    type: deploymentsActions.receivedDeploymentsForStatus.type,
    payload: { deploymentIds: Object.keys(defaultState.deployments.byId), status: 'finished', total: Object.keys(defaultState.deployments.byId).length }
  },
  { type: deploymentsActions.receivedDeployments.type, payload: defaultState.deployments.byId },
  {
    type: deploymentsActions.receivedDeploymentsForStatus.type,
    payload: { deploymentIds: Object.keys(defaultState.deployments.byId), status: 'inprogress', total: Object.keys(defaultState.deployments.byId).length }
  },
  {
    type: deploymentsActions.selectDeploymentsForStatus.type,
    payload: { deploymentIds: Object.keys(defaultState.deployments.byId), status: 'inprogress', total: Object.keys(defaultState.deployments.byId).length }
  },
  { type: getDeploymentsByStatus.fulfilled.type },
  { type: getDeploymentsByStatus.fulfilled.type },
  { type: deviceActions.setDeviceLimit.type, payload: 500 },
  { type: getDeviceLimit.fulfilled.type },
  {
    type: deviceActions.receivedGroups.type,
    payload: {
      testGroup: defaultState.devices.groups.byId.testGroup,
      testGroupDynamic: { filters: [{ key: 'group', operator: '$eq', scope: 'system', value: 'things' }], id: 'filter1' }
    }
  },
  { type: getDevicesByStatus.pending.type },
  { type: deviceActions.setFilterAttributes.type },
  { type: getDeviceAttributes.fulfilled.type },
  {
    type: deviceActions.receivedGroups.type,
    payload: {
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
  { type: getDynamicGroups.fulfilled.type },
  {
    type: deviceActions.receivedDevices.type,
    payload: {
      [defaultState.devices.byId.a1.id]: { ...receivedInventoryDevice, group: 'test' },
      [defaultState.devices.byId.b1.id]: {
        ...receivedInventoryDevice,
        id: defaultState.devices.byId.b1.id,
        group: 'test',
        identity_data: { ...defaultState.devices.byId.b1.identity_data, status: DEVICE_STATES.accepted }
      }
    }
  },
  {
    type: deviceActions.setDevicesByStatus.type,
    payload: {
      deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id],
      status: DEVICE_STATES.accepted,
      total: defaultState.devices.byStatus.accepted.deviceIds.length
    }
  },
  { type: getDevicesWithAuth.pending.type },
  { type: deviceActions.receivedDevices.type, payload: { [expectedDevice.id]: { ...receivedInventoryDevice, group: 'test', status: 'pending' } } },
  {
    type: deviceActions.setDevicesByStatus.type,
    payload: {
      deviceIds: Array.from({ length: defaultState.devices.byStatus.pending.total }, () => defaultState.devices.byId.a1.id),
      status: DEVICE_STATES.pending,
      total: defaultState.devices.byStatus.pending.deviceIds.length
    }
  },
  { type: getDevicesWithAuth.pending.type },
  { type: deviceActions.receivedDevices.type, payload: {} },
  { type: deviceActions.setDevicesByStatus.type, payload: { deviceIds: [], status: 'preauthorized', total: 0 } },
  { type: deviceActions.receivedDevices.type, payload: {} },
  { type: deviceActions.setDevicesByStatus.type, payload: { deviceIds: [], status: 'rejected', total: 0 } },
  {
    type: appActions.setVersionInformation.type,
    payload: {
      GUI: latestSaasReleaseTag,
      Integration: '1.2.3',
      'Mender-Artifact': '1.3.7',
      'Mender-Client': '3.2.1',
      backend: latestSaasReleaseTag,
      latestRelease: {
        releaseDate: '2022-02-02',
        repos: {
          integration: '1.2.3',
          mender: '3.2.1',
          'mender-artifact': '1.3.7',
          'other-service': '1.1.0',
          service: '3.0.0'
        }
      }
    }
  },
  { type: organizationActions.setOrganization.type, payload: defaultState.organization.organization },
  { type: appActions.setAnnouncement.type, payload: tenantDataDivergedMessage },
  { type: getDevicesByStatus.fulfilled.type },
  { type: getDevicesByStatus.fulfilled.type },
  { type: getLatestReleaseInfo.fulfilled.type },
  { type: getUserOrganization.fulfilled.type },
  {
    type: organizationActions.receiveExternalDeviceIntegrations.type,
    payload: [
      { connection_string: 'something_else', id: 1, provider: EXTERNAL_PROVIDER['iot-hub'].provider },
      { id: 2, provider: EXTERNAL_PROVIDER['iot-core'].provider, something: 'new' }
    ]
  },
  { type: getIntegrations.fulfilled.type },
  { type: releasesActions.receiveReleases.type, payload: defaultState.releases.byId },
  {
    type: releasesActions.setReleaseListState.type,
    payload: { ...defaultState.releases.releasesList, releaseIds: [defaultState.releases.byId.r1.name], page: 42 }
  },
  { type: getReleases.fulfilled.type },
  {
    type: deviceActions.receivedDevices.type,
    payload: {
      [expectedDevice.id]: { ...defaultState.devices.byId.a1, group: undefined, isNew: false, isOffline: true, monitor: {}, tags: {} },
      [defaultState.devices.byId.b1.id]: { ...defaultState.devices.byId.b1, group: undefined, isNew: false, isOffline: true, monitor: {}, tags: {} }
    }
  },
  {
    type: deviceActions.receivedDevices.type,
    payload: {
      [expectedDevice.id]: { ...defaultState.devices.byId.a1, group: undefined, isNew: false, isOffline: true, monitor: {}, tags: {} }
    }
  },
  { type: getDevicesWithAuth.fulfilled.type },
  { type: getDevicesWithAuth.fulfilled.type },
  {
    type: deviceActions.receivedDevices.type,
    payload: {
      [expectedDevice.id]: { ...receivedInventoryDevice, group: 'test' },
      [defaultState.devices.byId.b1.id]: { ...receivedInventoryDevice, id: defaultState.devices.byId.b1.id, group: 'test' }
    }
  },
  { type: getDevicesWithAuth.pending.type },
  { type: getDevicesByStatus.fulfilled.type },
  { type: getDevicesByStatus.fulfilled.type },
  { type: userActions.setGlobalSettings.type, payload: { ...defaultState.users.globalSettings } },
  { type: setOfflineThreshold.pending.type },
  { type: appActions.setOfflineThreshold.type, payload: '2019-01-12T13:00:06.900Z' },
  { type: setOfflineThreshold.fulfilled.type },
  { type: userActions.setUserSettings.type, payload: { ...defaultState.users.userSettings } },
  { type: getGlobalSettings.fulfilled.type },
  { type: getUserSettings.fulfilled.type },
  { type: userActions.receivedPermissionSets.type, payload: receivedPermissionSets },
  { type: getPermissionSets.fulfilled.type },
  { type: userActions.receivedRoles.type, payload: receivedRoles },
  { type: getRoles.fulfilled.type },
  {
    type: deviceActions.receivedDevices.type,
    payload: {
      [defaultState.devices.byId.a1.id]: { ...defaultState.devices.byId.a1, group: undefined, isNew: false, isOffline: true, monitor: {}, tags: {} },
      [defaultState.devices.byId.b1.id]: { ...defaultState.devices.byId.b1, group: undefined, isNew: false, isOffline: true, monitor: {}, tags: {} }
    }
  },
  { type: getDevicesWithAuth.fulfilled.type },
  { type: getDevicesByStatus.fulfilled.type },
  {
    type: deviceActions.addGroup.type,
    payload: {
      groupName: UNGROUPED_GROUP.id,
      group: {
        filters: [{ key: 'group', operator: '$nin', scope: 'system', value: [Object.keys(defaultState.devices.groups.byId)[0]] }]
      }
    }
  },
  { type: getGroups.fulfilled.type },
  { type: deviceActions.setDeviceListState.type, payload: { selectedAttributes: [] } },
  { type: userActions.setTooltipsState.type, payload: {} },
  { type: saveUserSettings.pending.type },
  { type: getUserSettings.pending.type },
  { type: userActions.setUserSettings.type, payload: { ...defaultState.users.userSettings } },
  { type: getUserSettings.fulfilled.type },
  { type: userActions.setUserSettings.type, payload: { ...defaultState.users.userSettings, onboarding: defaultOnboardingState } },
  { type: saveUserSettings.fulfilled.type },
  ...expectedOnboardingActions
];

it('should try to get all required app information', async () => {
  const store = mockStore({
    ...defaultState,
    app: { ...defaultState.app, features: { ...defaultState.app.features, isHosted: true } },
    users: {
      ...defaultState.users,
      currentSession: getSessionInfo(),
      globalSettings: { ...defaultState.users.globalSettings, id_attribute: { attribute: 'mac', scope: 'identity' } }
    },
    releases: { ...defaultState.releases, releasesList: { ...defaultState.releases.releasesList, page: 42 } }
  });
  const wrapper = ({ children }) => <Provider store={store}>{children}</Provider>;
  renderHook(() => useAppInit(userId), { wrapper });
  await act(async () => {
    jest.runAllTimers();
    jest.runAllTicks();
  });

  const storeActions = store.getActions();
  expect(storeActions.length).toEqual(appInitActions.length);
  appInitActions.map((action, index) => Object.keys(action).map(key => expect(storeActions[index][key]).toEqual(action[key])));
});
it('should execute the offline threshold migration for multi day thresholds', async () => {
  const store = mockStore({
    ...defaultState,
    app: { ...defaultState.app, features: { ...defaultState.app.features, isHosted: true } },
    users: {
      ...defaultState.users,
      currentSession: getSessionInfo(),
      globalSettings: {
        ...defaultState.users.globalSettings,
        id_attribute: { attribute: 'mac', scope: 'identity' },
        offlineThreshold: { interval: 48, intervalUnit: timeUnits.hours }
      }
    },
    releases: { ...defaultState.releases, releasesList: { ...defaultState.releases.releasesList, page: 42 } }
  });
  const wrapper = ({ children }) => <Provider store={store}>{children}</Provider>;
  renderHook(() => useAppInit(userId), { wrapper });
  await act(async () => {
    jest.runAllTimers();
    jest.runAllTicks();
  });

  const storeActions = store.getActions();
  expect(storeActions.length).toEqual(appInitActions.length + 9); // 3 = get settings + set settings + set offline threshold
  const settingStorageAction = storeActions.find(action => action.type === userActions.setGlobalSettings.type && action.payload.offlineThreshold);
  expect(settingStorageAction.payload.offlineThreshold.interval).toEqual(2);
  expect(settingStorageAction.payload.offlineThreshold.intervalUnit).toEqual(timeUnits.days);
});
it('should trigger the offline threshold migration dialog', async () => {
  const store = mockStore({
    ...defaultState,
    app: { ...defaultState.app, features: { ...defaultState.app.features, isHosted: true } },
    users: {
      ...defaultState.users,
      currentSession: getSessionInfo(),
      globalSettings: {
        ...defaultState.users.globalSettings,
        id_attribute: { attribute: 'mac', scope: 'identity' },
        offlineThreshold: { interval: 15, intervalUnit: 'minutes' }
      }
    },
    releases: { ...defaultState.releases, releasesList: { ...defaultState.releases.releasesList, page: 42 } }
  });

  const wrapper = ({ children }) => <Provider store={store}>{children}</Provider>;
  renderHook(() => useAppInit(userId), { wrapper });
  await jest.advanceTimersByTimeAsync(TIMEOUTS.fiveSeconds + TIMEOUTS.oneSecond);
  await jest.runAllTimersAsync();
  await act(async () => {
    jest.runAllTicks();
  });
  const storeActions = store.getActions();
  expect(storeActions.length).toEqual(appInitActions.length + 1); // only setShowStartupNotification should be addded
  const notificationAction = storeActions.find(action => action.type === userActions.setShowStartupNotification.type);
  expect(notificationAction.payload).toBeTruthy();
});
