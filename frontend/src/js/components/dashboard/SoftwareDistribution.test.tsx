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
import { TIMEOUTS, chartTypes, rootfsImageVersion } from '@northern.tech/store/constants';
import { act, waitFor } from '@testing-library/react';
import { vi } from 'vitest';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import SoftwareDistribution from './SoftwareDistribution';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      isEnterprise: true
    }
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

describe('Devices Component', () => {
  it('renders correctly', async () => {
    const DeviceActions = await import('@northern.tech/store/devicesSlice/thunks');
    const reportsSpy = vi.spyOn(DeviceActions, 'getReportDataWithoutBackendSupport');
    const reportsUpdateSpy = vi.spyOn(DeviceActions, 'updateReportData');

    const ui = <SoftwareDistribution />;

    const { baseElement, rerender } = render(ui, { preloadedState });
    await act(async () => {
      vi.runAllTimers();
      vi.runAllTicks();
      return new Promise(resolve => resolve(), TIMEOUTS.threeSeconds);
    });
    await waitFor(() => expect(reportsSpy).toHaveBeenCalled());
    await waitFor(() => expect(reportsUpdateSpy).toHaveBeenCalled());
    await waitFor(() => rerender(ui));
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    reportsSpy.mockClear();
  });

  it('renders correctly without data retrieval', async () => {
    const DeviceActions = await import('@northern.tech/store/devicesSlice/thunks');
    const reportsSpy = vi.spyOn(DeviceActions, 'getReportDataWithoutBackendSupport');
    const reportsUpdateSpy = vi.spyOn(DeviceActions, 'updateReportData');

    const ui = <SoftwareDistribution />;
    const { baseElement, rerender } = render(ui, {
      preloadedState: {
        ...preloadedState,
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
        }
      }
    });
    await act(async () => {
      vi.runAllTimers();
      vi.runAllTicks();
      return new Promise(resolve => resolve(), TIMEOUTS.threeSeconds);
    });
    await waitFor(() => expect(reportsUpdateSpy).toHaveBeenCalled());
    expect(reportsSpy).not.toHaveBeenCalled();
    await waitFor(() => rerender(ui));
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders correctly for non-enterprise', async () => {
    const { baseElement } = render(<SoftwareDistribution />, {
      preloadedState: {
        ...preloadedState,
        app: {
          ...preloadedState.app,
          features: {
            ...preloadedState.app.features,
            isEnterprise: false
          }
        }
      }
    });
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders correctly for too many devices', async () => {
    const { baseElement } = render(<SoftwareDistribution />, {
      preloadedState: {
        ...preloadedState,
        devices: {
          ...preloadedState.devices,
          byStatus: {
            ...preloadedState.devices.byStatus,
            accepted: { ...preloadedState.devices.byStatus.accepted, total: 200000000 }
          }
        }
      }
    });
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders correctly while waiting for initialization', async () => {
    const { baseElement } = render(<SoftwareDistribution />, {
      preloadedState: {
        ...preloadedState,
        users: {
          ...preloadedState.users,
          settingsInitialized: false
        }
      }
    });
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
