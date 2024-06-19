// Copyright 2023 Northern.tech AS
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

import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import configureStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render, selectMaterialUiSelectOption } from '../../../../../tests/setupTests';
import * as DeploymentActions from '../../../actions/deploymentActions';
import Deployments from './deployments';

const mockStore = configureStore([thunk]);
let store;

describe('Deployments Component', () => {
  beforeEach(() => {
    store = mockStore({ ...defaultState });
  });
  it('renders correctly', async () => {
    const deviceDeployments = [
      {
        id: 'someId',
        created: '2021-07-08T17:56:49.366Z',
        deviceId: 'somne-id',
        finished: '2021-07-08T17:58:38.23Z',
        log: true,
        status: 'failure',
        release: 'some-release',
        deploymentStatus: 'inprogress',
        target: 'ALL THE DEVICES'
      }
    ];
    const getDeploymentsSpy = jest.spyOn(DeploymentActions, 'getDeviceDeployments');

    const { baseElement } = render(
      <Provider store={store}>
        <Deployments device={{ ...defaultState.devices.byId.a1, deploymentsCount: 4, deviceDeployments }} />\
      </Provider>
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    expect(getDeploymentsSpy).toHaveBeenCalled();
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });

    await selectMaterialUiSelectOption(screen.getByText(/any/i), /in progress/i, user);
    expect(getDeploymentsSpy).toHaveBeenLastCalledWith('a1', { filterSelection: ['downloading', 'installing', 'rebooting'], page: 1, perPage: 10 });
  });
});
