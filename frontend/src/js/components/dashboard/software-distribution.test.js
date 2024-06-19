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

import { act, waitFor } from '@testing-library/react';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import * as DeviceActions from '../../actions/deviceActions';
import { TIMEOUTS, chartTypes } from '../../constants/appConstants';
import { rootfsImageVersion } from '../../constants/releaseConstants';
import SoftwareDistribution from './software-distribution';

const preloadedState = {
  ...defaultState,
  devices: {
    ...defaultState.devices,
    reports: [
      {
        items: [
          { key: 'something', count: 10 },
          { key: 'somethingMore', count: 20 }
        ],
        otherCount: 12,
        total: 42
      },
      {
        items: [
          { key: 'something', count: 10 },
          { key: 'somethingMore', count: 20 }
        ],
        otherCount: 12,
        total: 42
      }
    ]
  },
  users: {
    ...defaultState.users,
    globalSettings: {
      ...defaultState.users.globalSettings,
      [defaultState.users.currentUser]: {
        ...defaultState.users.globalSettings[defaultState.users.currentUser],
        reports: [
          { group: Object.keys(defaultState.devices.groups.byId)[0], attribute: 'artifact_name', type: 'distribution', chartType: chartTypes.pie.key },
          { group: Object.keys(defaultState.devices.groups.byId)[1], software: rootfsImageVersion, type: 'distribution', chartType: chartTypes.bar.key }
        ]
      }
    }
  }
};

const reportsSpy = jest.spyOn(DeviceActions, 'getReportsDataWithoutBackendSupport');

describe('Devices Component', () => {
  it('renders correctly', async () => {
    const ui = <SoftwareDistribution />;

    const { baseElement, rerender } = render(ui, { preloadedState });
    await act(async () => {
      jest.runAllTimers();
      jest.runAllTicks();
      return new Promise(resolve => resolve(), TIMEOUTS.threeSeconds);
    });
    await waitFor(() => expect(reportsSpy).toHaveBeenCalled());
    await waitFor(() => rerender(ui));
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    reportsSpy.mockClear();
  });

  it('renders correctly for enterprise', async () => {
    const testState = {
      ...preloadedState,
      app: {
        ...preloadedState.app,
        features: {
          ...preloadedState.app.features,
          isEnterprise: true
        }
      }
    };
    const ui = <SoftwareDistribution />;
    const { baseElement, rerender } = render(ui, { preloadedState: testState });
    await act(async () => {
      jest.runAllTimers();
      jest.runAllTicks();
      return new Promise(resolve => resolve(), TIMEOUTS.threeSeconds);
    });
    await waitFor(() => expect(reportsSpy).toHaveBeenCalled());
    await waitFor(() => rerender(ui));
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
