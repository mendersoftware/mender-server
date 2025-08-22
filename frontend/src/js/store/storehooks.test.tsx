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
import { Provider } from 'react-redux';

import { getUserOrganization, tenantDataDivergedMessage } from '@northern.tech/store/thunks';
import { deepCompare } from '@northern.tech/utils/helpers';
import { renderHook, waitFor } from '@testing-library/react';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';
import { vi } from 'vitest';

import { inventoryDevice } from '../../../tests/__mocks__/deviceHandlers';
import { defaultState, userId } from '../../../tests/mockData';
import { actions as appActions } from './appSlice';
import { latestSaasReleaseTag } from './appSlice/thunks.test';
import { getSessionInfo } from './auth';
import { EXTERNAL_PROVIDER, timeUnits } from './commonConstants';
import { DEVICE_STATES } from './constants';
import { expectedOnboardingActions } from './onboardingSlice/thunks.test';
import { actions as organizationActions } from './organizationSlice';
import { useAppInit } from './storehooks';
import { actions as userActions } from './usersSlice';

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
  { type: userActions.successfullyLoggedIn.type },
  { type: appActions.setFeatures.type, payload: { ...defaultState.app.features, hasMultitenancy: true, isHosted: false } },
  {
    type: appActions.setVersionInformation.type,
    payload: {
      docsVersion: '',
      Integration: 'master',
      'Mender-Artifact': undefined,
      'Mender-Client': 'next',
      'Meta-Mender': 'saas-123.34'
    }
  },
  {
    type: appActions.setEnvironmentData.type,
    payload: {
      commit: '',
      feedbackProbability: 0.3,
      hostAddress: null,
      hostedAnnouncement: '',
      recaptchaSiteKey: '',
      sentry: { location: '', replaysSessionSampleRate: 0.1, tracesSampleRate: 1 },
      stripeAPIKey: '',
      trackerCode: ''
    }
  },
  { type: appActions.setFirstLoginAfterSignup.type, payload: false },
  { type: getUserOrganization.pending.type },
  {
    type: appActions.setVersionInformation.type,
    payload: {
      Integration: '1.2.3',
      'Mender-Artifact': '1.3.7',
      'Mender-Client': '3.2.1',
      Server: latestSaasReleaseTag,
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
  { type: appActions.setOfflineThreshold.type, payload: '2019-01-12T13:00:00.950Z' },
  {
    type: organizationActions.receiveExternalDeviceIntegrations.type,
    payload: [
      { connection_string: 'something_else', id: 1, provider: EXTERNAL_PROVIDER['iot-hub'].provider },
      { id: 2, provider: EXTERNAL_PROVIDER['iot-core'].provider, something: 'new' }
    ]
  },
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
  const { result } = renderHook(() => useAppInit(userId), { wrapper });
  await waitFor(() => expect(result.current.coreInitDone).toBeTruthy());
  await vi.runAllTimersAsync();
  const storeActions = store.getActions();
  appInitActions.forEach(initAction => {
    const handledAction = storeActions.some(storeAction => Object.keys(initAction).every(key => deepCompare(storeAction[key], initAction[key])));
    expect(handledAction).toBeTruthy();
  });
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
  const { result } = renderHook(() => useAppInit(userId), { wrapper });
  await waitFor(() => expect(result.current.coreInitDone).toBeTruthy());
  await vi.runAllTimersAsync();

  const storeActions = store.getActions();
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
  const { result } = renderHook(() => useAppInit(userId), { wrapper });
  await waitFor(() => expect(result.current.coreInitDone).toBeTruthy());
  await vi.runAllTimersAsync();
  const storeActions = store.getActions();
  const notificationAction = storeActions.find(action => action.type === userActions.setShowStartupNotification.type);
  expect(notificationAction.payload).toBeTruthy();
});
