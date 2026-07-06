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
import { Route, Routes } from 'react-router';

import { defaultState, render, server } from '@/testUtils';
import { TIMEOUTS, inventoryApiUrlV2 } from '@northern.tech/store/constants';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { HttpResponse, http } from 'msw';
import { vi } from 'vitest';

import { Dashboard } from './Dashboard';

describe('Dashboard Component', () => {
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
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
      return new Promise(resolve => resolve(), TIMEOUTS.threeSeconds);
    });
    await waitFor(() => rerender(ui));
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('allows navigating to pending devices', async () => {
    // the dashboard refreshes device counts from the backend on render, so the mocked response has to align with the accepted-devices-free scenario
    server.use(
      http.get(`${inventoryApiUrlV2}/statistics`, () =>
        HttpResponse.json({
          devices_by_status: {
            accepted: { micro: 0, standard: 0, system: 0 },
            pending: { micro: 0, standard: 1, system: 0 }
          }
        })
      )
    );
    const preloadedState = {
      ...defaultState,
      devices: {
        ...defaultState.devices,
        byStatus: {
          ...defaultState.devices.byStatus,
          accepted: { deviceIds: [], counts: { standard: 0 } }
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
    const { rerender } = render(ui, { preloadedState });
    await waitFor(() => rerender(ui));
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
