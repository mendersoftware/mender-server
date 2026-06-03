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
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import { defaultState, render } from '@/testUtils';
import GeneralApi from '@northern.tech/store/api/general-api';
import { ALL_DEVICES, TIMEOUTS } from '@northern.tech/store/constants';
import { mockDate, undefineds } from '@northern.tech/testing/mockData';
import { selectMaterialUiSelectOption } from '@northern.tech/testing/utils';
import { act, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import Deployments from './Deployments';

const defaultLocationProps = { location: { search: 'startDate=2019-01-01' }, match: {} };

const specialKeys = {
  ArrowDown: '{ArrowDown}',
  ArrowUp: '{ArrowUp}',
  Enter: '{Enter}'
};

const getSelectWrapper = (el: HTMLElement) => el.closest('[role=combobox]')?.parentElement ?? el;

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
        releaseIds: [],
        selectedTags: ['123']
      }
    }
  };

  it('renders correctly', async () => {
    const get = vi.spyOn(GeneralApi, 'get');
    const ui = <Deployments {...defaultLocationProps} />;
    const { asFragment } = render(ui, { preloadedState: mockState });
    await waitFor(() => expect(screen.getAllByRole('button', { name: /View details/i })).toBeTruthy());
    const view = asFragment();
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await waitFor(() => expect(get).toHaveBeenCalledWith('/api/management/v2/inventory/filters?per_page=500'));
    expect(get).toHaveBeenCalledWith('/api/management/v2/inventory/filters?per_page=500');
  });

  it('works as expected', { timeout: 6 * TIMEOUTS.fiveSeconds }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
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
        },
        selectedTags: ['123']
      }
    };
    const ui = (
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <Deployments {...defaultLocationProps} />
      </LocalizationProvider>
    );
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('tab', { name: /Finished/i }));
    await act(async () => vi.runOnlyPendingTimers());
    await user.click(screen.getByRole('tab', { name: /Scheduled/i }));
    await act(async () => vi.runOnlyPendingTimers());
    await user.click(screen.getByRole('tab', { name: /Active/i }));
    await user.click(screen.getByRole('button', { name: /Create a deployment/i }));
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.getByText(/Cancel/i)).toBeInTheDocument());
    await act(async () => vi.runOnlyPendingTimers());
    await user.click(screen.getByRole('button', { name: /Cancel/i }));
    await waitFor(() => rerender(ui));
    const inprogressDeployments = screen.getByText(/in progress now/i).parentElement.parentElement;
    const deployment = within(inprogressDeployments).getAllByText(/test deployment/i)[0].parentElement;
    await user.click(within(deployment).getByRole('button', { name: /Abort/i }));
    await waitFor(() => rerender(ui));
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
  });

  it('allows navigating the deployment creation dialog', { timeout: 8 * TIMEOUTS.fiveSeconds }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
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
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <Deployments {...defaultLocationProps} />
      </LocalizationProvider>
    );
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('tab', { name: /Finished/i }));
    await act(async () => vi.runOnlyPendingTimers());
    await user.click(screen.getByRole('button', { name: /Create a deployment/i }));
    const releaseId = 'release-500';
    await waitFor(() => rerender(ui));
    await act(() => vi.advanceTimersByTime(1000));
    await waitFor(() => expect(screen.getByRole('button', { name: /select software/i })).toBeInTheDocument(), { timeout: 3000 });
    const releaseSelect = screen.getByRole('button', { name: /select software/i });
    await user.click(releaseSelect);
    await user.click(screen.getByRole('heading', { name: releaseId }));
    const groupSelect = screen.getByPlaceholderText(/Select a device group/i);
    await user.click(groupSelect);
    await user.type(groupSelect, 'testGroupDyn');
    await user.keyboard(specialKeys.ArrowDown);
    await user.keyboard(specialKeys.Enter);

    await waitFor(() => expect(screen.getByRole('button', { name: /advanced options/i })).toBeInTheDocument(), { timeout: 3000 });
    await user.click(screen.getByRole('button', { name: /advanced options/i }));
    await user.click(screen.getByRole('checkbox', { name: /maximum number of devices/i }));
    await waitFor(() => rerender(ui));
    const limitInput = within(screen.getByText(/Finish deployment after/i)).getByRole('textbox');
    await user.clear(limitInput);
    await user.type(limitInput, '123');
    const post = vi.spyOn(GeneralApi, 'post');
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    await user.click(screen.getByRole('button', { name: 'Create deployment' }));
    expect(post).toHaveBeenCalledWith('/api/management/v2/deployments/deployments', {
      artifact_name: releaseId,
      filter_id: 'filter1',
      max_devices: 123,
      name: 'testGroupDynamic'
    });
    await act(() => vi.advanceTimersByTime(1000));
    await waitFor(() => expect(screen.queryByText(/Cancel/i)).not.toBeInTheDocument());
  });

  it('allows navigating the enterprise deployment creation dialog', { timeout: 9 * TIMEOUTS.fiveSeconds }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
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
            counts: { standard: Object.keys(mockState.devices.byId).length + 3 }
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
          previousPhases: [
            [
              { batch_size: 30, delay: 5, delayUnit: 'days', start_ts: 1 },
              { batch_size: 20, delay: 15, delayUnit: 'hours', start_ts: 1 },
              { batch_size: 50, start_ts: 2 }
            ]
          ]
        }
      }
    };
    const ui = (
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <Deployments {...defaultLocationProps} />
      </LocalizationProvider>
    );
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('button', { name: /Create a deployment/i }));
    const releaseId = 'release-500';
    const groupSelect = screen.getByPlaceholderText(/Select a device group/i);
    await act(async () => vi.runOnlyPendingTimers());
    await user.click(groupSelect);
    await user.keyboard(specialKeys.Enter);
    expect(groupSelect).toHaveValue(ALL_DEVICES);
    await waitFor(() => expect(screen.getByRole('button', { name: /select software/i })).toBeInTheDocument(), { timeout: 3000 });
    const releaseSelect = screen.getByRole('button', { name: /select software/i });
    await user.click(releaseSelect);
    await user.click(screen.getByRole('heading', { name: releaseId }));
    await waitFor(() => expect(screen.getByRole('button', { name: /advanced options/i })).toBeInTheDocument(), { timeout: 3000 });
    await user.click(screen.getByRole('button', { name: /advanced options/i }));
    await user.click(screen.getByRole('checkbox', { name: /select a rollout pattern/i }));
    await waitFor(() => rerender(ui));
    await selectMaterialUiSelectOption(getSelectWrapper(screen.getByText('Custom')), /3 phases/i, user);
    const getPhaseRow = (label: RegExp) => screen.getByText(label).closest('tr') as HTMLElement;
    const firstPhase = getPhaseRow(/Phase 1/i);
    await selectMaterialUiSelectOption(getSelectWrapper(within(firstPhase).getByText(/days/i)), /minutes/i, user);
    const [firstBatch, firstDelay] = within(firstPhase).getAllByRole('textbox') as HTMLInputElement[];
    await user.clear(firstBatch);
    await user.type(firstBatch, '40');
    await user.tab();
    await user.clear(firstDelay);
    await user.type(firstDelay, '30');
    await user.tab();
    const secondPhase = getPhaseRow(/Phase 2/i);
    await selectMaterialUiSelectOption(getSelectWrapper(within(secondPhase).getByText(/hours/i)), /days/i, user);
    const [secondBatch, secondDelay] = within(secondPhase).getAllByRole('textbox') as HTMLInputElement[];
    await user.clear(secondBatch);
    await user.type(secondBatch, '20');
    await user.tab();
    await user.clear(secondDelay);
    await user.type(secondDelay, '25');
    await user.tab();
    await user.click(screen.getByText(/Add a phase/i));
    const thirdPhase = getPhaseRow(/Phase 3/i);
    expect(within(thirdPhase).getByText(/rounds down to 0 devices/i)).toBeTruthy();
    const [thirdBatch] = within(thirdPhase).getAllByRole('textbox') as HTMLInputElement[];
    await user.clear(thirdBatch);
    await user.type(thirdBatch, '20');
    await user.tab();
    const retrySelect = document.querySelector('#retries');
    await user.click(retrySelect!);
    await user.keyboard(specialKeys.ArrowUp);
    await user.tab();
    await act(async () => {
      vi.advanceTimersByTime(1000);
      vi.runAllTicks();
    });
    expect(retrySelect).toHaveValue('2');

    // extra explicit here as the general date mocking seems to be ignored by the moment/ date combination
    vi.setSystemTime(mockDate);
    const secondBatchDate = new Date(new Date(mockDate).setMinutes(mockDate.getMinutes() + 30));
    const thirdBatchDate = new Date(new Date(secondBatchDate).setDate(secondBatchDate.getDate() + 25));
    const fourthBatchDate = new Date(new Date(thirdBatchDate).setHours(thirdBatchDate.getHours() + 2));
    const post = vi.spyOn(GeneralApi, 'post');
    const creationButton = screen.getByText(/Create deployment/i);
    await user.click(creationButton);
    expect(creationButton).toBeDisabled();
    await act(async () => {
      vi.advanceTimersByTime(1000);
      vi.runAllTicks();
    });
    expect(post).toHaveBeenCalledWith('/api/management/v1/deployments/deployments', {
      all_devices: true,
      artifact_name: releaseId,
      force_installation: false,
      name: ALL_DEVICES,
      phases: [
        { batch_size: 40 },
        { batch_size: 20, start_ts: secondBatchDate.toISOString() },
        { batch_size: 20, start_ts: thirdBatchDate.toISOString() },
        { start_ts: fourthBatchDate.toISOString() }
      ],
      retries: 1
    });
    expect(post).toHaveBeenCalledWith(
      '/api/management/v1/useradm/settings',
      expect.objectContaining({
        hasDeployments: true,
        previousPhases: [
          [
            { batch_size: 30, batch_size_devices: undefined, delay: 5, delayUnit: 'days', isUniform: false },
            { batch_size: 20, batch_size_devices: undefined, delay: 15, delayUnit: 'hours', start_ts: 1, isUniform: false },
            { batch_size: 50, batch_size_devices: undefined, start_ts: 2 }
          ],
          [{ batch_size: 40 }, { batch_size: 20, start_ts: 1 }, { batch_size: 20, start_ts: 2 }, { batch_size: undefined, start_ts: 3 }]
        ]
      }),
      { headers: {} }
    );
  });

  it('creates a scheduled deployment without phase definitions', { timeout: 8 * TIMEOUTS.fiveSeconds }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const preloadedState = {
      ...mockState,
      app: {
        ...mockState.app,
        features: {
          ...mockState.app.features,
          isHosted: false
        }
      },
      organization: {
        ...mockState.organization,
        organization: {
          ...mockState.organization.organization,
          plan: 'enterprise'
        }
      }
    };
    const ui = (
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <Deployments {...defaultLocationProps} />
      </LocalizationProvider>
    );
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('button', { name: /Create a deployment/i }));
    const releaseId = 'release-998';
    const groupSelect = screen.getByPlaceholderText(/Select a device group/i);
    await act(async () => vi.runOnlyPendingTimers());
    await user.click(groupSelect);
    await user.type(groupSelect, 'testGroupDyn');
    await user.keyboard(specialKeys.ArrowDown);
    await user.keyboard(specialKeys.Enter);
    await waitFor(() => expect(screen.getByRole('button', { name: /select a release/i })).toBeInTheDocument(), { timeout: 3000 });
    await user.click(screen.getByRole('button', { name: /select a release/i }));
    await user.click(screen.getByRole('heading', { name: releaseId }));
    await waitFor(() => expect(screen.getByText(/Start immediately/i)).toBeInTheDocument(), { timeout: 3000 });
    await selectMaterialUiSelectOption(getSelectWrapper(screen.getByText(/Start immediately/i)), /Schedule the start date/i, user);
    await waitFor(() => rerender(ui));
    const dialog = screen.getByRole('dialog');
    await user.click(within(dialog).getByRole('gridcell', { name: '14' }));
    const actionBar = dialog.querySelector('.MuiDialogActions-root') as HTMLElement;
    await user.click(within(actionBar).getByText(/Next/i));
    await user.click(within(actionBar).getByText(/OK/i));
    await act(async () => {
      vi.advanceTimersByTime(1000);
      vi.runAllTicks();
    });
    const post = vi.spyOn(GeneralApi, 'post');
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    await user.click(screen.getByRole('button', { name: 'Create deployment' }));
    expect(post).toHaveBeenCalledWith(
      '/api/management/v2/deployments/deployments',
      expect.objectContaining({
        artifact_name: releaseId,
        filter_id: 'filter1',
        name: 'testGroupDynamic',
        phases: [expect.objectContaining({ batch_size: 100, start_ts: expect.stringMatching(/2019-01-14/) })]
      })
    );
  });

  it('creates a deployment with uniform phase settings', { timeout: 9 * TIMEOUTS.fiveSeconds }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
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
            counts: { standard: Object.keys(mockState.devices.byId).length + 3 }
          }
        }
      },
      organization: {
        ...mockState.organization,
        organization: {
          ...mockState.organization.organization,
          plan: 'enterprise'
        }
      }
    };
    const ui = (
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <Deployments {...defaultLocationProps} />
      </LocalizationProvider>
    );
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('button', { name: /Create a deployment/i }));
    const releaseId = 'release-998';
    const groupSelect = screen.getByPlaceholderText(/Select a device group/i);
    await act(async () => vi.runOnlyPendingTimers());
    await user.click(groupSelect);
    await user.type(groupSelect, 'testGroupDyn');
    await user.keyboard(specialKeys.ArrowDown);
    await user.keyboard(specialKeys.Enter);
    await waitFor(() => expect(screen.getByRole('button', { name: /select a release/i })).toBeInTheDocument(), { timeout: 3000 });
    await user.click(screen.getByRole('button', { name: /select a release/i }));
    await user.click(screen.getByRole('heading', { name: releaseId }));
    await waitFor(() => expect(screen.getByRole('button', { name: /advanced options/i })).toBeInTheDocument(), { timeout: 3000 });
    await user.click(screen.getByRole('button', { name: /advanced options/i }));
    await user.click(screen.getByRole('checkbox', { name: /select a rollout pattern/i }));
    await waitFor(() => rerender(ui));
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    await waitFor(() => expect(screen.getByText('Custom')).toBeInTheDocument(), { timeout: 3000 });
    await selectMaterialUiSelectOption(getSelectWrapper(screen.getByText('Custom')), /Uniform/i, user);
    await waitFor(() => expect(screen.getByText('Summary')).toBeInTheDocument());
    const post = vi.spyOn(GeneralApi, 'post');
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    await user.click(screen.getByRole('button', { name: 'Create deployment' }));
    expect(post).toHaveBeenCalledWith(
      '/api/management/v2/deployments/deployments',
      expect.objectContaining({
        artifact_name: releaseId,
        filter_id: 'filter1',
        name: 'testGroupDynamic',
        uniform_phases: { batch_size: 10, time_interval: '7200s' }
      })
    );
  });
});
