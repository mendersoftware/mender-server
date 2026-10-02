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
import { defaultState, render } from '@/testUtils';
import { ColumnWidthProvider } from '@northern.tech/common-ui/TwoColumnData';
import { undefineds } from '@northern.tech/testing/mockData';
import { selectMaterialUiSelectOption } from '@northern.tech/testing/utils';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { produce } from 'immer';
import { vi } from 'vitest';

import { RolloutSchedule } from './RolloutSchedule';

const uniformPhaseCount = 25;
const uniformDeployment = {
  ...defaultState.deployments.byId.d3,
  device_count: 250,
  filter: { id: 'filter1', name: 'dynamic group', filters: [] },
  uniform_phases: { batch_size_devices: 10, start_ts: '2019-02-01T11:45:10.002Z', time_interval: '3600s' },
  phases: Array.from({ length: uniformPhaseCount }, (_, index) => ({
    // the backend may still return uuids as phase ids
    id: `a1b2c3d4-0000-4000-8000-${`${index}`.padStart(12, '0')}`,
    batch_size_devices: 10,
    device_count: index < 3 ? 10 : 0,
    initial_batch_device_count: 10,
    start_ts: `2019-02-${`${index + 1}`.padStart(2, '0')}T11:45:10.002Z`,
    status: index < 3 ? 'finished' : index === 3 ? 'inprogress' : 'pending'
  }))
};

describe('RolloutSchedule Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <ColumnWidthProvider>
        <RolloutSchedule
          deployment={produce({ ...defaultState.deployments.byId.d2, phases: [{ id: '0', batch_size: 100, device_count: 1 }] }, i => i)}
          innerRef={vi.fn()}
        />
      </ColumnWidthProvider>
    );
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('shows uniform phases by position and allows changing the page size', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(
      <ColumnWidthProvider>
        <RolloutSchedule deployment={uniformDeployment} />
      </ColumnWidthProvider>
    );
    expect(screen.getByText('Uniform, 10 devices per phase')).toBeInTheDocument();
    expect(screen.getByText('Phase 1')).toBeInTheDocument();
    expect(screen.getByText('Phase 20')).toBeInTheDocument();
    expect(screen.queryByText('Phase 21')).not.toBeInTheDocument();
    expect(screen.queryByText(/a1b2c3d4/)).not.toBeInTheDocument();
    expect(screen.queryByText('(Final phase)')).not.toBeInTheDocument();

    await selectMaterialUiSelectOption(document.querySelector('[name=pagination]'), '50', user);
    expect(screen.getByText(`Phase ${uniformPhaseCount}`)).toBeInTheDocument();
    expect(screen.getAllByText('(Final phase)')).toHaveLength(1);
    // the final phase of an uncapped dynamic deployment keeps growing
    expect(screen.getByText('10*')).toBeInTheDocument();
    expect(screen.getByText(/final phase may adjust as devices change/i)).toBeInTheDocument();
  });
});
