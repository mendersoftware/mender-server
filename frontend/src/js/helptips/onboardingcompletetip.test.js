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
import { MemoryRouter } from 'react-router-dom';

import { getConfiguredStore } from '@northern.tech/store/store';
import { act, render as testingLibRender, waitFor } from '@testing-library/react';

import { defaultState, undefineds } from '../../../tests/mockData';
import OnboardingCompleteTip from './onboardingcompletetip';

describe('OnboardingCompleteTip Component', () => {
  let store;
  beforeEach(() => {
    jest.spyOn(global, 'encodeURIComponent').mockImplementationOnce(() => 'http%3A%2F%2Ftest.com');
    store = getConfiguredStore({
      preloadedState: {
        ...defaultState,
        app: {
          ...defaultState.app,
          features: {
            ...defaultState.app.features,
            hasMultitenancy: true,
            isHosted: true
          }
        }
      }
    });
  });

  it('renders correctly', async () => {
    const ui = (
      <MemoryRouter initialEntries={[`/password`]}>
        <Provider store={store}>
          <OnboardingCompleteTip targetUrl="https://test.com" />
        </Provider>
      </MemoryRouter>
    );
    const { baseElement, rerender } = testingLibRender(ui);
    await waitFor(() => rerender(ui));
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
  });
});
