// Copyright 2021 Northern.tech AS
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

import configureStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import DeviceTwin, { Title, TwinError, TwinSyncStatus } from './devicetwin';

const mockStore = configureStore([thunk]);

let store;

describe('DeviceTwin Component', () => {
  beforeEach(() => {
    store = mockStore({ ...defaultState });
  });
  it('renders correctly', async () => {
    const { baseElement } = render(
      <Provider store={store}>
        <DeviceTwin
          device={{ ...defaultState.devices.byId.a1, twinsByIntegration: { a123: { something: 'test', other: 'misc', ab: 12, nest: { here: 'some' } } } }}
          integration={{ id: 'a123', provider: 'iot-hub' }}
        />
      </Provider>
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  // ordered like this to trigger empty state + diff count state
  [TwinSyncStatus, TwinSyncStatus, TwinSyncStatus, Title, TwinError].forEach((Component, index) => {
    it(`renders sub component ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(
        <Provider store={store}>
          <Component
            diffCount={index}
            twinError={index > 1 ? 'twinError' : ''}
            providerTitle="Test"
            twinTitle="Device Twin"
            updateTime={defaultState.devices.byId.a1.updated_ts}
          />
        </Provider>
      );
      const view = baseElement.lastChild.firstChild;
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
    });
  });
});
