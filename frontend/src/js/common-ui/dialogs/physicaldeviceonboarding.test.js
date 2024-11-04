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

import { EXTERNAL_PROVIDER } from '@northern.tech/store/constants';
import { act, waitFor } from '@testing-library/react';

import { token, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import PhysicalDeviceOnboarding, { ConvertedImageNote, DeviceTypeSelectionStep, ExternalProviderTip, InstallationStep } from './physicaldeviceonboarding';

const oldHostname = window.location.hostname;

describe('PhysicalDeviceOnboarding Component', () => {
  beforeEach(() => {
    window.location = {
      ...window.location,
      hostname: 'hosted.mender.io'
    };
  });
  afterEach(() => {
    window.location = {
      ...window.location,
      hostname: oldHostname
    };
  });

  describe('tiny onboarding tips', () => {
    [DeviceTypeSelectionStep, InstallationStep, ConvertedImageNote, ExternalProviderTip].forEach(async (Component, index) => {
      it(`renders ${Component.displayName || Component.name} correctly`, () => {
        const { baseElement } = render(
          <Component
            advanceOnboarding={jest.fn}
            connectionString="test"
            docsVersion={''}
            hasConvertedImage={true}
            integrationProvider={EXTERNAL_PROVIDER['iot-hub'].provider}
            hasExternalIntegration={index % 2}
            hasMonitor
            ipAddress="test.address"
            isEnterprise={false}
            isHosted={true}
            isDemoMode={false}
            onboardingState={{ complete: false, showTips: true }}
            onSelect={jest.fn}
            selection="raspberrypi7"
            tenantToken="testtoken"
            token={token}
          />
        );
        const view = baseElement.firstChild;
        expect(view).toMatchSnapshot();
        expect(view).toEqual(expect.not.stringMatching(undefineds));
      });
    });
  });

  it('renders correctly', async () => {
    const { baseElement, store } = render(<PhysicalDeviceOnboarding progress={1} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await waitFor(() => expect(store.getState().onboarding.approach === 'physical').toBeTruthy());
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
  });
});
