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

import { defaultState, server } from '@/testUtils';
import { actions as appActions } from '@northern.tech/store/appSlice';
import { getSessionInfo } from '@northern.tech/store/auth';
import { EXTERNAL_PROVIDER, timeUnits } from '@northern.tech/store/commonConstants';
import { DEVICE_STATES, locations } from '@northern.tech/store/constants';
import { actions as onboardingActions } from '@northern.tech/store/onboardingSlice';
import { actions as organizationActions } from '@northern.tech/store/organizationSlice';
import { getOnboardingState, getUserOrganization, getUserSettings, saveUserSettings } from '@northern.tech/store/thunks';
import { actions as userActions } from '@northern.tech/store/usersSlice';
import { userId } from '@northern.tech/testing/mockData';
import { inventoryDevice } from '@northern.tech/testing/requestHandlers/deviceHandlers';
import { tenantadmApiUrlv2 } from '@northern.tech/utils/constants';
import { deepCompare } from '@northern.tech/utils/helpers';
import { renderHook, waitFor } from '@testing-library/react';
import { HttpResponse, http } from 'msw';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';
import { expect, it, vi } from 'vitest';

import { parseEnvironmentInfo, useAppInit } from './appInit';

const oldHostname = window.location.hostname;

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
  isNew: true,
  isOffline: true,
  monitor: {},
  tags: {},
  updated_ts: inventoryDevice.updated_ts
};

const defaultOnboardingState = {
  approach: null,
  complete: false,
  demoArtifactPort: 85,
  deviceType: null,
  progress: undefined,
  showConnectDeviceDialog: false,
  showTips: undefined,
  showTipsDialog: false,
  something: 'here'
};

const expectedOnboardingActions = [
  { type: getOnboardingState.pending.type },
  { type: onboardingActions.setOnboardingComplete.type, payload: false },
  {
    type: onboardingActions.setOnboardingState.type,
    payload: {
      ...defaultOnboardingState,
      address: 'http://192.168.10.141:85',
      approach: 'physical',
      deviceType: ['raspberrypi4'],
      progress: 'devices-accepted-onboarding',
      showTips: true
    }
  },
  { type: saveUserSettings.pending.type },
  { type: getUserSettings.pending.type },
  { type: userActions.setUserSettings.type, payload: { ...defaultState.users.userSettings } },
  { type: getUserSettings.fulfilled.type },
  {
    type: userActions.setUserSettings.type,
    payload: {
      ...defaultState.users.userSettings,
      firstLoginTimestamp: '2019-01-13T13:00:00+00:00',
      onboarding: {
        ...defaultOnboardingState,
        address: 'http://192.168.10.141:85',
        approach: 'physical',
        deviceType: ['raspberrypi4'],
        progress: 'devices-accepted-onboarding',
        showTips: true
      }
    }
  },
  { type: saveUserSettings.fulfilled.type },
  { type: getOnboardingState.fulfilled.type }
];

const appInitActions = [
  { type: userActions.successfullyLoggedIn.type },
  { type: appActions.setFeatures.type, payload: { ...defaultState.app.features, hasMultitenancy: true, isHosted: false } },
  { type: appActions.setVersionInformation.type, payload: { docsVersion: '', version: 'saas-123.34' } },
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
  { type: appActions.setVersionInformation.type, payload: { docsVersion: '', version: 'saas-123.34' } },
  { type: organizationActions.setOrganization.type, payload: defaultState.organization.organization },
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
it('should treat hosted domains and their subdomains as hosted, except per-PR preview deployments', async () => {
  const expectations = [
    { hostname: locations.us.location, isHosted: true },
    { hostname: locations.eu.location, isHosted: true },
    { hostname: locations.cn.location, isHosted: true },
    // staging and other subdomains of a hosted domain are valid hosted instances
    { hostname: `staging.${locations.us.location}`, isHosted: true },
    { hostname: `testing.staging.${locations.us.location}`, isHosted: true },
    // per-PR preview deployments (leading <component>-pr-<number> label) must not be treated as hosted
    { hostname: `os-pr-2012.staging.${locations.us.location}`, isHosted: false },
    { hostname: 'localhost', isHosted: false }
  ];
  for (const { hostname, isHosted } of expectations) {
    window.location.hostname = hostname;
    const store = mockStore({ ...defaultState, app: { ...defaultState.app, features: { ...defaultState.app.features, isHosted: false } } });
    await store.dispatch(parseEnvironmentInfo());
    const setFeaturesAction = store.getActions().find(action => action.type === appActions.setFeatures.type);
    expect(setFeaturesAction.payload.isHosted).toBe(isHosted);
  }
  window.location.hostname = oldHostname;
});

const makeCoreInitStore = overrides =>
  mockStore({
    ...defaultState,
    ...overrides,
    app: { ...defaultState.app, ...overrides.app, features: { ...defaultState.app.features, ...overrides.app?.features } },
    organization: { ...defaultState.organization, organization: {}, ...overrides.organization },
    users: {
      ...defaultState.users,
      currentSession: getSessionInfo(),
      globalSettings: { ...defaultState.users.globalSettings, id_attribute: { attribute: 'mac', scope: 'identity' } }
    }
  });

const onPremEnterpriseFeatures = { hasMultitenancy: true, isEnterprise: true };
const osFeatures = { hasMultitenancy: false, isEnterprise: false };
const setFeaturesActionsFor = (storeActions, payload) =>
  storeActions.filter(action => action.type === appActions.setFeatures.type && deepCompare(action.payload, payload));

it('should mark a non-hosted deployment with a reachable organization endpoint as multitenant enterprise', async () => {
  const store = makeCoreInitStore({ app: { features: { isHosted: false } } });
  const wrapper = ({ children }) => <Provider store={store}>{children}</Provider>;
  const { result } = renderHook(() => useAppInit(userId), { wrapper });
  await waitFor(() => expect(result.current.coreInitDone).toBeTruthy());
  await vi.runAllTimersAsync();
  const storeActions = store.getActions();
  expect(storeActions.some(action => action.type === getUserOrganization.fulfilled.type)).toBeTruthy();
  // a reachable org endpoint on a non-hosted deployment identifies an on-prem enterprise install
  expect(setFeaturesActionsFor(storeActions, onPremEnterpriseFeatures)).toHaveLength(1);
});
it('should mark a non-hosted deployment as OS when the organization endpoint is unreachable', async () => {
  server.use(http.get(`${tenantadmApiUrlv2}/tenants/me`, () => new HttpResponse(null, { status: 500 })));
  const store = makeCoreInitStore({ app: { features: { isHosted: false } } });
  const wrapper = ({ children }) => <Provider store={store}>{children}</Provider>;
  const { result } = renderHook(() => useAppInit(userId), { wrapper });
  await waitFor(() => expect(result.current.coreInitDone).toBeTruthy());
  await vi.runAllTimersAsync();
  const storeActions = store.getActions();
  // a failed probe must be caught so core init still completes, and the deployment is marked non-enterprise/OS
  expect(storeActions.some(action => action.type === getUserOrganization.rejected.type)).toBeTruthy();
  expect(storeActions.some(action => action.type === organizationActions.setOrganization.type)).toBeFalsy();
  expect(setFeaturesActionsFor(storeActions, osFeatures)).toHaveLength(1);
});
it('should keep hosted deployments enterprise/multitenant without overriding them from the probe', async () => {
  window.location.hostname = locations.us.location;
  const store = makeCoreInitStore({ app: { features: { isHosted: true } } });
  const wrapper = ({ children }) => <Provider store={store}>{children}</Provider>;
  const { result } = renderHook(() => useAppInit(userId), { wrapper });
  await waitFor(() => expect(result.current.coreInitDone).toBeTruthy());
  await vi.runAllTimersAsync();
  const storeActions = store.getActions();
  const setFeaturesAction = storeActions.find(action => action.type === appActions.setFeatures.type);
  expect(setFeaturesAction.payload.isHosted).toBe(true);
  // hosted enterprise status stays plan-derived, so the probe must not push on-prem/OS overrides
  expect(setFeaturesActionsFor(storeActions, onPremEnterpriseFeatures)).toHaveLength(0);
  expect(setFeaturesActionsFor(storeActions, osFeatures)).toHaveLength(0);
  window.location.hostname = oldHostname;
});
