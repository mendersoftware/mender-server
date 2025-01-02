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

import BaseOnboardingTip from '@northern.tech/helptips/BaseOnoardingTip';
import OnboardingCompleteTip from '@northern.tech/helptips/OnboardingCompleteTip';
import {
  DashboardOnboardingPendings,
  DashboardOnboardingState,
  DeploymentsInprogress,
  DeploymentsPast,
  DeploymentsPastCompletedFailure,
  DevicePendingTip,
  DevicesAcceptedOnboarding,
  DevicesDeployReleaseOnboarding,
  DevicesPendingAcceptingOnboarding,
  DevicesPendingDelayed,
  GetStartedTip,
  SchedulingAllDevicesSelection,
  SchedulingArtifactSelection,
  SchedulingGroupSelection,
  SchedulingReleaseToDevices
} from '@northern.tech/helptips/OnboardingTips';
import { DEPLOYMENT_STATES, onboardingSteps as stepNames, yes } from '@northern.tech/store/constants';

export const onboardingSteps = {
  [stepNames.DASHBOARD_ONBOARDING_START]: {
    condition: { min: stepNames.ONBOARDING_START },
    specialComponent: <GetStartedTip />
  },
  [stepNames.DEVICES_PENDING_ONBOARDING_START]: {
    condition: { min: stepNames.DASHBOARD_ONBOARDING_START, max: stepNames.DEVICES_PENDING_ONBOARDING },
    fallbackStep: stepNames.DASHBOARD_ONBOARDING_START,
    specialComponent: <DevicePendingTip />
  },
  [stepNames.DEVICES_DELAYED_ONBOARDING]: {
    condition: { min: stepNames.DASHBOARD_ONBOARDING_START },
    component: DevicesPendingDelayed,
    fallbackStep: stepNames.DASHBOARD_ONBOARDING_START,
    extra: ({ deviceConnection }) => {
      const now = new Date();
      const then = new Date(deviceConnection);
      then.setMinutes(then.getMinutes() + 5);
      return !!deviceConnection && then < now;
    }
  },
  [stepNames.DEVICES_PENDING_ONBOARDING]: {
    condition: { min: stepNames.DASHBOARD_ONBOARDING_START },
    component: DashboardOnboardingState,
    fallbackStep: stepNames.DASHBOARD_ONBOARDING_START
  },
  [stepNames.DEVICES_PENDING_ACCEPTING_ONBOARDING]: {
    condition: { min: stepNames.DEVICES_PENDING_ONBOARDING, max: stepNames.DEVICES_ACCEPTED_ONBOARDING },
    component: DevicesPendingAcceptingOnboarding
  },
  [stepNames.DASHBOARD_ONBOARDING_PENDINGS]: {
    condition: { min: stepNames.DEVICES_PENDING_ONBOARDING },
    component: DashboardOnboardingPendings
  },
  [stepNames.DEVICES_ACCEPTED_ONBOARDING]: {
    condition: { min: stepNames.DASHBOARD_ONBOARDING_PENDINGS },
    specialComponent: <DevicesAcceptedOnboarding />
  },
  [stepNames.DEVICES_DEPLOY_RELEASE_ONBOARDING]: {
    condition: {},
    component: DevicesDeployReleaseOnboarding
  },
  [stepNames.SCHEDULING_ALL_DEVICES_SELECTION]: {
    condition: { min: stepNames.DEVICES_ACCEPTED_ONBOARDING, max: stepNames.DEPLOYMENTS_INPROGRESS },
    component: SchedulingAllDevicesSelection
  },
  [stepNames.SCHEDULING_GROUP_SELECTION]: {
    condition: { min: stepNames.DEVICES_ACCEPTED_ONBOARDING, max: stepNames.DEPLOYMENTS_INPROGRESS },
    component: SchedulingGroupSelection
  },
  [stepNames.SCHEDULING_ARTIFACT_SELECTION]: {
    condition: { min: stepNames.SCHEDULING_ALL_DEVICES_SELECTION, max: stepNames.DEPLOYMENTS_INPROGRESS },
    component: SchedulingArtifactSelection
  },
  [stepNames.SCHEDULING_RELEASE_TO_DEVICES]: {
    condition: { min: stepNames.SCHEDULING_ARTIFACT_SELECTION, max: stepNames.DEPLOYMENTS_INPROGRESS },
    component: SchedulingReleaseToDevices
  },
  [stepNames.DEPLOYMENTS_INPROGRESS]: {
    condition: {},
    component: DeploymentsInprogress
  },
  [stepNames.DEPLOYMENTS_PAST]: {
    condition: { min: stepNames.DEPLOYMENTS_INPROGRESS, extra: () => !window.location.pathname.includes(DEPLOYMENT_STATES.finished) },
    component: DeploymentsPast
  },
  [stepNames.DEPLOYMENTS_PAST_COMPLETED]: {
    condition: { max: stepNames.DEPLOYMENTS_PAST_COMPLETED_FAILURE },
    specialComponent: <OnboardingCompleteTip targetUrl="destination-unreachable" />
  },
  [stepNames.DEPLOYMENTS_PAST_COMPLETED_FAILURE]: {
    condition: { max: stepNames.ONBOARDING_CANCELED },
    component: DeploymentsPastCompletedFailure
  },
  [stepNames.ONBOARDING_CANCELED]: {
    condition: {},
    specialComponent: <div />
  }
};

const getOnboardingStepCompleted = (id, onboardingState) => {
  const { progress, complete, showTips } = onboardingState;
  const keys = Object.keys(onboardingSteps);
  const {
    condition: { min = id, max = id },
    extra,
    progressIndex
  } = Object.entries(onboardingSteps).reduce(
    (accu, [key, value], index) => {
      accu = key === id ? { ...accu, ...value } : accu;
      accu.progressIndex = key === progress ? index : accu.progressIndex;
      return accu;
    },
    { progressIndex: 0, extra: yes, condition: { min: id, max: id } }
  );
  return (
    !complete &&
    showTips &&
    progressIndex >= keys.findIndex(step => step === min) &&
    progressIndex <= keys.findIndex(step => step === max) &&
    extra(onboardingState)
  );
};

export const getOnboardingComponentFor = (id, componentProps, params = {}, previousComponent = null) => {
  const step = onboardingSteps[id];
  const isValid = getOnboardingStepCompleted(id, componentProps);
  if (!isValid) {
    return previousComponent;
  }
  if (step.specialComponent) {
    return React.cloneElement(step.specialComponent, params);
  }
  const component = step.component(componentProps);
  return <BaseOnboardingTip id={id} component={component} progress={step.progress || params.progress || null} {...params} />;
};
