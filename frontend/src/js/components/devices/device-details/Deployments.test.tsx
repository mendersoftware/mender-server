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
import { defaultState, render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { selectMaterialUiSelectOption } from '@northern.tech/testing/utils';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import Deployments from './Deployments';

const deviceDeployments = [
  {
    id: 'deployment123',
    created: '2021-07-08T17:56:49.366Z',
    deviceId: 'a1',
    finished: '2021-07-08T17:58:38.23Z',
    log: true,
    status: 'failure',
    release: 'some-release',
    deploymentStatus: 'inprogress',
    target: 'ALL THE DEVICES'
  }
];

describe('Deployments Component', () => {
  it('renders correctly', async () => {
    const DeploymentActions = await import('@northern.tech/store/deploymentsSlice/thunks');
    const getDeploymentsSpy = vi.spyOn(DeploymentActions, 'getDeviceDeployments');

    const { baseElement } = render(<Deployments device={{ ...defaultState.devices.byId.a1, deploymentsCount: 4, deviceDeployments }} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    expect(getDeploymentsSpy).toHaveBeenCalled();
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });

    await selectMaterialUiSelectOption(screen.getByText(/any/i), /in progress/i, user);
    expect(getDeploymentsSpy).toHaveBeenLastCalledWith({ deviceId: 'a1', filterSelection: ['downloading', 'installing', 'rebooting'], page: 1, perPage: 10 });
  });

  it('retrieves deployment data when clicking the log button', async () => {
    const DeploymentActions = await import('@northern.tech/store/deploymentsSlice/thunks');
    const getDeviceLogSpy = vi.spyOn(DeploymentActions, 'getDeviceLog');
    const getSingleDeploymentSpy = vi.spyOn(DeploymentActions, 'getSingleDeployment');

    render(<Deployments device={{ ...defaultState.devices.byId.a1, deploymentsCount: 1, deviceDeployments }} />);

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const logButton = screen.getByRole('button', { name: /log/i });
    await user.click(logButton);

    expect(getDeviceLogSpy).toHaveBeenCalledWith({ deploymentId: 'deployment123', deviceId: 'a1' });
    expect(getSingleDeploymentSpy).toHaveBeenCalledWith('deployment123');
  });
});
