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
import { combineReducers, configureStore } from '@reduxjs/toolkit';

import { defaultState } from '../../../tests/mockData';
import { SET_SNACKBAR, UPLOAD_PROGRESS } from '../constants/appConstants';
import { RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS } from '../constants/organizationConstants';
import { USER_LOGOUT } from '../constants/userConstants';
import appReducer from './appReducer';
import deploymentReducer from './deploymentReducer';
import deviceReducer from './deviceReducer';
import monitorReducer from './monitorReducer';
import onboardingReducer from './onboardingReducer';
import organizationReducer from './organizationReducer';
import releaseReducer from './releaseReducer';
import userReducer from './userReducer';

const rootReducer = combineReducers({
  app: appReducer,
  devices: deviceReducer,
  deployments: deploymentReducer,
  monitor: monitorReducer,
  onboarding: onboardingReducer,
  organization: organizationReducer,
  releases: releaseReducer,
  users: userReducer
});

export const sessionReducer = (state, action) => {
  if (action.type === USER_LOGOUT) {
    state = undefined;
  }
  return rootReducer(state, action);
};

export const getConfiguredStore = (options = {}) => {
  const { preloadedState = { ...defaultState }, ...config } = options;
  return configureStore({
    ...config,
    preloadedState,
    reducer: sessionReducer,
    middleware: getDefaultMiddleware =>
      getDefaultMiddleware({
        immutableCheck: {
          ignoredPaths: ['app.uploadsById']
        },
        serializableCheck: {
          ignoredActions: [RECEIVE_EXTERNAL_DEVICE_INTEGRATIONS, SET_SNACKBAR, UPLOAD_PROGRESS],
          ignoredActionPaths: ['uploads', 'snackbar'],
          ignoredPaths: ['app.uploadsById', 'app.snackbar', 'organization.externalDeviceIntegrations']
        }
      })
  });
};

export default getConfiguredStore({ preloadedState: {} });
