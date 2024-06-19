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
import Cookies from 'universal-cookie';

import { DEVICE_STATES } from '../constants/deviceConstants';
import {
  SET_DEMO_ARTIFACT_PORT,
  SET_ONBOARDING_APPROACH,
  SET_ONBOARDING_COMPLETE,
  SET_ONBOARDING_DEVICE_TYPE,
  SET_ONBOARDING_PROGRESS,
  SET_ONBOARDING_STATE,
  SET_SHOW_ONBOARDING_HELP,
  SET_SHOW_ONBOARDING_HELP_DIALOG,
  onboardingSteps as onboardingStepNames
} from '../constants/onboardingConstants';
import { getDemoDeviceAddress } from '../helpers';
import { getOnboardingState as getCurrentOnboardingState, getUserCapabilities } from '../selectors';
import Tracking from '../tracking';
import { applyOnboardingFallbacks, onboardingSteps } from '../utils/onboardingmanager';
import { saveUserSettings } from './userActions';

const cookies = new Cookies();

const deductOnboardingState = ({ devicesById, devicesByStatus, onboardingState, pastDeployments, releases, userCapabilities, userId }) => {
  const { canDeploy, canManageDevices, canReadDeployments, canReadDevices, canReadReleases, canUploadReleases } = userCapabilities;
  const userCookie = cookies.get(`${userId}-onboarded`);
  const acceptedDevices = devicesByStatus[DEVICE_STATES.accepted].deviceIds;
  const pendingDevices = devicesByStatus[DEVICE_STATES.pending].deviceIds;
  let deviceType = onboardingState.deviceType ?? [];
  deviceType =
    !deviceType.length && acceptedDevices.length && devicesById[acceptedDevices[0]].hasOwnProperty('attributes')
      ? devicesById[acceptedDevices[0]].attributes.device_type
      : deviceType;
  const progress = applyOnboardingFallbacks(onboardingState.progress || determineProgress(acceptedDevices, pendingDevices, releases, pastDeployments));
  return {
    ...onboardingState,
    complete: !!(
      Boolean(userCookie) ||
      onboardingState.complete ||
      (acceptedDevices.length > 1 && pendingDevices.length > 0 && releases.length > 1 && pastDeployments.length > 1) ||
      (acceptedDevices.length >= 1 && releases.length >= 2 && pastDeployments.length > 2) ||
      (acceptedDevices.length >= 1 && pendingDevices.length > 0 && releases.length >= 2 && pastDeployments.length >= 2) ||
      Object.keys(onboardingSteps).findIndex(step => step === progress) >= Object.keys(onboardingSteps).length - 1 ||
      onboardingState.disable ||
      ![canDeploy, canManageDevices, canReadDeployments, canReadDevices, canReadReleases, canUploadReleases].every(i => i)
    ),
    showTips: onboardingState.showTips != null ? onboardingState.showTips : true,
    deviceType,
    approach: onboardingState.approach || (deviceType.some(type => type.startsWith('qemu')) ? 'virtual' : 'physical'),
    progress
  };
};

export const getOnboardingState = () => (dispatch, getState) => {
  const store = getState();
  let onboardingState = getCurrentOnboardingState(store);
  if (!onboardingState.complete) {
    const userId = getState().users.currentUser;
    onboardingState = deductOnboardingState({
      devicesById: store.devices.byId,
      devicesByStatus: store.devices.byStatus,
      onboardingState,
      pastDeployments: store.deployments.byStatus.finished.deploymentIds,
      releases: Object.values(store.releases.byId),
      userCapabilities: getUserCapabilities(store),
      userId
    });
  }
  onboardingState.progress = onboardingState.progress || onboardingStepNames.DASHBOARD_ONBOARDING_START;
  const demoDeviceAddress = `http://${getDemoDeviceAddress(Object.values(store.devices.byId), onboardingState.approach)}`;
  onboardingState.address = store.onboarding.demoArtifactPort ? `${demoDeviceAddress}:${store.onboarding.demoArtifactPort}` : demoDeviceAddress;
  return Promise.resolve(dispatch(setOnboardingState(onboardingState)));
};

export const setShowOnboardingHelp =
  (show, update = true) =>
  dispatch => {
    let tasks = [dispatch({ type: SET_SHOW_ONBOARDING_HELP, show })];
    if (update) {
      tasks.push(dispatch(saveUserSettings({ onboarding: { showTips: show } })));
    }
    return Promise.all(tasks);
  };

const setOnboardingProgress = value => dispatch => dispatch({ type: SET_ONBOARDING_PROGRESS, value });

export const setOnboardingDeviceType =
  (value, update = true) =>
  dispatch => {
    let tasks = [dispatch({ type: SET_ONBOARDING_DEVICE_TYPE, value })];
    if (update) {
      tasks.push(dispatch(saveUserSettings({ onboarding: { deviceType: value } })));
    }
    return Promise.all(tasks);
  };

export const setOnboardingApproach =
  (value, update = true) =>
  dispatch => {
    let tasks = [dispatch({ type: SET_ONBOARDING_APPROACH, value })];
    if (update) {
      tasks.push(dispatch(saveUserSettings({ onboarding: { approach: value } })));
    }
    return Promise.all(tasks);
  };

export const setShowDismissOnboardingTipsDialog = show => dispatch => dispatch({ type: SET_SHOW_ONBOARDING_HELP_DIALOG, show });

export const setDemoArtifactPort = port => dispatch => dispatch({ type: SET_DEMO_ARTIFACT_PORT, value: port });

export const setOnboardingComplete = val => dispatch => {
  let tasks = [Promise.resolve(dispatch({ type: SET_ONBOARDING_COMPLETE, complete: val }))];
  if (val) {
    tasks.push(Promise.resolve(dispatch({ type: SET_SHOW_ONBOARDING_HELP, show: false })));
    tasks.push(Promise.resolve(dispatch(advanceOnboarding(onboardingStepNames.DEPLOYMENTS_PAST_COMPLETED))));
  }
  return Promise.all(tasks);
};

export const setOnboardingCanceled = () => dispatch =>
  Promise.all([
    Promise.resolve(dispatch(setShowOnboardingHelp(false, false))),
    Promise.resolve(dispatch(setShowDismissOnboardingTipsDialog(false))),
    Promise.resolve(dispatch({ type: SET_ONBOARDING_COMPLETE, complete: true }))
  ])
    // using DEPLOYMENTS_PAST_COMPLETED to ensure we get the intended onboarding state set after
    // _advancing_ the onboarding progress
    .then(() => dispatch(advanceOnboarding(onboardingStepNames.DEPLOYMENTS_PAST_COMPLETED_FAILURE)))
    // since we can't advance after ONBOARDING_CANCELED, track the step manually here
    .then(() => Tracking.event({ category: 'onboarding', action: onboardingSteps.ONBOARDING_CANCELED }));

const setOnboardingState = state => dispatch =>
  Promise.all([
    dispatch(setOnboardingComplete(state.complete)),
    dispatch({ type: SET_ONBOARDING_STATE, value: state }),
    dispatch(saveUserSettings({ onboarding: state }))
  ]);

export const advanceOnboarding = stepId => (dispatch, getState) => {
  const steps = Object.keys(onboardingSteps);
  const progress = steps.findIndex(step => step === getState().onboarding.progress);
  const stepIndex = steps.findIndex(step => step === stepId);
  // if there is no progress set yet, the onboarding state deduction hasn't happened
  // and the subsequent settings persistence would overwrite what we stored
  if (progress > stepIndex || getState().onboarding.progress === null) {
    return;
  }
  const madeProgress = steps[stepIndex + 1];
  const state = { ...getCurrentOnboardingState(getState()), progress: madeProgress };
  state.complete =
    stepIndex + 1 >= Object.keys(onboardingSteps).findIndex(step => step === onboardingStepNames.DEPLOYMENTS_PAST_COMPLETED_FAILURE) ? true : state.complete;
  Tracking.event({ category: 'onboarding', action: stepId });
  return Promise.all([dispatch(setOnboardingProgress(madeProgress)), dispatch(saveUserSettings({ onboarding: state }))]);
};

const determineProgress = (acceptedDevices, pendingDevices, releases, pastDeployments) => {
  const steps = Object.keys(onboardingSteps);
  let progress = -1;
  progress = pendingDevices.length > 1 ? steps.findIndex(step => step === onboardingStepNames.DEVICES_PENDING_ACCEPTING_ONBOARDING) : progress;
  progress = acceptedDevices.length >= 1 ? steps.findIndex(step => step === onboardingStepNames.DEVICES_ACCEPTED_ONBOARDING) : progress;
  progress =
    acceptedDevices.length > 1 && releases.length > 1 && pastDeployments.length > 1
      ? steps.findIndex(step => step === onboardingStepNames.DEPLOYMENTS_PAST_COMPLETED)
      : progress;
  return steps[progress];
};
