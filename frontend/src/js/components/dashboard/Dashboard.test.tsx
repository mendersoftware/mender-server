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
import { Route, Routes } from 'react-router-dom';

import { defaultState, render } from '@/testUtils';
import { TIMEOUTS } from '@northern.tech/store/commonConstants';
import { actions as deviceActions } from '@northern.tech/store/devicesSlice';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { Dashboard } from './Dashboard';

describe('Dashboard Component', () => {
  afterEach(async () => {
    // wait for all requests to settle
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
  });
  it('renders correctly', async () => {
    const preloadedState = {
      ...defaultState,
      deployments: {
        ...defaultState.deployments,
        byStatus: {
          ...defaultState.deployments.byStatus,
          finished: { deploymentIds: ['d1', 'd2'], total: 2 },
          inprogress: { deploymentIds: ['d1', 'd2'], total: 2 },
          pending: { deploymentIds: ['d1', 'd2'], total: 2 }
        }
      }
    };
    const ui = <Dashboard />;
    const { baseElement, rerender } = render(ui, { preloadedState });
    await act(async () => {
      vi.runAllTimers();
      vi.runAllTicks();
      return new Promise(resolve => resolve(), TIMEOUTS.threeSeconds);
    });
    await waitFor(() => rerender(ui));
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('allows navigating to pending devices', async () => {
    const preloadedState = {
      ...defaultState,
      devices: {
        ...defaultState.devices,
        byStatus: {
          ...defaultState.devices.byStatus,
          accepted: { deviceIds: [], total: 0 }
        }
      }
    };
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const ui = (
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/devices/pending" element={<div>pendings route</div>} />
      </Routes>
    );
    const { rerender, store } = render(ui, { preloadedState });
    await waitFor(() => rerender(ui));
    await act(() => store.dispatch({ type: deviceActions.setDevicesCountByStatus.type, payload: { status: 'accepted', count: 0 } }));
    await user.click(screen.getByText(/pending devices/i));
    await waitFor(() => screen.queryByText(/pendings route/i));
    expect(screen.getByText(/pendings route/i)).toBeVisible();
  });

  it('allows navigating to accepted devices', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const ui = (
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/devices/*" element={<div>accepted devices route</div>} />
      </Routes>
    );
    const { rerender } = render(ui);
    await waitFor(() => rerender(ui));
    await user.click(screen.getByText(/Accepted devices/i));
    await waitFor(() => screen.queryByText(/accepted devices route/i));
    expect(screen.getByText(/accepted devices route/i)).toBeVisible();
  });

  it('allows navigating to deployments', async () => {
    const preloadedState = {
      ...defaultState,
      deployments: {
        ...defaultState.deployments,
        byStatus: {
          ...defaultState.deployments.byStatus,
          inprogress: { deploymentIds: ['d2'], total: 1 }
        }
      }
    };
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const ui = (
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/deployments/*" element={<div>deployments route</div>} />
      </Routes>
    );
    const { rerender } = render(ui, { preloadedState });
    await waitFor(() => rerender(ui));
    await user.click(screen.getAllByText('test deployment 2')[0]);
    await waitFor(() => screen.queryByText(/deployments route/i));
    expect(screen.getByText(/deployments route/i)).toBeVisible();
  });
});
