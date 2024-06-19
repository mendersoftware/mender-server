// Copyright 2020 Northern.tech AS
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
import { SORTING_OPTIONS } from '../constants/appConstants';
import { DEVICE_LIST_DEFAULTS } from '../constants/deviceConstants';
import {
  RECEIVE_AUDIT_LOGS,
  RECEIVE_CURRENT_CARD,
  RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS,
  RECEIVE_SETUP_INTENT,
  RECEIVE_SSO_CONFIGS,
  RECEIVE_WEBHOOK_EVENTS,
  SET_AUDITLOG_STATE,
  SET_ORGANIZATION
} from '../constants/organizationConstants';

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

const organizationReducer = (state = initialState, action) => {
  switch (action.type) {
    case RECEIVE_AUDIT_LOGS:
      return {
        ...state,
        auditlog: {
          ...state.auditlog,
          events: action.events,
          selectionState: {
            ...state.auditlog.selectionState,
            total: action.total
          }
        }
      };
    case SET_AUDITLOG_STATE:
      return {
        ...state,
        auditlog: {
          ...state.auditlog,
          selectionState: action.state
        }
      };
    case RECEIVE_CURRENT_CARD:
      return {
        ...state,
        card: action.card
      };
    case RECEIVE_SETUP_INTENT:
      return {
        ...state,
        intentId: action.intentId
      };
    case SET_ORGANIZATION:
      return {
        ...state,
        organization: {
          ...action.organization
        }
      };
    case RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS:
      return {
        ...state,
        externalDeviceIntegrations: action.value
      };
    case RECEIVE_SSO_CONFIGS:
      return {
        ...state,
        ssoConfigs: action.value
      };
    case RECEIVE_WEBHOOK_EVENTS:
      return {
        ...state,
        webhooks: {
          ...state.webhooks,
          events: action.value,
          eventsTotal: action.total
        }
      };
    default:
      return state;
  }
};

export default organizationReducer;
