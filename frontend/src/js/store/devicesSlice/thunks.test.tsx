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
import React from 'react';
import { Link } from 'react-router-dom';

import { getSingleDeployment } from '@northern.tech/store/thunks';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { actions } from '.';
import { inventoryDevice } from '../../../../tests/__mocks__/deviceHandlers';
import { defaultState } from '../../../../tests/mockData';
import { act, mockAbortController } from '../../../../tests/setupTests';
import { actions as appActions } from '../appSlice';
import { EXTERNAL_PROVIDER, UNGROUPED_GROUP } from '../constants';
import { actions as deploymentActions } from '../deploymentsSlice';
import { DEVICE_STATES } from './constants';
import {
  addDevicesToGroup,
  addDynamicGroup,
  addStaticGroup,
  applyDeviceConfig,
  decommissionDevice,
  deleteAuthset,
  deriveInactiveDevices,
  deriveReportsData,
  deviceFileUpload,
  getAllDeviceCounts,
  getAllDevicesByStatus,
  getAllDynamicGroupDevices,
  getAllGroupDevices,
  getDeviceAttributes,
  getDeviceAuth,
  getDeviceById,
  getDeviceConfig,
  getDeviceConnect,
  getDeviceCount,
  getDeviceFileDownloadLink,
  getDeviceInfo,
  getDeviceLimit,
  getDeviceTwin,
  getDevicesByStatus,
  getDevicesWithAuth,
  getDynamicGroups,
  getGatewayDevices,
  getGroupDevices,
  getGroups,
  getReportingLimits,
  getReportsData,
  getReportsDataWithoutBackendSupport,
  getSessionDetails,
  getSystemDevices,
  preauthDevice,
  removeDevicesFromGroup,
  removeDynamicGroup,
  removeStaticGroup,
  selectGroup,
  setDeviceConfig,
  setDeviceListState,
  setDeviceTags,
  setDeviceTwin,
  updateDeviceAuth,
  updateDevicesAuth,
  updateDynamicGroup
} from './thunks';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

const groupUpdateSuccessMessage = 'The group was updated successfully';
const getGroupSuccessNotification = groupName => (
  <>
    {groupUpdateSuccessMessage} - <Link to={`/devices?inventory=group:eq:${groupName}`}>click here</Link> to see it.
  </>
);

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const { attributes, check_in_time, updated_ts, ...expectedDevice } = defaultState.devices.byId.a1;
const receivedExpectedDevice = { type: actions.receivedDevices.type, payload: { [defaultState.devices.byId.a1.id]: expectedDevice } };
const defaultDeviceListState = {
  type: actions.setDeviceListState.type,
  payload: {
    deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id],
    isLoading: false,
    total: 2
  }
};
const acceptedDevices = {
  type: actions.setDevicesByStatus.type,
  payload: {
    deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id],
    status: DEVICE_STATES.accepted,
    total: defaultState.devices.byStatus.accepted.total
  }
};

const defaultResults = {
  receivedDynamicGroups: {
    type: actions.receivedGroups.type,
    payload: {
      testGroupDynamic: {
        deviceIds: [],
        filters: [
          { key: 'id', operator: '$in', scope: 'identity', value: ['a1'] },
          { key: 'mac', operator: '$nexists', scope: 'identity', value: false },
          { key: 'kernel', operator: '$exists', scope: 'identity', value: true }
        ],
        id: 'filter1',
        total: 0
      }
    }
  },
  addedUngroupedGroup: {
    type: actions.addGroup.type,
    payload: {
      groupName: UNGROUPED_GROUP.id,
      group: {
        filters: [{ key: 'group', operator: '$nin', scope: 'system', value: [Object.keys(defaultState.devices.groups.byId)[0]] }]
      }
    }
  },
  receiveDefaultDevice: { type: actions.receivedDevices.type, payload: { [defaultState.devices.byId.a1.id]: defaultState.devices.byId.a1 } },
  acceptedDevices,
  receivedExpectedDevice,
  defaultDeviceListState,
  postDeviceAuthActions: [
    { type: setDeviceListState.pending.type },
    { type: getDevicesByStatus.pending.type },
    { type: actions.setDeviceListState.type, payload: { deviceIds: [], isLoading: true, refreshTrigger: true } },
    {
      type: actions.receivedDevices.type,
      payload: { [defaultState.devices.byId.a1.id]: { ...defaultState.devices.byId.a1, updated_ts: inventoryDevice.updated_ts } }
    },
    acceptedDevices,
    { type: getDevicesWithAuth.pending.type },
    receivedExpectedDevice,
    { type: getDevicesWithAuth.fulfilled.type },
    { type: getDevicesByStatus.fulfilled.type },
    {
      type: actions.setDeviceListState.type,
      payload: { deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id], isLoading: false, total: 2 }
    },
    { type: setDeviceListState.fulfilled.type }
  ]
};

/* eslint-disable sonarjs/no-identical-functions */
describe('selecting things', () => {
  it('should allow device list selections', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: setDeviceListState.pending.type },
      { type: getDevicesByStatus.pending.type },
      { type: actions.setDeviceListState.type, payload: { deviceIds: ['a1'], isLoading: true } },
      defaultResults.receivedExpectedDevice,
      defaultResults.acceptedDevices,
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      { type: actions.setDeviceListState.type, payload: { deviceIds: ['a1', 'b1'], isLoading: false } },
      { type: setDeviceListState.fulfilled.type }
    ];
    await store.dispatch(setDeviceListState({ deviceIds: ['a1'] }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow device list selections without device retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: setDeviceListState.pending.type },
      { type: actions.setDeviceListState.type, payload: { deviceIds: ['a1'], isLoading: false } },
      { type: setDeviceListState.fulfilled.type }
    ];
    await store.dispatch(setDeviceListState({ deviceIds: ['a1'], setOnly: true }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow static group selection', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'testGroup';
    await store.dispatch(selectGroup({ group: groupName }));
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { attributes, updated_ts, ...expectedDevice } = defaultState.devices.byId.a1;
    const expectedActions = [
      { type: selectGroup.pending.type },
      { type: actions.setDeviceFilters.type, payload: [] },
      { type: getGroupDevices.pending.type },
      { type: getDevicesByStatus.pending.type },
      { type: actions.selectGroup.type, payload: groupName },
      { type: actions.receivedDevices.type, payload: { [defaultState.devices.byId.a1.id]: { ...expectedDevice, attributes } } },
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      {
        type: actions.addGroup.type,
        payload: { group: { deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id], total: 2 }, groupName }
      },
      { type: getGroupDevices.fulfilled.type },
      { type: selectGroup.fulfilled.type }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow dynamic group selection', async () => {
    const store = mockStore({ ...defaultState });
    await store.dispatch(selectGroup({ group: 'testGroupDynamic' }));
    const expectedActions = [
      { type: selectGroup.pending.type },
      { type: actions.setDeviceFilters.type, payload: [{ scope: 'system', key: 'group', operator: '$eq', value: 'things' }] },
      { type: actions.selectGroup.type, payload: 'testGroupDynamic' },
      { type: selectGroup.fulfilled.type }
    ];
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow dynamic group selection with extra filters', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: selectGroup.pending.type },
      {
        type: actions.setDeviceFilters.type,
        payload: [
          { scope: 'system', key: 'group', operator: '$eq', value: 'things' },
          { scope: 'system', key: 'group2', operator: '$eq', value: 'things2' }
        ]
      },
      { type: actions.selectGroup.type, payload: 'testGroupDynamic' },
      { type: selectGroup.fulfilled.type }
    ];
    await store.dispatch(
      selectGroup({
        group: 'testGroupDynamic',
        filters: [...defaultState.devices.groups.byId.testGroupDynamic.filters, { scope: 'system', key: 'group2', operator: '$eq', value: 'things2' }]
      })
    );
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});

describe('overall device information retrieval', () => {
  it('should allow count retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDeviceCount.pending.type },
      { type: getDeviceCount.pending.type },
      { type: getDeviceCount.pending.type },
      { type: getDeviceCount.pending.type },
      {
        type: actions.setDevicesCountByStatus.type,
        payload: { count: defaultState.devices.byStatus.accepted.total, status: DEVICE_STATES.accepted }
      },
      {
        type: actions.setDevicesCountByStatus.type,
        payload: { count: defaultState.devices.byStatus.pending.total, status: DEVICE_STATES.pending }
      },
      {
        type: actions.setDevicesCountByStatus.type,
        payload: { count: defaultState.devices.byStatus.preauthorized.total, status: DEVICE_STATES.preauth }
      },
      {
        type: actions.setDevicesCountByStatus.type,
        payload: { count: defaultState.devices.byStatus.rejected.total, status: DEVICE_STATES.rejected }
      },
      { type: getDeviceCount.fulfilled.type },
      { type: getDeviceCount.fulfilled.type },
      { type: getDeviceCount.fulfilled.type },
      { type: getDeviceCount.fulfilled.type }
    ];
    await Promise.all(Object.values(DEVICE_STATES).map(status => store.dispatch(getDeviceCount(status)))).then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow count retrieval for all state counts', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getAllDeviceCounts.pending.type },
      { type: getDeviceCount.pending.type },
      { type: getDeviceCount.pending.type },
      ...[DEVICE_STATES.accepted, DEVICE_STATES.pending].map(status => ({
        type: actions.setDevicesCountByStatus.type,
        payload: { count: defaultState.devices.byStatus[status].total, status }
      })),
      { type: getDeviceCount.fulfilled.type },
      { type: getDeviceCount.fulfilled.type },
      { type: getAllDeviceCounts.fulfilled.type }
    ];
    await store.dispatch(getAllDeviceCounts());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should allow limit retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDeviceLimit.pending.type },
      { type: actions.setDeviceLimit.type, payload: defaultState.devices.limit },
      { type: getDeviceLimit.fulfilled.type }
    ];
    await store.dispatch(getDeviceLimit());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow attribute retrieval and group results', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDeviceAttributes.pending.type },
      { type: actions.setFilterAttributes.type, payload: {} },
      { type: getDeviceAttributes.fulfilled.type }
    ];
    await store.dispatch(getDeviceAttributes());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    const receivedAttributes = storeActions.find(item => item.type === actions.setFilterAttributes.type).payload;
    expect(Object.keys(receivedAttributes)).toHaveLength(4);
    Object.entries(receivedAttributes).forEach(([key, value]) => {
      expect(key).toBeTruthy();
      expect(value).toBeTruthy();
    });
  });
  it('should allow attribute config + limit retrieval and group results', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getReportingLimits.pending.type },
      {
        type: actions.setFilterablesConfig.type,
        payload: {
          attributes: {
            identity: ['status', 'mac'],
            inventory: [
              'artifact_name',
              'cpu_model',
              'device_type',
              'hostname',
              'ipv4_wlan0',
              'ipv6_wlan0',
              'kernel',
              'mac_eth0',
              'mac_wlan0',
              'mem_total_kB',
              'mender_bootloader_integration',
              'mender_client_version',
              'network_interfaces',
              'os',
              'rootfs_type'
            ],
            system: ['created_ts', 'updated_ts', 'group']
          },
          count: 20,
          limit: 100
        }
      },
      { type: getReportingLimits.fulfilled.type }
    ];
    await store.dispatch(getReportingLimits());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should allow getting device aggregation data for use in the dashboard/ reports', async () => {
    const store = mockStore({
      ...defaultState,
      devices: { ...defaultState.devices, byStatus: { ...defaultState.devices.byStatus, accepted: { ...defaultState.devices.byStatus.accepted, total: 50 } } }
    });
    const expectedActions = [
      { type: getReportsData.pending.type },
      {
        type: actions.setDeviceReports.type,
        payload: [
          {
            items: [
              { count: 6, key: 'test' },
              { count: 1, key: 'original' }
            ],
            otherCount: 43,
            total: 50
          }
        ]
      },
      { type: getReportsData.fulfilled.type }
    ];
    await store.dispatch(getReportsData());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow getting device aggregation data for use in the dashboard/ reports even if the reporting service is not ready', async () => {
    const groupName = 'testGroup';
    const groupNameDynamic = 'testGroupDynamic';
    const store = mockStore({
      ...defaultState,
      users: {
        ...defaultState.users,
        userSettings: {
          ...defaultState.users.userSettings,
          reports: [{ attribute: 'ipv4_wlan0', chartType: 'bar', group: groupName, type: 'distribution' }]
        }
      }
    });
    const expectedActions = [
      { type: getReportsDataWithoutBackendSupport.pending.type },
      { type: getAllDevicesByStatus.pending.type },
      { type: getGroups.pending.type },
      { type: getDynamicGroups.pending.type },
      { type: actions.receivedGroups.type, payload: { testGroup: defaultState.devices.groups.byId.testGroup } },
      { type: getDevicesByStatus.pending.type },
      defaultResults.receivedDynamicGroups,
      { type: getDynamicGroups.fulfilled.type },
      defaultResults.receivedExpectedDevice,
      defaultResults.acceptedDevices,
      { type: deriveInactiveDevices.pending.type },
      { type: actions.setInactiveDevices.type, payload: { activeDeviceTotal: 0, inactiveDeviceTotal: 2 } },
      { type: deriveReportsData.pending.type },
      { type: actions.setDeviceReports.type, payload: [{ items: [{ count: 2, key: '192.168.10.141/24' }], otherCount: 0, total: 2 }] },
      { type: deriveInactiveDevices.fulfilled.type },
      { type: deriveReportsData.fulfilled.type },
      { type: getAllDevicesByStatus.fulfilled.type },
      defaultResults.receiveDefaultDevice,
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      defaultResults.addedUngroupedGroup,
      { type: getGroups.fulfilled.type },
      { type: getAllGroupDevices.pending.type },
      { type: getAllDynamicGroupDevices.pending.type },
      defaultResults.receivedExpectedDevice,
      {
        type: actions.addGroup.type,
        payload: { group: { deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id], total: 2 }, groupName }
      },
      { type: actions.receivedDevices.type, payload: {} },
      { type: actions.addGroup.type, payload: { group: { deviceIds: [], total: 0 }, groupName: groupNameDynamic } },
      { type: getAllGroupDevices.fulfilled.type },
      { type: getAllDynamicGroupDevices.fulfilled.type },
      { type: deriveReportsData.pending.type },
      { type: actions.setDeviceReports.type, payload: [{ items: [{ count: 2, key: '192.168.10.141/24' }], otherCount: 0, total: 2 }] },
      { type: deriveReportsData.fulfilled.type },
      { type: getReportsDataWithoutBackendSupport.fulfilled.type }
    ];
    await store.dispatch(getReportsDataWithoutBackendSupport());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow system devices retrieval', async () => {
    const store = mockStore({
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          isEnterprise: true
        }
      }
    });
    const expectedActions = [
      { type: getSystemDevices.pending.type },
      {
        type: actions.receivedDevices.type,
        payload: {
          [defaultState.devices.byId.a1.id]: {
            ...defaultState.devices.byId.a1,
            systemDeviceIds: [],
            systemDeviceTotal: 0
          }
        }
      },
      { type: getSystemDevices.fulfilled.type }
    ];
    await store.dispatch(getSystemDevices({ id: defaultState.devices.byId.a1.id }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow system devices retrieval', async () => {
    const gatewayDevice = defaultState.devices.byId.a1;
    const store = mockStore({
      ...defaultState,
      app: {
        ...defaultState.app,
        features: {
          ...defaultState.app.features,
          isEnterprise: true
        }
      },
      devices: {
        ...defaultState.devices,
        byId: {
          ...defaultState.devices.byId,
          [gatewayDevice.id]: {
            ...gatewayDevice,
            attributes: {
              ...gatewayDevice.attributes,
              mender_gateway_system_id: 'gatewaySystem'
            }
          }
        }
      }
    });
    const expectedActions = [
      { type: getGatewayDevices.pending.type },
      { type: actions.receivedDevice.type, payload: { id: gatewayDevice.id, gatewayIds: [] } },
      { type: getGatewayDevices.fulfilled.type }
    ];
    await store.dispatch(getGatewayDevices(defaultState.devices.byId.a1.id));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});

describe('device auth handling', () => {
  const deviceUpdateSuccessMessage = 'Device authorization status was updated successfully';
  it('should allow device auth information retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDeviceAuth.pending.type },
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDeviceAuth.fulfilled.type }
    ];
    await store.dispatch(getDeviceAuth(defaultState.devices.byId.a1.id));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should return device auth device as a promise result', async () => {
    const store = mockStore({ ...defaultState });
    const device = await store.dispatch(getDeviceAuth(defaultState.devices.byId.a1.id));
    expect(device).toBeDefined();
  });
  it('should allow single device auth updates', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: updateDeviceAuth.pending.type },
      { type: getDeviceAuth.pending.type },
      { type: getDevicesWithAuth.pending.type },
      { type: appActions.setSnackbar.type, payload: deviceUpdateSuccessMessage },
      defaultResults.receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDeviceAuth.fulfilled.type },
      { type: actions.maybeUpdateDevicesByStatus.type },
      { type: setDeviceListState.pending.type },
      { type: getDevicesByStatus.pending.type },
      { type: actions.setDeviceListState.type, payload: { deviceIds: [], isLoading: true, refreshTrigger: true } },
      {
        type: actions.receivedDevices.type,
        payload: { [defaultState.devices.byId.a1.id]: { ...defaultState.devices.byId.a1, updated_ts: inventoryDevice.updated_ts } }
      },
      acceptedDevices,
      { type: getDevicesWithAuth.pending.type },
      receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      {
        type: actions.setDeviceListState.type,
        payload: { deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id], total: 2, isLoading: false }
      },
      { type: setDeviceListState.fulfilled.type },
      { type: updateDeviceAuth.fulfilled.type }
    ];
    await store.dispatch(
      updateDeviceAuth({ deviceId: defaultState.devices.byId.a1.id, authId: defaultState.devices.byId.a1.auth_sets[0].id, status: DEVICE_STATES.pending })
    );
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow multiple device auth updates', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: updateDevicesAuth.pending.type },
      { type: getDevicesWithAuth.pending.type },
      { type: getDevicesWithAuth.fulfilled.type },
      { type: updateDeviceAuth.pending.type },
      { type: getDeviceAuth.pending.type },
      { type: getDevicesWithAuth.pending.type },
      { type: appActions.setSnackbar.type, payload: deviceUpdateSuccessMessage },
      defaultResults.receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDeviceAuth.fulfilled.type },
      { type: actions.maybeUpdateDevicesByStatus.type },
      { type: setDeviceListState.pending.type },
      { type: getDevicesByStatus.pending.type },
      { type: actions.setDeviceListState.type, payload: { deviceIds: [], total: 0, isLoading: true } },
      receivedExpectedDevice,
      defaultResults.acceptedDevices,
      { type: getDevicesWithAuth.pending.type },
      receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      {
        type: actions.setDeviceListState.type,
        payload: { deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id], total: 2, isLoading: false }
      },
      { type: setDeviceListState.fulfilled.type },
      { type: updateDeviceAuth.fulfilled.type },
      {
        type: appActions.setSnackbar.type,
        payload:
          '1 device was updated successfully. 1 device has more than one pending authset. Expand this device to individually adjust its authorization status. '
      },
      { type: updateDevicesAuth.fulfilled.type }
    ];
    await store.dispatch(updateDevicesAuth({ deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.c1.id], status: DEVICE_STATES.pending }));
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow preauthorizing devices', async () => {
    const store = mockStore({ ...defaultState });
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const expectedActions = [
      { type: preauthDevice.pending.type },
      { type: appActions.setSnackbar.type, payload: 'Device was successfully added to the preauthorization list' },
      { type: preauthDevice.fulfilled.type }
    ];
    await store.dispatch(
      preauthDevice({
        ...defaultState.devices.byId.a1.auth_sets[0],
        identity_data: { ...defaultState.devices.byId.a1.auth_sets[0].identity_data, mac: '12:34:56' },
        pubkey: 'test'
      })
    );
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should notify about duplicate device preauthorization attempts', async () => {
    const store = mockStore({ ...defaultState });
    await store
      .dispatch(preauthDevice(defaultState.devices.byId.a1.auth_sets[0]))
      .unwrap()
      .catch(({ message }) => expect(message).toContain('identity data set already exists'));
  });
  it('should allow single device auth set deletion', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: deleteAuthset.pending.type },
      { type: appActions.setSnackbar.type, payload: deviceUpdateSuccessMessage },
      { type: actions.maybeUpdateDevicesByStatus.type },
      ...defaultResults.postDeviceAuthActions,
      { type: deleteAuthset.fulfilled.type }
    ];
    await store.dispatch(deleteAuthset({ deviceId: defaultState.devices.byId.a1.id, authId: defaultState.devices.byId.a1.auth_sets[0].id }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow single device decomissioning', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: decommissionDevice.pending.type },
      { type: appActions.setSnackbar.type, payload: 'Device was decommissioned successfully' },
      { type: actions.maybeUpdateDevicesByStatus.type },
      ...defaultResults.postDeviceAuthActions,
      { type: decommissionDevice.fulfilled.type }
    ];
    await store.dispatch(decommissionDevice({ deviceId: defaultState.devices.byId.a1.id }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});

describe('static grouping related actions', () => {
  it('should allow retrieving static groups', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getGroups.pending.type },
      { type: actions.receivedGroups.type, payload: { testGroup: defaultState.devices.groups.byId.testGroup } },
      { type: getDevicesByStatus.pending.type },
      {
        type: actions.receivedDevices.type,
        payload: { [defaultState.devices.byId.a1.id]: { ...defaultState.devices.byId.a1, updated_ts: inventoryDevice.updated_ts } }
      },
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receiveDefaultDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      defaultResults.addedUngroupedGroup,
      { type: getGroups.fulfilled.type }
    ];
    await store.dispatch(getGroups());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow creating static groups', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'createdTestGroup';
    const expectedActions = [
      { type: addStaticGroup.pending.type },
      { type: addDevicesToGroup.pending.type },
      { type: actions.addToGroup.type, payload: { group: groupName, deviceIds: [defaultState.devices.byId.a1.id] } },
      { type: getGroups.pending.type },
      { type: actions.receivedGroups.type, payload: { testGroup: defaultState.devices.groups.byId.testGroup } },
      { type: getDevicesByStatus.pending.type },
      {
        type: actions.receivedDevices.type,
        payload: { [defaultState.devices.byId.a1.id]: { ...defaultState.devices.byId.a1, updated_ts: inventoryDevice.updated_ts } }
      },
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receiveDefaultDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      defaultResults.addedUngroupedGroup,
      { type: getGroups.fulfilled.type },
      { type: addDevicesToGroup.fulfilled.type },
      { type: actions.addGroup.type, payload: { groupName, group: { deviceIds: [], total: 0, filters: [] } } },
      { type: setDeviceListState.pending.type },
      { type: actions.setDeviceListState.type, payload: { ...defaultState.devices.deviceList, deviceIds: [], setOnly: true } },
      { type: getGroups.pending.type },
      { type: appActions.setSnackbar.type, payload: getGroupSuccessNotification(groupName) },
      { type: setDeviceListState.fulfilled.type },
      { type: actions.receivedGroups.type, payload: { testGroup: defaultState.devices.groups.byId.testGroup } },
      { type: getDevicesByStatus.pending.type },
      defaultResults.receiveDefaultDevice,
      { type: getDevicesWithAuth.pending.type },
      {
        type: actions.receivedDevices.type,
        payload: { [defaultState.devices.byId.a1.id]: { ...defaultState.devices.byId.a1, updated_ts: inventoryDevice.updated_ts } }
      },
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      defaultResults.addedUngroupedGroup,
      { type: getGroups.fulfilled.type },
      { type: addStaticGroup.fulfilled.type }
    ];
    await store.dispatch(addStaticGroup({ group: groupName, devices: [defaultState.devices.byId.a1] }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow extending static groups', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'createdTestGroup';
    const expectedActions = [
      { type: addDevicesToGroup.pending.type },
      { type: actions.addToGroup.type, payload: { group: groupName, deviceIds: [defaultState.devices.byId.b1.id] } },
      { type: addDevicesToGroup.fulfilled.type }
    ];
    await store.dispatch(addDevicesToGroup({ group: groupName, deviceIds: [defaultState.devices.byId.b1.id] }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow shrinking static groups', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'testGroup';
    const expectedActions = [
      { type: removeDevicesFromGroup.pending.type },
      { type: actions.removeFromGroup.type, payload: { group: groupName, deviceIds: [defaultState.devices.byId.b1.id] } },
      { type: appActions.setSnackbar.type, payload: 'The device was removed from the group' },
      { type: removeDevicesFromGroup.fulfilled.type }
    ];
    await store.dispatch(removeDevicesFromGroup({ group: groupName, deviceIds: [defaultState.devices.byId.b1.id] }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow removing static groups', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'testGroup';
    const expectedActions = [
      { type: removeStaticGroup.pending.type },
      { type: actions.removeGroup.type, payload: groupName },
      { type: getGroups.pending.type },
      { type: appActions.setSnackbar.type, payload: 'Group was removed successfully' },
      { type: actions.receivedGroups.type, payload: { testGroup: defaultState.devices.groups.byId.testGroup } },
      { type: getDevicesByStatus.pending.type },
      {
        type: actions.receivedDevices.type,
        payload: { [defaultState.devices.byId.a1.id]: { ...defaultState.devices.byId.a1, updated_ts: inventoryDevice.updated_ts } }
      },
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receiveDefaultDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      defaultResults.addedUngroupedGroup,
      { type: getGroups.fulfilled.type },
      { type: removeStaticGroup.fulfilled.type }
    ];
    await store.dispatch(removeStaticGroup(groupName));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow device retrieval for static groups', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'testGroup';
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { attributes, updated_ts, ...expectedDevice } = defaultState.devices.byId.a1;
    const expectedActions = [
      { type: getGroupDevices.pending.type },
      { type: getDevicesByStatus.pending.type },
      { type: actions.receivedDevices.type, payload: { [defaultState.devices.byId.a1.id]: { ...expectedDevice, attributes } } },
      {
        type: actions.setDevicesByStatus.type,
        payload: { deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id], status: DEVICE_STATES.accepted, total: 2 }
      },
      { type: getDevicesWithAuth.pending.type },
      { type: actions.receivedDevices.type, payload: { [expectedDevice.id]: { ...expectedDevice, updated_ts } } },
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      { type: getGroupDevices.fulfilled.type }
    ];
    await store.dispatch(getGroupDevices(groupName));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    const devicesById = storeActions.find(item => item.type === actions.receivedDevices.type).payload;
    expect(devicesById[defaultState.devices.byId.a1.id]).toBeTruthy();
    expect(new Date(devicesById[defaultState.devices.byId.a1.id].updated_ts).getTime()).toBeGreaterThanOrEqual(new Date(updated_ts).getTime());
  });
  it('should allow complete device retrieval for static groups', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'testGroup';
    const expectedActions = [
      { type: getAllGroupDevices.pending.type },
      defaultResults.receivedExpectedDevice,
      {
        type: actions.addGroup.type,
        payload: { group: { deviceIds: [defaultState.devices.byId.a1.id, defaultState.devices.byId.b1.id], total: 2 }, groupName }
      },
      { type: getAllGroupDevices.fulfilled.type }
    ];
    await store.dispatch(getAllGroupDevices(groupName));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});

describe('dynamic grouping related actions', () => {
  it('should allow retrieving dynamic groups', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [{ type: getDynamicGroups.pending.type }, defaultResults.receivedDynamicGroups, { type: getDynamicGroups.fulfilled.type }];
    await store.dispatch(getDynamicGroups());
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should allow creating dynamic groups', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'createdTestGroup';
    const expectedActions = [
      { type: addDynamicGroup.pending.type },
      {
        type: actions.addGroup.type,
        payload: { groupName, group: { filters: [{ key: 'group', operator: '$nin', scope: 'system', value: ['testGroup'] }] } }
      },
      { type: actions.setDeviceFilters.type, payload: [] },
      { type: appActions.setSnackbar.type, payload: getGroupSuccessNotification(groupName) },
      { type: getDynamicGroups.pending.type },
      defaultResults.receivedDynamicGroups,
      { type: getDynamicGroups.fulfilled.type },
      { type: addDynamicGroup.fulfilled.type }
    ];
    await store.dispatch(addDynamicGroup({ groupName, filterPredicates: [{ key: 'group', operator: '$nin', scope: 'system', value: ['testGroup'] }] }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow complete device retrieval for dynamic groups', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'testGroupDynamic';
    const expectedActions = [
      { type: getAllDynamicGroupDevices.pending.type },
      { type: actions.receivedDevices.type, payload: {} },
      { type: actions.addGroup.type, payload: { group: { deviceIds: [], total: 0 }, groupName } },
      { type: getAllDynamicGroupDevices.fulfilled.type }
    ];
    await store.dispatch(getAllDynamicGroupDevices(groupName));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow dynamic group updates', async () => {
    const groupName = 'testGroupDynamic';
    const store = mockStore({
      ...defaultState,
      devices: {
        ...defaultState.devices,
        groups: {
          ...defaultState.devices.groups,
          selectedGroup: groupName
        }
      }
    });
    const expectedActions = [
      { type: updateDynamicGroup.pending.type },
      { type: addDynamicGroup.pending.type },
      { type: actions.addGroup.type, payload: { groupName, group: { filters: [] } } },
      { type: actions.setDeviceFilters.type, payload: defaultState.devices.groups.byId.testGroupDynamic.filters },
      { type: appActions.setSnackbar.type, payload: groupUpdateSuccessMessage },
      { type: getDynamicGroups.pending.type },
      defaultResults.receivedDynamicGroups,
      { type: getDynamicGroups.fulfilled.type },
      { type: addDynamicGroup.fulfilled.type },
      { type: updateDynamicGroup.fulfilled.type }
    ];
    await store.dispatch(updateDynamicGroup({ groupName, filterPredicates: [] }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow removing dynamic groups', async () => {
    const store = mockStore({ ...defaultState });
    const groupName = 'testGroupDynamic';
    const expectedActions = [
      { type: removeDynamicGroup.pending.type },
      { type: actions.removeGroup.type, payload: groupName },
      { type: appActions.setSnackbar.type, payload: 'Group was removed successfully' },
      { type: removeDynamicGroup.fulfilled.type }
    ];
    await store.dispatch(removeDynamicGroup(groupName));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});

describe('device retrieval ', () => {
  it('should allow single device retrieval from inventory', async () => {
    const store = mockStore({
      ...defaultState
    });
    const { attributes, id } = defaultState.devices.byId.a1;
    const expectedActions = [
      { type: getDeviceById.pending.type },
      { type: actions.receivedDevice.type, payload: { attributes, id } },
      { type: getDeviceById.fulfilled.type }
    ];
    await store.dispatch(getDeviceById(defaultState.devices.byId.a1.id));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow single device retrieval from detailed sources', async () => {
    const store = mockStore({
      ...defaultState,
      app: { ...defaultState.app, features: { ...defaultState.app.features, hasDeviceConnect: true } },
      organization: { ...defaultState.organization, addons: [], externalDeviceIntegrations: [{ ...EXTERNAL_PROVIDER['iot-hub'], id: 'test' }] }
    });
    const { attributes, updated_ts, id, ...expectedDevice } = defaultState.devices.byId.a1;
    const expectedActions = [
      { type: getDeviceInfo.pending.type },
      { type: getDeviceAuth.pending.type },
      { type: getDevicesWithAuth.pending.type },
      { type: getDeviceTwin.pending.type },
      { type: getDeviceById.pending.type },
      { type: getDeviceConnect.pending.type },
      { type: actions.receivedDevices.type, payload: { [id]: { ...expectedDevice, id } } },
      { type: actions.receivedDevice.type, payload: { attributes, id } },
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDeviceById.fulfilled.type },
      { type: getDeviceAuth.fulfilled.type },
      { type: actions.receivedDevice.type, payload: { connect_status: 'connected', connect_updated_ts: updated_ts, id } },
      { type: actions.receivedDevice.type, payload: expectedDevice },
      { type: getDeviceConnect.fulfilled.type },
      { type: getDeviceTwin.fulfilled.type },
      { type: getDeviceInfo.fulfilled.type }
    ];
    await store.dispatch(getDeviceInfo(defaultState.devices.byId.a1.id));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow retrieving multiple devices by status', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDevicesByStatus.pending.type },
      defaultResults.receivedExpectedDevice,
      defaultResults.acceptedDevices,
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type }
    ];
    await store.dispatch(getDevicesByStatus({ status: DEVICE_STATES.accepted }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow retrieving multiple devices by status and select if requested', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDevicesByStatus.pending.type },
      defaultResults.receivedExpectedDevice,
      {
        type: actions.setDevicesByStatus.type,
        payload: { deviceIds: [defaultState.devices.byId.a1.id], status: DEVICE_STATES.accepted, total: defaultState.devices.byStatus.accepted.total }
      },
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type }
    ];
    await store.dispatch(getDevicesByStatus({ status: DEVICE_STATES.accepted, perPage: 1, shouldSelectDevices: true }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow retrieving devices based on devicelist state', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: setDeviceListState.pending.type },
      { type: getDevicesByStatus.pending.type },
      { type: actions.setDeviceListState.type, payload: { ...defaultState.devices.deviceList, perPage: 2, deviceIds: [], isLoading: true } },
      defaultResults.receivedExpectedDevice,
      defaultResults.acceptedDevices,
      { type: getDevicesWithAuth.pending.type },
      defaultResults.receivedExpectedDevice,
      { type: getDevicesWithAuth.fulfilled.type },
      { type: getDevicesByStatus.fulfilled.type },
      // the following perPage setting should be 2 as well, but the test backend seems to respond too fast for the state change to propagate
      defaultResults.defaultDeviceListState,
      { type: setDeviceListState.fulfilled.type }
    ];
    await store.dispatch(setDeviceListState({ page: 1, perPage: 2, refreshTrigger: true }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow retrieving all devices per status', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getAllDevicesByStatus.pending.type },
      defaultResults.receivedExpectedDevice,
      defaultResults.acceptedDevices,
      { type: deriveInactiveDevices.pending.type },
      { type: actions.setInactiveDevices.type, payload: { activeDeviceTotal: 0, inactiveDeviceTotal: 2 } },
      { type: deriveReportsData.pending.type },
      { type: actions.setDeviceReports.type, payload: [{ items: [{ count: 2, key: 'undefined' }], otherCount: 0, total: 2 }] },
      { type: deriveInactiveDevices.fulfilled.type },
      { type: deriveReportsData.fulfilled.type },
      { type: getAllDevicesByStatus.fulfilled.type }
    ];
    await store.dispatch(getAllDevicesByStatus(DEVICE_STATES.accepted));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow retrieving devices per status and their auth data', async () => {
    const store = mockStore({ ...defaultState });
    const {
      a1: { attributes: attributes1, ...expectedDevice1 }, // eslint-disable-line @typescript-eslint/no-unused-vars
      b1: { attributes: attributes2, auth_sets, ...expectedDevice2 } // eslint-disable-line @typescript-eslint/no-unused-vars
    } = defaultState.devices.byId;
    const expectedActions = [
      { type: getDevicesWithAuth.pending.type },
      { type: actions.receivedDevices.type, payload: { [expectedDevice1.id]: expectedDevice1, [expectedDevice2.id]: expectedDevice2 } },
      { type: getDevicesWithAuth.fulfilled.type }
    ];
    await store.dispatch(getDevicesWithAuth([defaultState.devices.byId.a1, defaultState.devices.byId.b1]));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});

const deviceConfig = {
  configured: { aNumber: 42, something: 'else', test: true },
  reported: { aNumber: 42, something: 'else', test: true },
  updated_ts: defaultState.devices.byId.a1.updated_ts,
  reported_ts: '2019-01-01T09:25:01.000Z'
};

describe('device config ', () => {
  it('should allow single device config retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDeviceConfig.pending.type },
      { type: actions.receivedDevice.type, payload: { config: deviceConfig, id: defaultState.devices.byId.a1.id } },
      { type: getDeviceConfig.fulfilled.type }
    ];
    await store.dispatch(getDeviceConfig(defaultState.devices.byId.a1.id));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should not have a problem with unknown devices on config retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [{ type: getDeviceConfig.pending.type }, { type: getDeviceConfig.fulfilled.type }];
    await store.dispatch(getDeviceConfig('testId'));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });

  it('should allow single device config update', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: setDeviceConfig.pending.type },
      { type: getDeviceConfig.pending.type },
      { type: actions.receivedDevice.type, payload: { config: deviceConfig, id: defaultState.devices.byId.a1.id } },
      { type: getDeviceConfig.fulfilled.type },
      { type: setDeviceConfig.fulfilled.type }
    ];
    await store.dispatch(setDeviceConfig({ deviceId: defaultState.devices.byId.a1.id, config: { something: 'asdl' } }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow single device config deployment', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: applyDeviceConfig.pending.type },
      { type: actions.receivedDevice.type, payload: { ...defaultState.devices.byId.a1, config: { deployment_id: '' } } },
      { type: getSingleDeployment.type },
      { type: deploymentActions.receivedDeployment.type, payload: { ...defaultState.deployments.byId.d1, id: 'config1', created: '2019-01-01T09:25:01.000Z' } },
      { type: getSingleDeployment.type },
      { type: applyDeviceConfig.fulfilled.type }
    ];
    const result = store.dispatch(applyDeviceConfig({ deviceId: defaultState.devices.byId.a1.id, config: { something: 'asdl' } }));
    await act(async () => jest.runAllTicks());
    result.then(() => {
      const storeActions = store.getActions();
      expect(storeActions.length).toEqual(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should allow setting device tags', async () => {
    const store = mockStore({ ...defaultState });
    const { attributes, id } = defaultState.devices.byId.a1;
    const expectedActions = [
      { type: setDeviceTags.pending.type },
      { type: getDeviceById.pending.type },
      { type: actions.receivedDevice.type, payload: { attributes, id } },
      { type: getDeviceById.fulfilled.type },
      { type: actions.receivedDevice.type, payload: { id, tags: { something: 'asdl' } } },
      { type: appActions.setSnackbar.type, payload: 'Device name changed' },
      { type: setDeviceTags.fulfilled.type }
    ];
    await store.dispatch(setDeviceTags({ deviceId: defaultState.devices.byId.a1.id, tags: { something: 'asdl' } }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});

describe('troubleshooting related actions', () => {
  it('should allow session info retrieval', async () => {
    const store = mockStore({ ...defaultState });
    const endDate = '2019-01-01T12:16:22.667Z';
    const sessionId = 'abd313a8-ee88-48ab-9c99-fbcd80048e6e';
    const result = await store
      .dispatch(getSessionDetails({ sessionId, deviceId: defaultState.devices.byId.a1.id, userId: defaultState.users.currentUser, endDate }))
      .unwrap();

    expect(result).toMatchObject({ start: new Date(endDate), end: new Date(endDate) });
  });

  it('should allow device file transfers', async () => {
    const store = mockStore({ ...defaultState });
    const link = await store.dispatch(getDeviceFileDownloadLink({ deviceId: 'aDeviceId', path: '/tmp/file' })).unwrap();
    expect(link).toBe('http://localhost/api/management/v1/deviceconnect/devices/aDeviceId/download?path=%2Ftmp%2Ffile');
    const expectedActions = [
      { type: getDeviceFileDownloadLink.pending.type },
      { type: getDeviceFileDownloadLink.fulfilled.type },
      { type: deviceFileUpload.pending.type },
      { type: appActions.setSnackbar.type, payload: 'Uploading file' },
      {
        type: appActions.initUpload.type,
        payload: { id: 'mock-uuid', upload: { cancelSource: mockAbortController, uploadProgress: 0 } }
      },
      { type: appActions.uploadProgress.type, payload: { id: 'mock-uuid', progress: 100 } },
      { type: appActions.setSnackbar.type, payload: 'Upload successful' },
      { type: appActions.cleanUpUpload.type, payload: 'mock-uuid' },
      { type: deviceFileUpload.fulfilled.type }
    ];
    await store.dispatch(deviceFileUpload({ deviceId: defaultState.devices.byId.a1.id, path: '/tmp/file', file: 'file' }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});

describe('device twin related actions', () => {
  it('should allow retrieving twin data from azure', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: getDeviceTwin.pending.type },
      { type: actions.receivedDevice.type, payload: defaultState.devices.byId.a1 },
      { type: getDeviceTwin.fulfilled.type }
    ];
    await store.dispatch(getDeviceTwin({ deviceId: defaultState.devices.byId.a1.id, integration: EXTERNAL_PROVIDER['iot-hub'] }));
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
  it('should allow configuring twin data on azure', async () => {
    const store = mockStore({ ...defaultState });
    const expectedActions = [
      { type: setDeviceTwin.pending.type },
      { type: actions.receivedDevice.type, payload: defaultState.devices.byId.a1 },
      { type: setDeviceTwin.fulfilled.type }
    ];
    await store.dispatch(
      setDeviceTwin({
        deviceId: defaultState.devices.byId.a1.id,
        integration: EXTERNAL_PROVIDER['iot-hub'],
        settings: { something: 'asdl' }
      })
    );
    const storeActions = store.getActions();
    expect(storeActions.length).toEqual(expectedActions.length);
    expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
  });
});
