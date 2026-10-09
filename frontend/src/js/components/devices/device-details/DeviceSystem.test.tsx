// Copyright 2026 Northern.tech AS
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
import { manifestVersion } from '@northern.tech/store/constants';
import { undefineds } from '@northern.tech/testing/mockData';
import { ATTRIBUTE_SCOPES } from '@northern.tech/utils/constants';
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { DeviceSystem, orchestratorManifestAttribute, systemComponentAttribute } from './DeviceSystem';

const components = {
  R123: [
    { name: 'artifact_name', value: 'rtos-image-v1', scope: ATTRIBUTE_SCOPES.inventory },
    { name: 'component_type', value: 'rtos', scope: ATTRIBUTE_SCOPES.inventory },
    { name: 'version', value: 'v1', scope: ATTRIBUTE_SCOPES.inventory }
  ],
  S789: [
    { name: 'artifact_name', value: 'sensor-firmware-v3', scope: ATTRIBUTE_SCOPES.inventory },
    { name: 'component_type', value: 'sensor', scope: ATTRIBUTE_SCOPES.inventory },
    { name: 'version', value: 'v3.1.0', scope: ATTRIBUTE_SCOPES.inventory }
  ],
  T456: [
    { name: 'component_type', value: 'telematics_control_unit', scope: ATTRIBUTE_SCOPES.inventory },
    { name: 'version', value: 'tcu-v4', scope: ATTRIBUTE_SCOPES.inventory },
    { name: systemComponentAttribute, value: 'true', scope: ATTRIBUTE_SCOPES.inventory }
  ]
};

const device = {
  ...defaultState.devices.byId.a1,
  attributes: {
    ...defaultState.devices.byId.a1.attributes,
    device_type: ['gateway'],
    [orchestratorManifestAttribute]: 'bmw-v8_INCONSISTENT',
    [manifestVersion]: 'bmw-v8_INCONSISTENT',
    system_type: 'system-core'
  },
  components
};

describe('DeviceSystem Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<DeviceSystem device={device} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('shows the system type instead of the device type', async () => {
    render(<DeviceSystem device={device} />);
    expect(screen.getByText('system-core')).toBeVisible();
    expect(screen.queryByText('gateway')).not.toBeInTheDocument();
  });
  it('shows the reported manifest name but links to the manifest without the inconsistency suffix', async () => {
    render(<DeviceSystem device={device} />);
    expect(screen.getByRole('link', { name: 'bmw-v8_INCONSISTENT' })).toHaveAttribute('href', '/software?tab=manifests&id=bmw-v8');
  });
  it('lists the system device component first and links to its inventory and software', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const setDetailsTab = vi.fn();
    render(<DeviceSystem device={device} setDetailsTab={setDetailsTab} />);
    const [, firstRow] = screen.getAllByRole('row');
    expect(within(firstRow).getByText('T456')).toBeVisible();
    expect(within(firstRow).getByText('System device')).toBeVisible();
    expect(within(firstRow).queryByRole('row', { name: /expand row/i })).not.toBeInTheDocument();
    await user.click(within(firstRow).getByRole('button', { name: 'Inventory' }));
    expect(setDetailsTab).toHaveBeenCalledWith('inventory');
    await user.click(within(firstRow).getByRole('button', { name: 'Software' }));
    expect(setDetailsTab).toHaveBeenCalledWith('software');
  });
  it('expands a component row on click', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<DeviceSystem device={device} />);
    expect(screen.queryByText('rtos-image-v1')).not.toBeInTheDocument();
    await user.click(screen.getByText('R123'));
    expect(screen.getByText('rtos-image-v1')).toBeVisible();
  });
});
