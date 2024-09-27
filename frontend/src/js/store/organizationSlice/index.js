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
import { DEVICE_LIST_DEFAULTS, SORTING_OPTIONS } from '@northern.tech/store/commonConstants';
import { createSlice } from '@reduxjs/toolkit';

export const sliceName = 'organization';

export const initialState = {
  card: {
    last4: '',
    expiration: { month: 1, year: 2020 },
    brand: ''
  },
  intentId: null,
  organization: {
    // id, name, status, tenant_token, plan
  },
  auditlog: {
    events: [],
    selectionState: {
      ...DEVICE_LIST_DEFAULTS,
      detail: undefined,
      endDate: undefined,
      selectedIssue: undefined,
      sort: { direction: SORTING_OPTIONS.desc },
      startDate: undefined,
      total: 0,
      type: undefined,
      user: undefined
    }
  },
  externalDeviceIntegrations: [
    // { <connection_string|x509|...>, id, provider }
  ],
  ssoConfigs: [],
  webhooks: {
    // [id]: { events: [] }
    // for now:
    events: [],
    eventsTotal: 0
  }
};

export const organizationSlice = createSlice({
  name: sliceName,
  initialState,
  reducers: {
    receiveAuditLogs: (state, action) => {
      const { events, total } = action.payload;
      state.auditlog.events = events;
      state.auditlog.selectionState.total = total;
    },
    setAuditLogState: (state, action) => {
      state.auditlog.selectionState = {
        ...state.auditlog.selectionState,
        ...action.payload
      };
    },
    receiveCurrentCard: (state, action) => {
      state.card = action.payload;
    },
    receiveSetupIntent: (state, action) => {
      state.intentId = action.payload;
    },
    setOrganization: (state, action) => {
      state.organization = action.payload;
    },
    receiveExternalDeviceIntegrations: (state, action) => {
      state.externalDeviceIntegrations = action.payload;
    },
    receiveSsoConfigs: (state, action) => {
      state.ssoConfigs = action.payload;
    },
    receiveWebhookEvents: (state, action) => {
      const { value, total } = action.payload;
      state.webhooks.events = value;
      state.webhooks.eventsTotal = total;
    }
  }
});

export const actions = organizationSlice.actions;
export default organizationSlice.reducer;
