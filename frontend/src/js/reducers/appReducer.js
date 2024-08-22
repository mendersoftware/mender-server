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
import * as AppConstants from '../constants/appConstants';
import * as UserConstants from '../constants/userConstants';

const getYesterday = () => {
  const today = new Date();
  today.setDate(today.getDate() - 1);
  return today.toISOString();
};

export const initialState = {
  cancelSource: undefined,
  hostAddress: null,
  snackbar: {
    open: false,
    message: ''
  },
  // return boolean rather than organization details
  features: {
    hasAuditlogs: false,
    hasDeltaProgress: false,
    hasMultitenancy: false,
    hasDeviceConfig: false,
    hasDeviceConnect: false,
    hasMonitor: false,
    hasReporting: false,
    isDemoMode: false,
    isHosted: false,
    isEnterprise: false
  },
  firstLoginAfterSignup: false,
  hostedAnnouncement: '',
  docsVersion: '',
  recaptchaSiteKey: '',
  searchState: {
    deviceIds: [],
    searchTerm: '',
    searchTotal: 0,
    sort: {
      direction: AppConstants.SORTING_OPTIONS.desc
      // key: null,
      // scope: null
    }
  },
  stripeAPIKey: '',
  trackerCode: '',
  uploadsById: {
    // id: { uploading: false, uploadProgress: 0, cancelSource: undefined }
  },
  newThreshold: getYesterday(),
  offlineThreshold: getYesterday(),
  versionInformation: {
    Integration: '',
    'Mender-Client': '',
    'Mender-Artifact': '',
    'Meta-Mender': '',
    Deployments: '',
    Deviceauth: '',
    Inventory: '',
    GUI: 'latest'
  },
  yesterday: undefined
};

// exclude 'pendings-redirect' since this is expected to persist refreshes - the rest should be better to be redone
const keys = ['sessionDeploymentChecker', UserConstants.settingsKeys.initialized];
const resetEnvironment = () => {
  keys.map(key => window.sessionStorage.removeItem(key));
};

resetEnvironment();

const appReducer = (state = initialState, action) => {
  switch (action.type) {
    case AppConstants.SET_FEATURES:
      return {
        ...state,
        features: {
          ...state.features,
          ...action.value
        }
      };
    case AppConstants.SET_SNACKBAR:
      return {
        ...state,
        snackbar: action.snackbar
      };
    case AppConstants.SET_FIRST_LOGIN_AFTER_SIGNUP:
      return {
        ...state,
        firstLoginAfterSignup: action.firstLoginAfterSignup
      };
    case AppConstants.SET_ANNOUNCEMENT:
      return {
        ...state,
        hostedAnnouncement: action.announcement
      };
    case AppConstants.SET_SEARCH_STATE:
      return {
        ...state,
        searchState: action.state
      };
    case AppConstants.SET_OFFLINE_THRESHOLD:
      return {
        ...state,
        offlineThreshold: action.value
      };
    case AppConstants.UPLOAD_PROGRESS:
      return {
        ...state,
        uploadsById: action.uploads
      };
    case AppConstants.SET_VERSION_INFORMATION:
      return {
        ...state,
        docsVersion: action.docsVersion,
        versionInformation: action.value
      };
    case AppConstants.SET_ENVIRONMENT_DATA:
      return {
        ...state,
        ...action.value
      };
    default:
      return state;
  }
};

export default appReducer;
