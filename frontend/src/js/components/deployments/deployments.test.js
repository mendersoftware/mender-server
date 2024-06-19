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

import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';

import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, mockDate, undefineds } from '../../../../tests/mockData';
import { render, selectMaterialUiSelectOption } from '../../../../tests/setupTests';
import GeneralApi from '../../api/general-api';
import { ALL_DEVICES } from '../../constants/deviceConstants';
import Deployments from './deployments';

const defaultLocationProps = { location: { search: 'startDate=2019-01-01' }, match: {} };

const specialKeys = {
  ArrowDown: '{ArrowDown}',
  Enter: '{Enter}'
};

describe('Deployments Component', () => {
  const mockState = {
    ...defaultState,
    app: {
      ...defaultState.app,
      features: {
        ...defaultState.app.features,
        isEnterprise: true
      }
    },
    deployments: {
      ...defaultState.deployments,
      byId: {},
      byStatus: {
        ...defaultState.deployments.byStatus,
        finished: { deploymentIds: [], total: 0 },
        inprogress: { deploymentIds: [], total: 0 },
        pending: { deploymentIds: [], total: 0 }
      },
      selectionState: {
        ...defaultState.deployments.selectionState,
        finished: { ...defaultState.deployments.selectionState.finished, selection: [] },
        inprogress: { ...defaultState.deployments.selectionState.inprogress, selection: [] },
        pending: { ...defaultState.deployments.selectionState.pending, selection: [] }
      }
    },
    releases: {
      ...defaultState.releases,
      releasesList: {
        ...defaultState.releases.releasesList,
        releaseIds: []
      }
    }
  };

  afterEach(async () => {
    await act(async () => {
      jest.advanceTimersByTime(2000);
      jest.runAllTicks();
    });
  });

  it('renders correctly', async () => {
    const get = jest.spyOn(GeneralApi, 'get');
    const ui = <Deployments {...defaultLocationProps} />;
    const { asFragment } = render(ui, { preloadedState: mockState });
    await waitFor(() => expect(screen.getAllByRole('button', { name: /View details/i })).toBeTruthy());
    const view = asFragment();
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await waitFor(() => expect(get).toHaveBeenCalledWith('/api/management/v2/inventory/filters?per_page=500'));
    expect(get).toHaveBeenCalledWith('/api/management/v2/inventory/filters?per_page=500');
  });

  it('works as expected', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const preloadedState = {
      ...mockState,
      deployments: {
        ...mockState.deployments,
        byId: {
          ...defaultState.deployments.byId,
          d1: {
            ...defaultState.deployments.byId.d1,
            artifact_name: 'a1'
          }
        },
        byStatus: {
          ...mockState.deployments.byStatus,
          inprogress: { deploymentIds: ['d1'], total: 1 },
          pending: { deploymentIds: ['d2'], total: 1 }
        },
        selectionState: {
          ...defaultState.deployments.selectionState,
          inprogress: { ...defaultState.deployments.selectionState.inprogress, selection: ['d1'] },
          pending: { ...defaultState.deployments.selectionState.pending, selection: ['d2'] }
        }
      },
      releases: {
        ...defaultState.releases,
        byId: {
          ...defaultState.releases.byId,
          test: {
            ...defaultState.releases.byId.r1
          }
        }
      }
    };
    const ui = (
      <LocalizationProvider dateAdapter={AdapterMoment}>
        <Deployments {...defaultLocationProps} />
      </LocalizationProvider>
    );
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('tab', { name: /Finished/i }));
    await user.click(screen.getByRole('tab', { name: /Scheduled/i }));
    await user.click(screen.getByRole('tab', { name: /Active/i }));
    await user.click(screen.getByRole('button', { name: /Create a deployment/i }));
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.getByText(/Cancel/i)).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /Cancel/i }));
    await waitFor(() => rerender(ui));
    const inprogressDeployments = screen.getByText(/in progress now/i).parentElement.parentElement;
    const deployment = within(inprogressDeployments).getAllByText(/test deployment/i)[0].parentElement.parentElement;
    await user.click(within(deployment).getByRole('button', { name: /Abort/i }));
    act(() => jest.advanceTimersByTime(200));
    await waitFor(() => expect(screen.getByText(/Confirm abort/i)).toBeInTheDocument());
    await user.click(document.querySelector('#confirmAbort').nextElementSibling);
    await waitFor(() => expect(within(deployment).getByRole('button', { name: /View details/i })).toBeVisible());
    await user.click(within(deployment).getByRole('button', { name: /View details/i }));
    await waitFor(() => rerender(ui));
    if (!screen.queryByText(/Deployment details/i)) {
      await user.click(within(deployment).getByRole('button', { name: /View details/i }));
      await waitFor(() => expect(screen.queryByText(/Deployment details/i)).toBeInTheDocument());
    }
    expect(screen.getByText(/Deployment details/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /Close/i }));
  }, 30000);

  it('allows navigating the deployment creation dialog', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const preloadedState = {
      ...mockState,
      app: {
        ...mockState.app,
        features: {
          ...mockState.app.features,
          isEnterprise: false
        }
      }
    };
    const ui = (
      <LocalizationProvider dateAdapter={AdapterMoment}>
        <Deployments {...defaultLocationProps} />
      </LocalizationProvider>
    );
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('tab', { name: /Finished/i }));
    await user.click(screen.getByRole('button', { name: /Create a deployment/i }));
    const releaseId = 'release-998';
    await waitFor(() => rerender(ui));
    act(() => jest.advanceTimersByTime(1000));
    await waitFor(() => expect(screen.queryByPlaceholderText(/Select a Release/i)).toBeInTheDocument(), { timeout: 3000 });
    const releaseSelect = screen.getByPlaceholderText(/Select a Release/i);
    expect(within(releaseSelect).queryByDisplayValue(releaseId)).not.toBeInTheDocument();
    await act(async () => await user.click(releaseSelect));
    await user.keyboard(specialKeys.ArrowDown);
    await user.keyboard(specialKeys.Enter);
    const groupSelect = screen.getByPlaceholderText(/Select a device group/i);
    await act(async () => await user.click(groupSelect));
    await user.keyboard(specialKeys.Enter);
    // await waitFor(() => rerender(ui));
    expect(groupSelect).toHaveValue(ALL_DEVICES);
    const post = jest.spyOn(GeneralApi, 'post');
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
    await user.click(screen.getByRole('button', { name: 'Create deployment' }));
    expect(post).toHaveBeenCalledWith('/api/management/v1/deployments/deployments', {
      all_devices: true,
      artifact_name: releaseId,
      autogenerate_delta: undefined,
      force_installation: false,
      devices: undefined,
      filter_id: undefined,
      group: undefined,
      name: ALL_DEVICES,
      phases: undefined,
      update_control_map: undefined
    });
    await act(() => jest.advanceTimersByTime(1000));
    expect(screen.queryByText(/Cancel/i)).not.toBeInTheDocument();
  }, 20000);

  it('allows navigating the enterprise deployment creation dialog', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const preloadedState = {
      ...mockState,
      app: {
        ...mockState.app,
        features: {
          ...mockState.app.features,
          isHosted: false
        }
      },
      devices: {
        ...mockState.devices,
        byStatus: {
          ...mockState.devices.byStatus,
          accepted: {
            ...mockState.devices.byStatus.accepted,
            deviceIds: [...Object.keys(mockState.devices.byId), 'test1', 'test2'],
            total: Object.keys(mockState.devices.byId).length + 2
          }
        }
      },
      organization: {
        ...mockState.organization,
        organization: {
          ...mockState.organization.organization,
          plan: 'enterprise'
        }
      },
      users: {
        ...mockState.users,
        globalSettings: {
          ...mockState.users.globalSettings,
          previousPhases: [[{ batch_size: 30, delay: 5, delayUnit: 'days' }, { batch_size: 70 }]]
        }
      }
    };
    const ui = (
      <LocalizationProvider dateAdapter={AdapterMoment}>
        <Deployments {...defaultLocationProps} />
      </LocalizationProvider>
    );
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('button', { name: /Create a deployment/i }));
    const releaseId = 'release-998';
    const groupSelect = screen.getByPlaceholderText(/Select a device group/i);
    await act(async () => await user.click(groupSelect));
    await user.keyboard(specialKeys.Enter);
    expect(groupSelect).toHaveValue(ALL_DEVICES);
    await waitFor(() => expect(screen.queryByPlaceholderText(/Select a Release/i)).toBeInTheDocument(), { timeout: 3000 });
    const releaseSelect = screen.getByPlaceholderText(/Select a Release/i);
    await act(async () => await user.click(releaseSelect));
    await user.keyboard(specialKeys.ArrowDown);
    await user.keyboard(specialKeys.Enter);
    await user.click(screen.getByRole('button', { name: /advanced options/i }));
    await user.click(screen.getByRole('checkbox', { name: /select a rollout pattern/i }));
    await waitFor(() => rerender(ui));
    await selectMaterialUiSelectOption(screen.getByText(/Single phase: 100%/i), /Custom/i, user);
    const firstPhase = screen.getByText(/Phase 1/i).parentElement.parentElement.parentElement;
    await selectMaterialUiSelectOption(within(firstPhase).getByDisplayValue(/hours/i), /minutes/i, user);
    fireEvent.change(within(firstPhase).getByDisplayValue(20), { target: { value: '50' } });
    fireEvent.change(within(firstPhase).getByDisplayValue('2'), { target: { value: '30' } });
    await user.click(screen.getByText(/Add a phase/i));
    const secondPhase = screen.getByText(/Phase 2/i).parentElement.parentElement.parentElement;
    await selectMaterialUiSelectOption(within(secondPhase).getByDisplayValue(/hours/i), /days/i, user);
    expect(within(secondPhase).getByText(/Phases must have at least 1 device/i)).toBeTruthy();
    fireEvent.change(within(secondPhase).getByDisplayValue(10), { target: { value: '25' } });
    fireEvent.change(within(secondPhase).getByDisplayValue('2'), { target: { value: '25' } });
    await user.click(screen.getByRole('checkbox', { name: /save as default/i }));
    const retrySelect = document.querySelector('#deployment-retries-selection');
    await act(async () => await user.click(retrySelect));
    await user.keyboard(specialKeys.ArrowDown);
    await user.keyboard(specialKeys.Enter);
    await user.tab();
    await act(async () => {
      jest.advanceTimersByTime(1000);
      jest.runAllTicks();
    });
    expect(retrySelect).toHaveValue(2);

    // extra explicit here as the general date mocking seems to be ignored by the moment/ date combination
    jest.setSystemTime(mockDate);
    const secondBatchDate = new Date(new Date(mockDate).setMinutes(mockDate.getMinutes() + 30));
    const thirdBatchDate = new Date(new Date(secondBatchDate).setDate(secondBatchDate.getDate() + 25));
    const post = jest.spyOn(GeneralApi, 'post');
    const creationButton = screen.getByText(/Create deployment/i);
    await user.click(creationButton);
    expect(creationButton).toBeDisabled();
    await act(async () => {
      jest.advanceTimersByTime(1000);
      jest.runAllTicks();
    });
    expect(post).toHaveBeenCalledWith('/api/management/v1/deployments/deployments', {
      all_devices: true,
      artifact_name: releaseId,
      autogenerate_delta: undefined,
      devices: undefined,
      filter_id: undefined,
      force_installation: false,
      group: undefined,
      name: ALL_DEVICES,
      phases: [
        { batch_size: 50, delay: 30, delayUnit: 'minutes', start_ts: undefined },
        { batch_size: 25, delay: 25, delayUnit: 'days', start_ts: secondBatchDate.toISOString() },
        { batch_size: null, start_ts: thirdBatchDate.toISOString() }
      ],
      retries: 1,
      update_control_map: undefined
    });
    expect(post).toHaveBeenCalledWith(
      '/api/management/v1/useradm/settings',
      {
        '2fa': 'enabled',
        id_attribute: undefined,
        previousFilters: [],
        previousPhases: [
          [
            { batch_size: 30, delay: 5, delayUnit: 'days' },
            { batch_size: 70, start_ts: 1 }
          ],
          [
            { batch_size: 50, delay: 30, delayUnit: 'minutes' },
            { batch_size: 25, delay: 25, delayUnit: 'days', start_ts: 1 },
            { batch_size: null, start_ts: 2 }
          ]
        ],
        retries: 1,
        hasDeployments: true
      },
      { headers: {} }
    );
  }, 20000);
});
