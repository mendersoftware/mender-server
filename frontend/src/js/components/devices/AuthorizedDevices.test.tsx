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
import { defaultState } from '@/testUtils';
import { render } from '@/testUtils';
import { TIMEOUTS } from '@northern.tech/store/commonConstants';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import Authorized from './AuthorizedDevices';
import { routes } from './BaseDevices';

const preloadedState = {
  ...defaultState,
  devices: {
    ...defaultState.devices,
    byStatus: {
      ...defaultState.devices.byStatus,
      accepted: {
        deviceIds: [],
        total: 0
      }
    }
  }
};

describe('AuthorizedDevices Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Authorized states={routes} />, { preloadedState });
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('behaves as expected', { timeout: 3 * TIMEOUTS.fiveSeconds }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const DeviceActions = await import('@northern.tech/store/devicesSlice/thunks');
    const UserActions = await import('@northern.tech/store/usersSlice/thunks');
    const setListStateSpy = vi.spyOn(DeviceActions, 'setDeviceListState');
    const setUserSettingsSpy = vi.spyOn(UserActions, 'saveUserSettings');
    const setColumnsSpy = vi.spyOn(UserActions, 'updateUserColumnSettings');

    const testKey = 'testKey';
    const attributeNames = {
      artifact: 'rootfs-image.version',
      deviceType: 'device_type',
      checkInTime: 'check_in_time'
    };
    const pageTotal = defaultState.devices.byStatus.accepted.deviceIds.length;
    const preloadedState = {
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          hasMonitor: true,
          isEnterprise: true
        }
      },
      devices: {
        ...defaultState.devices,
        deviceList: { ...defaultState.devices.deviceList, deviceIds: defaultState.devices.byStatus.accepted.deviceIds },
        byStatus: {
          ...defaultState.devices.byStatus,
          accepted: { ...defaultState.devices.byStatus.accepted, total: pageTotal },
          pending: { ...defaultState.devices.byStatus.pending, total: 4 },
          rejected: { ...defaultState.devices.byStatus.rejected, total: 38 }
        }
      },
      organization: {
        ...defaultState.organization,
        organization: {
          ...defaultState.organization.organization,
          addons: [{ enabled: true, name: 'monitor' }]
        }
      },
      users: {
        ...defaultState.users,
        customColumns: [{ attribute: { name: attributeNames.checkInTime, scope: 'system' }, size: 220 }]
      }
    };
    const ui = (
      <Authorized
        addDevicesToGroup={vi.fn}
        onGroupClick={vi.fn}
        onGroupRemoval={vi.fn}
        onMakeGatewayClick={vi.fn}
        onPreauthClick={vi.fn}
        openSettingsDialog={vi.fn}
        removeDevicesFromGroup={vi.fn}
        showsDialog={false}
      />
    );
    const { rerender } = render(ui, { preloadedState });
    await waitFor(() => expect(screen.getAllByRole('checkbox').length).toBeTruthy());
    await user.click(screen.getAllByRole('checkbox')[0]);
    expect(setListStateSpy).toHaveBeenCalledWith({ selection: [0, 1], setOnly: true, fetchAuth: false, forceRefresh: false, shouldSelectDevices: true });
    const combo = screen.getAllByRole('combobox').find(item => item.textContent?.includes('all'));
    await user.click(combo);
    await user.click(screen.getByRole('option', { name: /devices with issues/i }));
    await user.keyboard('{Escape}');
    expect(setListStateSpy).toHaveBeenCalledWith({
      page: 1,
      refreshTrigger: true,
      selectedIssues: ['offline', 'monitoring'],
      fetchAuth: false,
      forceRefresh: false,
      shouldSelectDevices: true
    });
    await waitFor(() => rerender(ui));
    await user.click(screen.getByRole('button', { name: /table options/i }));
    await waitFor(() => rerender(ui));
    await user.click(screen.getByRole('menuitem', { name: /customize/i }));
    await waitFor(() => expect(screen.queryByText(/Customize Columns/i)).toBeVisible());
    const attributeSelect = await screen.findByLabelText(/add a column/i);
    await user.type(attributeSelect, testKey);
    await user.keyboard('{Enter}');
    await act(() => vi.advanceTimersByTime(5000));
    await waitFor(() => expect(screen.getByLabelText(/add a column/i)).toBeVisible());
    const button = screen.getByRole('button', { name: /Save/i });
    expect(button).not.toBeDisabled();
    await user.click(button);

    expect(setColumnsSpy).toHaveBeenCalledWith({
      columns: [
        { attribute: { name: attributeNames.deviceType, scope: 'inventory' }, size: 150 },
        { attribute: { name: attributeNames.artifact, scope: 'inventory' }, size: 150 },
        { attribute: { name: attributeNames.checkInTime, scope: 'system' }, size: 220 },
        { attribute: { name: testKey, scope: 'inventory' }, size: 150 }
      ]
    });
    expect(setListStateSpy).toHaveBeenCalledWith({
      selectedAttributes: [
        { attribute: attributeNames.deviceType, scope: 'inventory' },
        { attribute: attributeNames.artifact, scope: 'inventory' },
        { attribute: attributeNames.checkInTime, scope: 'system' },
        { attribute: testKey, scope: 'inventory' }
      ],
      fetchAuth: false,
      forceRefresh: false,
      shouldSelectDevices: true
    });
    expect(setUserSettingsSpy).toHaveBeenCalledWith({
      columnSelection: [
        { id: 'inventory-device_type', key: attributeNames.deviceType, name: attributeNames.deviceType, scope: 'inventory', title: 'Device type' },
        { id: 'inventory-rootfs-image.version', key: attributeNames.artifact, name: attributeNames.artifact, scope: 'inventory', title: 'Current software' },
        { id: 'system-check_in_time', key: attributeNames.checkInTime, name: attributeNames.checkInTime, scope: 'system', title: 'Latest activity' },
        { id: 'inventory-testKey', key: testKey, name: testKey, scope: 'inventory', title: testKey }
      ]
    });
    await act(async () => vi.runAllTicks());
  });
});
