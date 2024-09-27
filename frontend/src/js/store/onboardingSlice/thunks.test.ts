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
// @ts-nocheck
import { getUserSettings, saveUserSettings } from '@northern.tech/store/thunks';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { actions } from '.';
import { defaultState } from '../../../../tests/mockData';
import { actions as userActions } from '../usersSlice';
import { onboardingSteps } from './constants';
import { advanceOnboarding, getOnboardingState, setOnboardingApproach, setOnboardingCanceled, setOnboardingComplete, setOnboardingDeviceType } from './thunks';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

export const defaultOnboardingState = {
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

export const expectedOnboardingActions = [
  { type: getOnboardingState.pending.type },
  { type: actions.setOnboardingComplete.type, payload: false },
  {
    type: actions.setOnboardingState.type,
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

describe('onboarding actions', () => {
  it('should pass on onboarding completion', async () => {
    const store = mockStore({ ...defaultState });
    await store.dispatch(setOnboardingComplete(true));
    const expectedActions = [
      { type: setOnboardingComplete.pending.type },
      { type: actions.setOnboardingComplete.type, payload: true },
      { type: actions.setShowOnboardingHelp.type, payload: false },
      { type: advanceOnboarding.pending.type },
      { type: actions.setOnboardingProgress.type, payload: onboardingSteps.DEPLOYMENTS_PAST_COMPLETED_FAILURE },
      { type: saveUserSettings.pending.type },
      { type: getUserSettings.pending.type },
      { type: userActions.setUserSettings.type, payload: { ...defaultState.users.userSettings } },
      { type: getUserSettings.fulfilled.type },
      {
        type: userActions.setUserSettings.type,
        payload: {
          ...defaultState.users.userSettings,
          onboarding: {
            ...defaultOnboardingState,
            complete: true,
            progress: onboardingSteps.DEPLOYMENTS_PAST_COMPLETED_FAILURE
          }
        }
      },
      { type: saveUserSettings.fulfilled.type },
      { type: advanceOnboarding.fulfilled.type },
      { type: setOnboardingComplete.fulfilled.type }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => Object.keys(action).map(key => expect(storeActions[index][key]).toEqual(action[key])));
  });
  it('should pass on onboarding approach', async () => {
    const store = mockStore({ ...defaultState });
    await store.dispatch(setOnboardingApproach('test'));
    const expectedActions = [
      { type: setOnboardingApproach.pending.type },
      { type: actions.setOnboardingApproach.type, payload: 'test' },
      { type: saveUserSettings.pending.type },
      { type: getUserSettings.pending.type },
      { type: userActions.setUserSettings.type, payload: { ...defaultState.users.userSettings } },
      { type: getUserSettings.fulfilled.type },
      {
        type: userActions.setUserSettings.type,
        payload: {
          ...defaultState.users.userSettings,
          onboarding: {
            ...defaultOnboardingState,
            approach: 'test'
          }
        }
      },
      { type: saveUserSettings.fulfilled.type },
      { type: setOnboardingApproach.fulfilled.type }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => Object.keys(action).map(key => expect(storeActions[index][key]).toEqual(action[key])));
  });
  it('should pass on onboarding device type', async () => {
    const store = mockStore({ ...defaultState });
    await store.dispatch(setOnboardingDeviceType('testtype'));
    const expectedActions = [
      { type: setOnboardingDeviceType.pending.type },
      { type: actions.setOnboardingDeviceType.type, payload: 'testtype' },
      { type: saveUserSettings.pending.type },
      { type: getUserSettings.pending.type },
      { type: userActions.setUserSettings.type, payload: { ...defaultState.users.userSettings } },
      { type: getUserSettings.fulfilled.type },
      {
        type: userActions.setUserSettings.type,
        payload: {
          ...defaultState.users.userSettings,
          columnSelection: [],
          onboarding: {
            ...defaultOnboardingState,
            deviceType: 'testtype'
          }
        }
      },
      { type: saveUserSettings.fulfilled.type },
      { type: setOnboardingDeviceType.fulfilled.type }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => Object.keys(action).map(key => expect(storeActions[index][key]).toEqual(action[key])));
  });
  it('should advance onboarding by one step', async () => {
    const store = mockStore({ ...defaultState });
    await store.dispatch(advanceOnboarding(onboardingSteps.DASHBOARD_ONBOARDING_START));
    const expectedActions = [
      { type: advanceOnboarding.pending.type },
      { type: actions.setOnboardingProgress.type, payload: onboardingSteps.DEVICES_PENDING_ONBOARDING_START },
      { type: saveUserSettings.pending.type },
      { type: getUserSettings.pending.type },
      { type: userActions.setUserSettings.type, payload: { ...defaultState.users.userSettings } },
      { type: getUserSettings.fulfilled.type },
      {
        type: userActions.setUserSettings.type,
        payload: {
          ...defaultState.users.userSettings,
          columnSelection: [],
          onboarding: {
            ...defaultOnboardingState,
            progress: onboardingSteps.DEVICES_PENDING_ONBOARDING_START
          }
        }
      },
      { type: saveUserSettings.fulfilled.type },
      { type: advanceOnboarding.fulfilled.type }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => Object.keys(action).map(key => expect(storeActions[index][key]).toEqual(action[key])));
  });
  it('should disable helptips and store a canceled state', async () => {
    const store = mockStore({ ...defaultState });
    await store.dispatch(setOnboardingCanceled());
    const expectedActions = [
      { type: setOnboardingCanceled.pending.type },
      { type: actions.setShowOnboardingHelp.type, payload: false },
      { type: actions.setShowDismissOnboardingTipsDialog.type, payload: false },
      { type: actions.setOnboardingComplete.type, payload: true },
      { type: advanceOnboarding.pending.type },
      { type: actions.setOnboardingProgress.type, payload: 'onboarding-canceled' },
      { type: saveUserSettings.pending.type },
      { type: getUserSettings.pending.type },
      { type: userActions.setUserSettings.type, payload: { ...defaultState.users.userSettings } },
      { type: getUserSettings.fulfilled.type },
      {
        type: userActions.setUserSettings.type,
        payload: {
          ...defaultState.users.userSettings,
          columnSelection: [],
          onboarding: {
            ...defaultOnboardingState,
            complete: true,
            progress: 'onboarding-canceled'
          }
        }
      },
      { type: saveUserSettings.fulfilled.type },
      { type: advanceOnboarding.fulfilled.type },
      { type: setOnboardingCanceled.fulfilled.type }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => Object.keys(action).map(key => expect(storeActions[index][key]).toEqual(action[key])));
  });
  it('should try to derive the onboarding state based on the stored state of the environment', async () => {
    const store = mockStore({ ...defaultState });
    const stepNames = Object.keys(onboardingSteps);
    await store.dispatch(getOnboardingState(stepNames[0]));
    const expectedActions = expectedOnboardingActions;
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => Object.keys(action).map(key => expect(storeActions[index][key]).toEqual(action[key])));
  });
});
