// Copyright 2025 Northern.tech AS
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
import { mockState as appMockState } from '@northern.tech/store/appSlice/mocks';
import { mockState as deploymentsMockState } from '@northern.tech/store/deploymentsSlice/mocks';
import { mockState as devicesMockState } from '@northern.tech/store/devicesSlice/mocks';
import { mockState as monitorMockState } from '@northern.tech/store/monitorSlice/mocks';
import { mockState as onboardingMockState } from '@northern.tech/store/onboardingSlice/mocks';
import { mockState as organizationMockState } from '@northern.tech/store/organizationSlice/mocks';
import { mockState as releasesMockState } from '@northern.tech/store/releasesSlice/mocks';
import { getConfiguredStore } from '@northern.tech/store/store';
import { mockState as usersMockState } from '@northern.tech/store/usersSlice/mocks';
import { token as mockToken } from '@northern.tech/testing/mockData';
import { render } from '@northern.tech/testing/setupTests';

export const defaultState = {
  app: { ...appMockState },
  deployments: { ...deploymentsMockState },
  devices: { ...devicesMockState },
  monitor: { ...monitorMockState },
  onboarding: { ...onboardingMockState },
  organization: { ...organizationMockState },
  releases: { ...releasesMockState },
  users: { ...usersMockState }
};

const customRender = (ui, options = {}) => {
  const {
    preloadedState = { ...defaultState, users: { ...defaultState.users, currentSession: { token: mockToken, expiresAt: undefined } } },
    store = getConfiguredStore({ preloadedState }),
    ...remainder
  } = options;
  return { store, ...render(ui, { preloadedState, store, ...remainder }) };
};

export * from '@northern.tech/testing/setupTests';

export { customRender as render };
