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

import { render } from '../../../tests/setupTests';
import {
  DashboardOnboardingPendings,
  DashboardOnboardingState,
  DeploymentsInprogress,
  DeploymentsPast,
  DeploymentsPastCompletedFailure,
  DevicePendingTip,
  DevicesAcceptedOnboarding,
  DevicesPendingAcceptingOnboarding,
  GetStartedTip,
  SchedulingAllDevicesSelection,
  SchedulingArtifactSelection,
  SchedulingGroupSelection,
  SchedulingReleaseToDevices
} from './onboardingtips';

describe('OnboardingTips Components', () => {
  describe('DevicePendingTip', () => {
    it('renders correctly', async () => {
      const { baseElement } = render(<DevicePendingTip />);
      const view = baseElement.firstChild.firstChild;
      expect(view).toMatchSnapshot();
    });
  });

  describe('tiny onboarding tips', () => {
    [
      DashboardOnboardingPendings,
      DashboardOnboardingState,
      DeploymentsInprogress,
      DeploymentsPast,
      DeploymentsPastCompletedFailure,
      DevicesAcceptedOnboarding,
      DevicesPendingAcceptingOnboarding,
      GetStartedTip,
      SchedulingAllDevicesSelection,
      SchedulingArtifactSelection,
      SchedulingGroupSelection,
      SchedulingReleaseToDevices
    ].forEach(async Component => {
      it(`renders ${Component.displayName || Component.name} correctly`, () => {
        const { baseElement } = render(<Component createdGroup="testgroup" selectedRelease={{ name: 'test', toString: () => 'test' }} />);
        const view = baseElement.firstChild.firstChild;
        expect(view).toMatchSnapshot();
      });
    });
  });
});
