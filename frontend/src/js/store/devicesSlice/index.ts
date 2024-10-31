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
// @ts-nocheck
import { DEVICE_LIST_DEFAULTS, SORTING_OPTIONS } from '@northern.tech/store/commonConstants';
import { deepCompare, duplicateFilter } from '@northern.tech/utils/helpers';
import { createSlice } from '@reduxjs/toolkit';

import { DEVICE_STATES } from './constants';

export const sliceName = 'devices';

export const initialState = {
  byId: {
    // [deviceId]: {
    //   ...,
    //   twinsByIntegration: { [external.provider.id]: twinData }
    // }
  },
  byStatus: {
    [DEVICE_STATES.accepted]: { deviceIds: [], total: 0 },
    active: { deviceIds: [], total: 0 },
    inactive: { deviceIds: [], total: 0 },
    [DEVICE_STATES.pending]: { deviceIds: [], total: 0 },
    [DEVICE_STATES.preauth]: { deviceIds: [], total: 0 },
    [DEVICE_STATES.rejected]: { deviceIds: [], total: 0 }
  },
  deviceList: {
    deviceIds: [],
    ...DEVICE_LIST_DEFAULTS,
    selectedAttributes: [],
    selectedIssues: [],
    selection: [],
    sort: {
      direction: SORTING_OPTIONS.desc
      // key: null,
      // scope: null
    },
    state: DEVICE_STATES.accepted,
    total: 0
  },
  filters: [
    // { key: 'device_type', value: 'raspberry', operator: '$eq', scope: 'inventory' }
  ],
  filteringAttributes: { identityAttributes: [], inventoryAttributes: [], systemAttributes: [], tagAttributes: [] },
  filteringAttributesLimit: 10,
  filteringAttributesConfig: {
    attributes: {
      // inventory: ['some_attribute']
    },
    count: 0,
    limit: 100
  },
  reports: [
    // { items: [{ key: "someKey", count: 42  }], otherCount: 123, total: <otherCount + itemsCount> }
  ],
  total: 0,
  limit: 0,
  groups: {
    byId: {
      // groupName: { deviceIds: [], total: 0, filters: [] },
      // dynamo: { deviceIds: [], total: 3, filters: [{ a: 1 }] }
    },
    selectedGroup: undefined
  }
};

export const devicesSlice = createSlice({
  name: sliceName,
  initialState,
  reducers: {
    receivedGroups: (state, action) => {
      state.groups.byId = action.payload;
    },
    addToGroup: (state, action) => {
      const { group, deviceIds } = action.payload;
      const maybeExistingGroup = {
        filters: [],
        deviceIds: [],
        ...state.groups.byId[group]
      };
      state.groups.byId[group] = {
        ...maybeExistingGroup,
        deviceIds: [...maybeExistingGroup.deviceIds, ...deviceIds].filter(duplicateFilter),
        total: (maybeExistingGroup.total || 0) + 1
      };
    },
    removeFromGroup: (state, action) => {
      const { group, deviceIds: removedIds } = action.payload;
      const { deviceIds = [], total = 0, ...maybeExistingGroup } = state.groups.byId[group] || {};
      const changedGroup = {
        ...maybeExistingGroup,
        deviceIds: deviceIds.filter(id => !removedIds.includes(id)),
        total: Math.max(total - removedIds.length, 0)
      };
      if (changedGroup.total || changedGroup.deviceIds.length) {
        state.groups.byId[group] = changedGroup;
        return;
      } else if (state.groups.selectedGroup === group) {
        state.groups.selectedGroup = undefined;
      }
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [group]: removal, ...remainingById } = state.groups.byId;
      state.groups.byId = remainingById;
    },
    addGroup: (state, action) => {
      const { groupName, group } = action.payload;
      state.groups.byId[groupName] = {
        ...state.groups.byId[groupName],
        ...group
      };
    },
    selectGroup: (state, { payload: group }) => {
      state.deviceList.deviceIds = state.groups.byId[group] && state.groups.byId[group].deviceIds.length > 0 ? state.groups.byId[group].deviceIds : [];
      state.groups.selectedGroup = group;
    },
    removeGroup: (state, action) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [action.payload]: removal, ...remainingById } = state.groups.byId;
      state.groups.byId = remainingById;
      state.groups.selectedGroup = state.groups.selectedGroup === action.payload ? undefined : state.groups.selectedGroup;
    },
    setDeviceListState: (state, action) => {
      state.deviceList = {
        ...state.deviceList,
        ...action.payload,
        sort: {
          ...state.deviceList.sort,
          ...action.payload.sort
        }
      };
    },
    setFilterAttributes: (state, action) => {
      state.filteringAttributes = action.payload;
    },
    setFilterablesConfig: (state, action) => {
      state.filteringAttributesConfig = action.payload;
    },
    receivedDevices: (state, action) => {
      state.byId = {
        ...state.byId,
        ...action.payload
      };
    },
    setDeviceFilters: (state, action) => {
      if (deepCompare(action.payload, state.filters)) {
        return;
      }
      state.filters = action.payload.filter(filter => filter.key && filter.operator && filter.scope && typeof filter.value !== 'undefined');
    },
    setInactiveDevices: (state, action) => {
      const { activeDeviceTotal, inactiveDeviceTotal } = action.payload;
      state.byStatus.active.total = activeDeviceTotal;
      state.byStatus.inactive.total = inactiveDeviceTotal;
    },
    setDeviceReports: (state, action) => {
      state.reports = action.payload;
    },
    setDevicesByStatus: (state, action) => {
      const { forceUpdate, status, total, deviceIds } = action.payload;
      state.byStatus[status] = total || forceUpdate ? { deviceIds, total } : state.byStatus[status];
    },
    setDevicesCountByStatus: (state, action) => {
      const { count, status } = action.payload;
      state.byStatus[status].total = count;
    },
    setTotalDevices: (state, action) => {
      state.total = action.payload;
    },
    setDeviceLimit: (state, action) => {
      state.limit = action.payload;
    },
    receivedDevice: (state, action) => {
      state.byId[action.payload.id] = {
        ...state.byId[action.payload.id],
        ...action.payload
      };
    },
    maybeUpdateDevicesByStatus: (state, action) => {
      const { deviceId, authId } = action.payload;
      const device = state.byId[deviceId];
      const hasMultipleAuthSets = authId ? device.auth_sets.filter(authset => authset.id !== authId).length > 0 : false;
      if (!hasMultipleAuthSets && Object.values(DEVICE_STATES).includes(device.status)) {
        const deviceIds = state.byStatus[device.status].deviceIds.filter(id => id !== deviceId);
        state.byStatus[device.status] = { deviceIds, total: Math.max(0, state.byStatus[device.status].total - 1) };
      }
    }
  }
  // extraReducers: {
  //   [setDeviceListState.fulfilled]: (state, action) => {
  //     console.log('action', action);
  //     console.log('action', action);
  //     console.log('action', action);
  //     console.log('action', action);
  //     state.deviceList = {
  //       ...state.deviceList,
  //       ...action.payload,
  //       sort: {
  //         ...state.deviceList.sort,
  //         ...action.payload.sort
  //       }
  //     };
  //   }
  // }
  // The following would likely be the way to go with a simpler store setup, but ours seems too entangled for this to provide much benefit
  // import * as actionCreators from './thunks';
  // extraReducers: {
  //   [actionCreators.setDeviceListState.fulfilled]: setDeviceListState
  // }
});

export const actions = devicesSlice.actions;
export default devicesSlice.reducer;
