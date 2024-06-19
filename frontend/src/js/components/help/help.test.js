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
import { Provider } from 'react-redux';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import { render as testingLibRender } from '@testing-library/react';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import { getConfiguredStore } from '../../reducers';
import { Downloads } from './downloads';
import GettingStarted from './getting-started';
import Help from './help';
import MenderHub from './mender-hub';
import { helpProps } from './mockData';
import Support from './support';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      hasDeviceConfig: true,
      hasDeviceConnect: true,
      hasMonitor: true,
      isEnterprise: true
    },
    versionInformation: { latestRelease: helpProps.versions }
  },
  organization: {
    ...defaultState.organization,
    organization: {
      ...defaultState.organization.organization,
      addons: [
        { enabled: true, name: 'configure' },
        { enabled: true, name: 'monitor' },
        { enabled: true, name: 'troubleshoot' }
      ]
    }
  }
};

describe('Help Component', () => {
  it('renders correctly', async () => {
    const store = getConfiguredStore({ preloadedState });
    const { baseElement } = testingLibRender(
      <Provider store={store}>
        <MemoryRouter initialEntries={['/help/get-started']}>
          <Routes>
            <Route path="help" element={<Help />}>
              <Route path=":section" element={null} />
            </Route>
          </Routes>
        </MemoryRouter>
      </Provider>
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  describe('static components', () => {
    [Downloads, GettingStarted, MenderHub, Support].forEach(Component => {
      it(`renders ${Component.displayName || Component.name} correctly`, () => {
        const { baseElement } = render(<Component {...helpProps} />, { preloadedState });
        const view = baseElement.firstChild.firstChild;
        expect(view).toMatchSnapshot();
        expect(view).toEqual(expect.not.stringMatching(undefineds));
      });
    });
  });
});
