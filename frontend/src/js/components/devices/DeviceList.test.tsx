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
import { TIMEOUTS } from '@northern.tech/store/constants';
import { undefineds } from '@northern.tech/testing/mockData';
import { prettyDOM, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { getHeaders } from './AuthorizedDevices';
import { routes } from './BaseDevices';
import DeviceList from './DeviceList';

describe('DeviceList Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <DeviceList
        columnHeaders={getHeaders([], routes.allDevices.defaultHeaders, 'id', vi.fn)}
        devices={[]}
        deviceListState={{ ...defaultState.devices.deviceList, total: 50 }}
        onPageChange={vi.fn}
      />
    );
    // special snapshot handling here to work around unstable ids in mui code...
    const view = prettyDOM(baseElement.firstChild, 100000, { highlight: false })
      .replace(/(:?aria-labelledby|id)=":.*:"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as expected', { timeout: 3 * TIMEOUTS.refreshDefault }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onExpandClickMock = vi.fn();
    const onPageChange = vi.fn();
    const onSelect = vi.fn();
    const onSort = vi.fn();
    const devices = defaultState.devices.byStatus.accepted.deviceIds.map(id => defaultState.devices.byId[id]);
    const headers = getHeaders([], routes.allDevices.defaultHeaders, 'id', vi.fn);
    const ui = (
      <DeviceList
        columnHeaders={headers}
        devices={devices}
        deviceListState={defaultState.devices.deviceList}
        idAttribute="id"
        onExpandClick={onExpandClickMock}
        onPageChange={onPageChange}
        onSelect={onSelect}
        onSort={onSort}
      />
    );
    render(ui);
    await user.click(screen.getByText(devices[0].id));
    expect(onExpandClickMock).toHaveBeenCalled();

    await user.click(screen.getAllByRole('checkbox')[0]);
    expect(onSelect).toHaveBeenCalledWith([0, 1]);
    await user.click(screen.getAllByRole('checkbox')[2]);
    expect(onSelect).toHaveBeenCalledWith([1]);
  });
});
