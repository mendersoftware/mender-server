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
import * as UserConstants from '../constants/userConstants';

export const initialState = {
  byId: {},
  currentUser: null,
  customColumns: [],
  currentSession: {
    // { token: window.localStorage.getItem('JWT'), expiresAt: '2023-01-01T00:15:00.000Z' | undefined }, // expiresAt depending on the stay logged in setting
  },
  qrCode: null,
  globalSettings: {
    id_attribute: undefined,
    previousFilters: [],
    previousPhases: [],
    retries: 0
  },
  permissionSetsById: {
    ...UserConstants.defaultPermissionSets
  },
  rolesById: {
    ...UserConstants.rolesById
  },
  showConnectDeviceDialog: false,
  showStartupNotification: false,
  tooltips: {
    byId: {
      // <id>: { readState: <read|unread> } // this object is getting enhanced by the tooltip texts in the app constants
    }
  },
  userSettings: {
    columnSelection: [],
    onboarding: {}
  }
};

const userReducer = (state = initialState, action) => {
  switch (action.type) {
    case UserConstants.RECEIVED_QR_CODE:
      return {
        ...state,
        qrCode: action.value
      };
    case UserConstants.SUCCESSFULLY_LOGGED_IN:
      return {
        ...state,
        currentSession: action.value
      };
    case UserConstants.RECEIVED_USER_LIST:
      return {
        ...state,
        byId: { ...action.users }
      };
    case UserConstants.RECEIVED_ACTIVATION_CODE:
      return {
        ...state,
        activationCode: action.code
      };
    case UserConstants.RECEIVED_USER:
      return {
        ...state,
        byId: {
          ...state.byId,
          [action.user.id]: {
            ...action.user
          }
        },
        currentUser: action.user.id
      };
    case UserConstants.CREATED_USER:
      // the new user gets a 0 as id, since this will be overwritten by the retrieved userlist anyway + there is no way to know the id before
      return {
        ...state,
        byId: {
          ...state.byId,
          0: action.user
        }
      };
    case UserConstants.REMOVED_USER: {
      // eslint-disable-next-line no-unused-vars
      const { [action.userId]: removedUser, ...byId } = state.byId;
      return {
        ...state,
        byId,
        currentUser: state.currentUser === action.userId ? null : state.currentUser
      };
    }
    case UserConstants.UPDATED_USER:
      return {
        ...state,
        byId: {
          ...state.byId,
          [action.userId]: {
            ...state.byId[action.userId],
            ...action.user
          }
        }
      };
    case UserConstants.RECEIVED_PERMISSION_SETS:
      return {
        ...state,
        permissionSetsById: action.value
      };
    case UserConstants.RECEIVED_ROLES:
    case UserConstants.REMOVED_ROLE:
      return {
        ...state,
        rolesById: action.value
      };
    case UserConstants.CREATED_ROLE:
    case UserConstants.UPDATED_ROLE:
      return {
        ...state,
        rolesById: {
          ...state.rolesById,
          [action.roleId]: {
            ...state.rolesById[action.roleId],
            ...action.role
          }
        }
      };
    case UserConstants.SET_CUSTOM_COLUMNS:
      return {
        ...state,
        customColumns: action.value
      };
    case UserConstants.SET_GLOBAL_SETTINGS:
      return {
        ...state,
        settingsInitialized: true,
        globalSettings: {
          ...state.globalSettings,
          ...action.settings
        }
      };
    case UserConstants.SET_USER_SETTINGS:
      return {
        ...state,
        userSettingsInitialized: true,
        userSettings: {
          ...state.userSettings,
          ...action.settings
        }
      };
    case UserConstants.SET_SHOW_CONNECT_DEVICE:
      return {
        ...state,
        showConnectDeviceDialog: action.show
      };
    case UserConstants.SET_SHOW_STARTUP_NOTIFICATION:
      return {
        ...state,
        showStartupNotification: action.value
      };
    case UserConstants.SET_TOOLTIP_STATE:
      return {
        ...state,
        tooltips: {
          ...state.tooltips,
          byId: {
            ...state.tooltips.byId,
            [action.id]: action.value
          }
        }
      };
    case UserConstants.SET_TOOLTIPS_STATE:
      return {
        ...state,
        tooltips: {
          ...state.tooltips,
          byId: action.value
        }
      };
    default:
      return state;
  }
};

export default userReducer;
