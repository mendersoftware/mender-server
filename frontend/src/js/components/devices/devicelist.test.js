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

import { prettyDOM, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import { getHeaders } from './authorized-devices';
import { routes } from './base-devices';
import DeviceList, { calculateResizeChange } from './devicelist';

describe('DeviceList Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <DeviceList
        columnHeaders={[{ attribute: { name: 1 } }, { attribute: { name: 2 } }, { attribute: { name: 3 } }, { attribute: { name: 4 } }]}
        customColumnSizes={[]}
        devices={[]}
        deviceListState={defaultState.devices.deviceList}
        onPageChange={jest.fn}
        pageTotal={50}
      />
    );
    // special snapshot handling here to work around unstable ids in mui code...
    const view = prettyDOM(baseElement.firstChild, 100000, { highlight: false })
      .replace(/id="mui-[0-9]*"/g, '')
      .replace(/aria-labelledby="(mui-[0-9]* *)*"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('allows column resizing calculations', async () => {
    const columnElements = [{ offsetWidth: 120 }, { offsetWidth: 120 }, { offsetWidth: 240 }, { offsetWidth: 120 }, { offsetWidth: 150 }, { offsetWidth: 120 }];
    const columnHeaders = [
      { attribute: { name: 'some', scope: 'thing' } },
      { attribute: { name: 'other', scope: 'thing' } },
      { attribute: { name: 'more', scope: 'thing' } },
      { attribute: { name: 'yet', scope: 'thing' } },
      { attribute: { name: 'different', scope: 'thing' } },
      { attribute: { name: 'last', scope: 'thing' } }
    ];
    let result = calculateResizeChange({ columnElements, columnHeaders, e: { clientX: 80 }, index: 2, prev: 90 });
    expect(result).toEqual([
      { attribute: { name: 'other', scope: 'thing' }, size: 120 },
      { attribute: { name: 'more', scope: 'thing' }, size: 120 },
      { attribute: { name: 'yet', scope: 'thing' }, size: 230 },
      { attribute: { name: 'different', scope: 'thing' }, size: 130 },
      { attribute: { name: 'last', scope: 'thing' }, size: 150 }
    ]);
    result = calculateResizeChange({ columnElements, columnHeaders, e: { clientX: 90 }, index: 2, prev: 80 });
    expect(result).toEqual([
      { attribute: { name: 'other', scope: 'thing' }, size: 120 },
      { attribute: { name: 'more', scope: 'thing' }, size: 120 },
      { attribute: { name: 'yet', scope: 'thing' }, size: 250 },
      { attribute: { name: 'different', scope: 'thing' }, size: 110 },
      { attribute: { name: 'last', scope: 'thing' }, size: 150 }
    ]);
    result = calculateResizeChange({ columnElements, columnHeaders, e: { clientX: 90 }, index: columnElements.length - 1, prev: 80 });
    expect(result).toEqual([
      { attribute: { name: 'other', scope: 'thing' }, size: 120 },
      { attribute: { name: 'more', scope: 'thing' }, size: 120 },
      { attribute: { name: 'yet', scope: 'thing' }, size: 240 },
      { attribute: { name: 'different', scope: 'thing' }, size: 120 },
      { attribute: { name: 'last', scope: 'thing' }, size: 150 }
    ]);
  });

  it('works as expected', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const onExpandClickMock = jest.fn();
    const onResizeColumns = jest.fn();
    const onPageChange = jest.fn();
    const onSelect = jest.fn();
    const onSort = jest.fn();
    const devices = defaultState.devices.byStatus.accepted.deviceIds.map(id => defaultState.devices.byId[id]);
    const pageTotal = devices.length;
    const headers = getHeaders([], routes.allDevices.defaultHeaders, 'id', jest.fn);
    const ui = (
      <DeviceList
        columnHeaders={headers}
        customColumnSizes={[]}
        devices={devices}
        deviceListState={defaultState.devices.deviceList}
        idAttribute="id"
        onExpandClick={onExpandClickMock}
        onResizeColumns={onResizeColumns}
        onPageChange={onPageChange}
        onSelect={onSelect}
        onSort={onSort}
        pageTotal={pageTotal}
      />
    );
    render(ui);
    await user.click(screen.getByText(devices[0].id));
    expect(onExpandClickMock).toHaveBeenCalled();

    await user.click(screen.getAllByRole('checkbox')[0]);
    expect(onSelect).toHaveBeenCalledWith([0, 1]);
    await user.click(screen.getAllByRole('checkbox')[2]);
    expect(onSelect).toHaveBeenCalledWith([1]);
  }, 30000);
});
