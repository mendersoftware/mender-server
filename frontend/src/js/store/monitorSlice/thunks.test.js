// Copyright 2021 Northern.tech AS
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
import { DEVICE_ISSUE_OPTIONS } from '@northern.tech/store/commonConstants';
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { actions } from '.';
import { defaultState } from '../../../../tests/mockData';
import { actions as appActions } from '../appSlice';
import { actions as deviceActions } from '../devicesSlice';
import { changeNotificationSetting, getDeviceAlerts, getDeviceMonitorConfig, getIssueCountsByType, getLatestDeviceAlerts } from './thunks';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

/* eslint-disable sonarjs/no-identical-functions */
describe('monitor actions', () => {
  it('should handle device based alert retrieval', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getDeviceAlerts.pending.type },
      { type: actions.receiveDeviceAlerts.type, payload: { deviceId: defaultState.devices.byId.a1.id, alerts: [] } },
      { type: actions.setAlertListState.type, payload: { total: 1 } },
      { type: getDeviceAlerts.fulfilled.type }
    ];
    const request = store.dispatch(getDeviceAlerts({ id: defaultState.devices.byId.a1.id }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should handle device based latest alert retrieval', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getLatestDeviceAlerts.pending.type },
      { type: actions.receiveLatestDeviceAlerts.type, payload: { deviceId: defaultState.devices.byId.a1.id, alerts: [] } },
      { type: getLatestDeviceAlerts.fulfilled.type }
    ];
    const request = store.dispatch(getLatestDeviceAlerts({ id: defaultState.devices.byId.a1.id }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should handle device issue count retrieval', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getIssueCountsByType.pending.type },
      { type: actions.receiveDeviceIssueCounts.type, payload: { issueType: DEVICE_ISSUE_OPTIONS.monitoring.key, counts: { filtered: 4, total: 4 } } },
      { type: getIssueCountsByType.fulfilled.type }
    ];
    const request = store.dispatch(getIssueCountsByType({ type: DEVICE_ISSUE_OPTIONS.monitoring.key }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should handle device monitor config retrieval', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: getDeviceMonitorConfig.pending.type },
      { type: deviceActions.receivedDevice.type, payload: { id: defaultState.devices.byId.a1.id, monitors: [{ something: 'here' }] } },
      { type: getDeviceMonitorConfig.fulfilled.type }
    ];
    const request = store.dispatch(getDeviceMonitorConfig(defaultState.devices.byId.a1.id));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
  it('should handle changes to alert notification settings', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      { type: changeNotificationSetting.pending.type },
      { type: actions.changeAlertChannel.type, payload: { channel: 'email', enabled: false } },
      { type: appActions.setSnackbar.type, payload: 'Successfully disabled email alerts' },
      { type: changeNotificationSetting.fulfilled.type }
    ];
    const request = store.dispatch(changeNotificationSetting({ enabled: false }));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
});
