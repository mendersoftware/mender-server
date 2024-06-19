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

import { act, render as testingLibRender } from '@testing-library/react';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { getSessionInfo } from '../../auth';
import { getConfiguredStore } from '../../reducers';
import Settings from './settings';

describe('Settings Component', () => {
  let store;
  beforeEach(() => {
    store = getConfiguredStore({
      preloadedState: {
        ...defaultState,
        app: {
          ...defaultState.app,
          features: {
            ...defaultState.app.features,
            isHosted: false,
            hasMultitenancy: true
          }
        },
        organization: { ...defaultState.organization, organization: {} },
        users: { ...defaultState.users, currentSession: getSessionInfo() }
      }
    });
  });

  it('renders correctly', async () => {
    const { baseElement } = testingLibRender(
      <MemoryRouter initialEntries={['/settings/my-profile']}>
        <Provider store={store}>
          <Routes>
            <Route path="settings" element={<Settings />}>
              <Route path=":section" element={null} />
            </Route>
          </Routes>
        </Provider>
      </MemoryRouter>
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(async () => jest.runAllTicks());
  });
});
