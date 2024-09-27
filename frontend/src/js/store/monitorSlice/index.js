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
import { DEVICE_ISSUE_OPTIONS, DEVICE_LIST_DEFAULTS } from '@northern.tech/store/commonConstants';
import { createSlice } from '@reduxjs/toolkit';

import { alertChannels } from './constants';

export const sliceName = 'monitor';

export const initialState = {
  alerts: {
    alertList: { ...DEVICE_LIST_DEFAULTS, total: 0 },
    byDeviceId: {}
  },
  issueCounts: {
    byType: Object.values(DEVICE_ISSUE_OPTIONS).reduce((accu, { key }) => ({ ...accu, [key]: { filtered: 0, total: 0 } }), {})
  },
  settings: {
    global: {
      channels: {
        ...Object.keys(alertChannels).reduce((accu, item) => ({ ...accu, [item]: { enabled: true } }), {})
      }
    }
  }
};

export const monitorSlice = createSlice({
  name: sliceName,
  initialState,
  reducers: {
    changeAlertChannel: (state, action) => {
      const { channel, enabled } = action.payload;
      state.settings.global.channels[channel] = { enabled };
    },
    receiveDeviceAlerts: (state, action) => {
      const { deviceId, alerts } = action.payload;
      state.alerts.byDeviceId[deviceId] = { alerts };
    },
    receiveLatestDeviceAlerts: (state, action) => {
      const { deviceId, alerts } = action.payload;
      state.alerts.byDeviceId[deviceId] = { ...state.alerts.byDeviceId[deviceId], latest: alerts };
    },
    receiveDeviceIssueCounts: (state, action) => {
      const { issueType, counts } = action.payload;
      state.issueCounts.byType[issueType] = counts;
    },
    setAlertListState: (state, action) => {
      state.alerts.alertList = { ...state.alerts.alertList, ...action.payload };
    }
  }
});

export const actions = monitorSlice.actions;
export default monitorSlice.reducer;
