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
import { undefineds } from '@northern.tech/testing/mockData';
import { ATTRIBUTE_SCOPES } from '@northern.tech/utils/constants';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import DeviceInventory from './DeviceInventory';
import { systemComponentAttribute } from './DeviceSystem';

describe('DeviceInventory Component', () => {
  it('renders correctly', async () => {
    const attributes = {
      ...defaultState.devices.byId.a1.attributes,
      artifact_name: 'myapp',
      'rootfs-image.version': 'stablev1-beta-final-v0',
      'rootfs-image.checksum': '12341143',
      'test.version': 'test-2',
      'a.whole.lot.of.dots.version': 'test-3',
      'a.whole.lot.of.dots.more': 'test-4',
      'even.more.dots.than.before.version': 'test-5',
      'even.more.dots.than.before.more': 'test-6'
    };
    const { baseElement } = render(<DeviceInventory device={{ attributes, id: 'a1' }} setSnackbar={vi.fn} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('points to the other components for system devices', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const setDetailsTab = vi.fn();
    const components = { T456: [{ name: systemComponentAttribute, value: 'true', scope: ATTRIBUTE_SCOPES.inventory }] };
    const device = { ...defaultState.devices.byId.a1, components, tier: 'system' };
    render(<DeviceInventory device={device} setDetailsTab={setDetailsTab} setSnackbar={vi.fn} />);
    expect(screen.getByText(/component: T456/i)).toBeVisible();
    expect(screen.getByText('System device')).toBeVisible();
    await user.click(screen.getByRole('button', { name: 'View all components' }));
    expect(setDetailsTab).toHaveBeenCalledWith('system');
  });
});
