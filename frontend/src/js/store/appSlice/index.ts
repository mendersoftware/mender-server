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
import { SORTING_OPTIONS } from '@northern.tech/store/constants';
import { createSlice } from '@reduxjs/toolkit';

export const sliceName = 'app';

const getYesterday = () => {
  const today = new Date();
  today.setDate(today.getDate() - 1);
  return today.toISOString();
};

export const initialState = {
  cancelSource: undefined,
  demoArtifactLink: 'https://dgsbl4vditpls.cloudfront.net/mender-demo-artifact.mender',
  hostAddress: null,
  snackbar: {
    open: false,
    message: '',
    maxWidth: 900,
    autoHideDuration: undefined,
    action: undefined,
    children: undefined,
    onClick: undefined,
    onClose: undefined
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
      direction: SORTING_OPTIONS.desc
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

export const appSlice = createSlice({
  name: sliceName,
  initialState,
  reducers: {
    setFeatures: (state, action) => {
      state.features = {
        ...state.features,
        ...action.payload
      };
    },
    setSnackbar: (state, { payload }) => {
      let { message, autoHideDuration, action, children, onClick, onClose } = payload;
      if (typeof payload === 'string' || payload instanceof String) {
        message = payload;
      }
      state.snackbar = {
        open: message ? true : false,
        message,
        maxWidth: 900,
        autoHideDuration,
        action,
        children,
        onClick,
        onClose
      };
    },
    setFirstLoginAfterSignup: (state, action) => {
      state.firstLoginAfterSignup = action.payload;
    },
    setAnnouncement: (state, action) => {
      state.hostedAnnouncement = action.payload;
    },
    setSearchState: (state, action) => {
      state.searchState = {
        ...state.searchState,
        ...action.payload
      };
    },
    setOfflineThreshold: (state, action) => {
      state.offlineThreshold = action.payload;
    },
    initUpload: (state, action) => {
      const { id, upload } = action.payload;
      state.uploadsById[id] = upload;
    },
    uploadProgress: (state, action) => {
      const { id, progress } = action.payload;
      state.uploadsById[id] = {
        ...state.uploadsById[id],
        progress
      };
    },
    cleanUpUpload: (state, action) => {
      // eslint-disable-next-line no-unused-vars
      const { [action.payload]: current, ...remainder } = state.uploadsById;
      state.uploadsById = remainder;
    },
    setVersionInformation: (state, action) => {
      state.versionInformation = {
        ...state.versionInformation,
        ...action.payload
      };
    },
    setEnvironmentData: (state, action) => {
      return { ...state, ...action.payload };
    }
  }
});

export const actions = appSlice.actions;
export default appSlice.reducer;
