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
import configureMockStore from 'redux-mock-store';
import { thunk } from 'redux-thunk';

import { defaultState } from '../../../tests/mockData';
import * as AppConstants from '../constants/appConstants';
import { DEVICE_ISSUE_OPTIONS } from '../constants/deviceConstants';
import * as MonitorConstants from '../constants/monitorConstants';
import { changeNotificationSetting, getDeviceAlerts, getDeviceMonitorConfig, getIssueCountsByType, getLatestDeviceAlerts } from './monitorActions';

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

/* eslint-disable sonarjs/no-identical-functions */
describe('monitor actions', () => {
  it('should handle device based alert retrieval', async () => {
    const store = mockStore({ ...defaultState });
    expect(store.getActions()).toHaveLength(0);
    const expectedActions = [
      {
        type: MonitorConstants.RECEIVE_DEVICE_ALERTS,
        deviceId: defaultState.devices.byId.a1.id,
        alerts: []
      },
      { type: MonitorConstants.SET_ALERT_LIST_STATE, value: { page: 1, perPage: 20, total: 1 } }
    ];
    const request = store.dispatch(getDeviceAlerts(defaultState.devices.byId.a1.id));
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
      {
        type: MonitorConstants.RECEIVE_LATEST_DEVICE_ALERTS,
        deviceId: defaultState.devices.byId.a1.id,
        alerts: []
      }
    ];
    const request = store.dispatch(getLatestDeviceAlerts(defaultState.devices.byId.a1.id));
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
      {
        type: MonitorConstants.RECEIVE_DEVICE_ISSUE_COUNTS,
        issueType: DEVICE_ISSUE_OPTIONS.monitoring.key,
        counts: { filtered: 4, total: 4 }
      }
    ];
    const request = store.dispatch(getIssueCountsByType(DEVICE_ISSUE_OPTIONS.monitoring.key));
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
      {
        type: MonitorConstants.RECEIVE_DEVICE_MONITOR_CONFIG,
        device: { id: defaultState.devices.byId.a1.id, monitors: [{ something: 'here' }] }
      }
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
      {
        type: MonitorConstants.CHANGE_ALERT_CHANNEL,
        channel: 'email',
        enabled: false
      },
      {
        type: AppConstants.SET_SNACKBAR,
        snackbar: {
          message: 'Successfully disabled email alerts'
        }
      }
    ];
    const request = store.dispatch(changeNotificationSetting(false));
    expect(request).resolves.toBeTruthy();
    await request.then(() => {
      const storeActions = store.getActions();
      expect(storeActions).toHaveLength(expectedActions.length);
      expectedActions.map((action, index) => expect(storeActions[index]).toMatchObject(action));
    });
  });
});
