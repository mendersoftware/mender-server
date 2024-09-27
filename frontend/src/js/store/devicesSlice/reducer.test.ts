// Copyright 2024 Northern.tech AS
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
// @ts-nocheck

/*eslint import/namespace: ['error', { allowComputed: true }]*/
import reducer, { actions, initialState } from '.';
import { defaultState } from '../../../../tests/mockData';
import { DEVICE_STATES } from './constants';

describe('device reducer', () => {
  it('should return the initial state', async () => {
    expect(reducer(undefined, {})).toEqual(initialState);
  });

  it('should handle RECEIVE_GROUPS', async () => {
    expect(reducer(undefined, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId }).groups.byId).toEqual(
      defaultState.devices.groups.byId
    );
    expect(reducer(initialState, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId }).groups.byId).toEqual(
      defaultState.devices.groups.byId
    );
    expect(
      reducer(initialState, { type: actions.receivedGroups, payload: { testExtra: { deviceIds: [], total: 0, filters: [] } } }).groups.byId.testExtra
    ).toEqual({
      deviceIds: [],
      total: 0,
      filters: []
    });
  });
  it('should handle RECEIVE_GROUP_DEVICES', async () => {
    expect(
      reducer(undefined, {
        type: actions.addGroup,
        payload: {
          groupName: 'testGroupDynamic',
          group: defaultState.devices.groups.byId.testGroupDynamic
        }
      }).groups.byId.testGroupDynamic
    ).toEqual(defaultState.devices.groups.byId.testGroupDynamic);
    expect(
      reducer(initialState, {
        type: actions.addGroup,
        payload: {
          groupName: 'testGroupDynamic',
          group: defaultState.devices.groups.byId.testGroupDynamic
        }
      }).groups.byId.testGroupDynamic
    ).toEqual(defaultState.devices.groups.byId.testGroupDynamic);
  });
  it('should handle RECEIVE_DYNAMIC_GROUPS', async () => {
    expect(reducer(undefined, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId }).groups.byId).toEqual(
      defaultState.devices.groups.byId
    );
    expect(reducer(initialState, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId }).groups.byId).toEqual(
      defaultState.devices.groups.byId
    );
    expect(
      reducer(initialState, { type: actions.receivedGroups, payload: { testExtra: { deviceIds: [], total: 0, filters: [] } } }).groups.byId.testExtra
    ).toEqual({ deviceIds: [], total: 0, filters: [] });
  });
  it('should handle ADD_TO_GROUP', async () => {
    let state = reducer(undefined, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId });
    expect(reducer(state, { type: actions.addToGroup, payload: { group: 'testExtra', deviceIds: ['d1'] } }).groups.byId.testExtra.deviceIds).toHaveLength(1);
    expect(
      reducer(initialState, { type: actions.addToGroup, payload: { group: 'testGroup', deviceIds: ['123', '1243'] } }).groups.byId.testGroup.deviceIds
    ).toHaveLength(2);
  });
  it('should handle REMOVE_FROM_GROUP', async () => {
    let state = reducer(undefined, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId });
    state = reducer(state, { type: actions.selectGroup, payload: 'testGroup' });
    expect(
      reducer(state, { type: actions.removeFromGroup, payload: { group: 'testGroup', deviceIds: [defaultState.devices.groups.byId.testGroup.deviceIds[0]] } })
        .groups.byId.testGroup.deviceIds
    ).toHaveLength(defaultState.devices.groups.byId.testGroup.deviceIds.length - 1);
    expect(
      reducer(state, { type: actions.removeFromGroup, payload: { group: 'testGroup', deviceIds: defaultState.devices.groups.byId.testGroup.deviceIds } }).groups
        .byId.testGroup
    ).toBeFalsy();
    expect(
      reducer(initialState, { type: actions.removeFromGroup, payload: { group: 'testExtra', deviceIds: ['123', '1243'] } }).groups.byId.testExtra
    ).toBeFalsy();
  });
  it('should handle ADD_DYNAMIC_GROUP', async () => {
    expect(
      reducer(undefined, { type: actions.addGroup, payload: { groupName: 'test', group: { something: 'test' } } }).groups.byId.test.something
    ).toBeTruthy();
    expect(
      reducer(initialState, { type: actions.addGroup, payload: { groupName: 'test', group: { something: 'test' } } }).groups.byId.test.something
    ).toBeTruthy();
  });
  it('should handle ADD_STATIC_GROUP', async () => {
    expect(
      reducer(undefined, { type: actions.addGroup, payload: { groupName: 'test', group: { something: 'test' } } }).groups.byId.test.something
    ).toBeTruthy();
    expect(
      reducer(initialState, { type: actions.addGroup, payload: { groupName: 'test', group: { something: 'test' } } }).groups.byId.test.something
    ).toBeTruthy();
  });

  it('should handle REMOVE_DYNAMIC_GROUP', async () => {
    let state = reducer(undefined, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId });
    expect(Object.keys(reducer(state, { type: actions.removeGroup, payload: 'testGroupDynamic' }).groups.byId)).toHaveLength(
      Object.keys(defaultState.devices.groups.byId).length - 1
    );
    expect(Object.keys(reducer(initialState, { type: actions.removeGroup, payload: 'testGroupDynamic' }).groups.byId)).toHaveLength(0);
  });
  it('should handle REMOVE_STATIC_GROUP', async () => {
    let state = reducer(undefined, { type: actions.receivedGroups, payload: defaultState.devices.groups.byId });
    expect(Object.keys(reducer(state, { type: actions.removeGroup, payload: 'testGroup' }).groups.byId)).toHaveLength(
      Object.keys(defaultState.devices.groups.byId).length - 1
    );
    expect(Object.keys(reducer(initialState, { type: actions.removeGroup, payload: 'testGroup' }).groups.byId)).toHaveLength(0);
  });
  it('should handle SET_DEVICE_LIST_STATE', async () => {
    expect(reducer(undefined, { type: actions.setDeviceListState, payload: { deviceIds: ['test'] } }).deviceList.deviceIds).toEqual(['test']);
    expect(reducer(initialState, { type: actions.setDeviceListState, payload: { deviceIds: ['test'] } }).deviceList.deviceIds).toEqual(['test']);
  });
  it('should handle SET_DEVICE_FILTERS', async () => {
    expect(reducer(undefined, { type: actions.setDeviceFilters, payload: defaultState.devices.groups.byId.testGroupDynamic.filters }).filters).toHaveLength(1);
    expect(reducer(initialState, { type: actions.setDeviceFilters, payload: [{ key: 'test', operator: 'test' }] }).filters).toHaveLength(0);
  });
  it('should handle SET_FILTERABLES_CONFIG', async () => {
    expect(reducer(undefined, { type: actions.setFilterablesConfig, payload: { attributes: { asd: true } } }).filteringAttributesConfig).toEqual({
      attributes: { asd: true },
      count: undefined,
      limit: undefined
    });
    expect(
      reducer(initialState, { type: actions.setFilterablesConfig, payload: { attributes: { asd: true }, count: 1, limit: 10 } }).filteringAttributesConfig
    ).toEqual({ attributes: { asd: true }, count: 1, limit: 10 });
  });
  it('should handle SET_FILTER_ATTRIBUTES', async () => {
    expect(reducer(undefined, { type: actions.setFilterAttributes, payload: { things: '12' } }).filteringAttributes).toEqual({ things: '12' });
    expect(reducer(initialState, { type: actions.setFilterAttributes, payload: { things: '12' } }).filteringAttributes).toEqual({ things: '12' });
  });
  it('should handle SET_TOTAL_DEVICES', async () => {
    expect(reducer(undefined, { type: actions.setTotalDevices, payload: 2 }).total).toEqual(2);
    expect(reducer(initialState, { type: actions.setTotalDevices, payload: 4 }).total).toEqual(4);
  });
  it('should handle SET_DEVICE_LIMIT', async () => {
    expect(reducer(undefined, { type: actions.setDeviceLimit, payload: 500 }).limit).toEqual(500);
    expect(reducer(initialState, { type: actions.setDeviceLimit, payload: 200 }).limit).toEqual(200);
  });

  it('should handle RECEIVE_DEVICE', async () => {
    expect(reducer(undefined, { type: actions.receivedDevice, payload: defaultState.devices.byId.b1 }).byId.b1).toEqual(defaultState.devices.byId.b1);
    expect(reducer(initialState, { type: actions.receivedDevice, payload: defaultState.devices.byId.b1 }).byId).not.toBe({});
  });
  it('should handle RECEIVE_DEVICES', async () => {
    expect(reducer(undefined, { type: actions.receivedDevices, payload: defaultState.devices.byId }).byId).toEqual(defaultState.devices.byId);
    expect(reducer(initialState, { type: actions.receivedDevices, payload: defaultState.devices.byId }).byId).toEqual(defaultState.devices.byId);
  });
  it('should handle SET_INACTIVE_DEVICES', async () => {
    expect(
      reducer(undefined, { type: actions.setInactiveDevices, payload: { activeDeviceTotal: 1, inactiveDeviceTotal: 1 } }).byStatus.active.total
    ).toBeTruthy();
    expect(
      reducer(initialState, { type: actions.setInactiveDevices, payload: { activeDeviceTotal: 1, inactiveDeviceTotal: 1 } }).byStatus.inactive.total
    ).toEqual(1);
  });
  it('should handle SET_DEVICE_REPORTS', async () => {
    expect(reducer(undefined, { type: actions.setDeviceReports, payload: [1, 2, 3] }).reports).toHaveLength(3);
    expect(reducer(initialState, { type: actions.setDeviceReports, payload: [{ something: 'here' }] }).reports).toEqual([{ something: 'here' }]);
  });
  it('should handle SET_<authstatus>_DEVICES', async () => {
    Object.values(DEVICE_STATES).forEach(status => {
      expect(reducer(undefined, { type: actions.setDevicesByStatus, payload: { deviceIds: ['a1'], total: 1, status } }).byStatus[status]).toEqual({
        deviceIds: ['a1'],
        total: 1
      });
      expect(reducer(initialState, { type: actions.setDevicesByStatus, payload: { deviceIds: ['a1'], status } }).byStatus[status]).toEqual({
        deviceIds: [],
        total: 0
      });
    });
  });
  it('should handle SET_<authstatus>_DEVICES_COUNT', async () => {
    Object.values(DEVICE_STATES).forEach(status => {
      expect(reducer(undefined, { type: actions.setDevicesCountByStatus, payload: { count: 1, status } }).byStatus[status].total).toEqual(1);
      expect(reducer(initialState, { type: actions.setDevicesCountByStatus, payload: { count: 1, status } }).byStatus[status].total).toEqual(1);
    });
  });
});
