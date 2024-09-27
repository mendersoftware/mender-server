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
// @ts-nocheck
import { DEVICE_ONLINE_CUTOFF, defaultIdAttribute } from '@northern.tech/store/constants';
import { isDarkMode } from '@northern.tech/store/utils';
import { createSelector } from '@reduxjs/toolkit';

import { READ_STATES, twoFAStates } from './constants';

export const getRolesById = state => state.users.rolesById;
export const getTooltipsById = state => state.users.tooltips.byId;
export const getGlobalSettings = state => state.users.globalSettings;

const getCurrentUserId = state => state.users.currentUser;
const getUsersById = state => state.users.byId;
export const getCurrentUser = createSelector([getUsersById, getCurrentUserId], (usersById, userId) => usersById[userId] ?? {});
export const getUserSettings = state => state.users.userSettings;

export const getIsDarkMode = createSelector([getUserSettings], ({ mode }) => isDarkMode(mode));

export const getShowHelptips = createSelector([getTooltipsById], tooltips =>
  Object.values(tooltips).reduce((accu, { readState }) => accu || readState === READ_STATES.unread, false)
);

export const getTooltipsState = createSelector([getTooltipsById, getUserSettings], (byId, { tooltips = {} }) =>
  Object.entries(byId).reduce(
    (accu, [id, value]) => {
      accu[id] = { ...accu[id], ...value };
      return accu;
    },
    { ...tooltips }
  )
);

export const getHas2FA = createSelector(
  [getCurrentUser],
  currentUser => currentUser.hasOwnProperty('tfa_status') && currentUser.tfa_status === twoFAStates.enabled
);

export const getIdAttribute = createSelector([getGlobalSettings], ({ id_attribute = { ...defaultIdAttribute } }) => id_attribute);

export const getOfflineThresholdSettings = createSelector([getGlobalSettings], ({ offlineThreshold }) => ({
  interval: offlineThreshold?.interval || DEVICE_ONLINE_CUTOFF.interval,
  intervalUnit: offlineThreshold?.intervalUnit || DEVICE_ONLINE_CUTOFF.intervalName
}));

export const getRolesList = createSelector([getRolesById], rolesById => Object.entries(rolesById).map(([id, role]) => ({ id, ...role })));

export const getCurrentSession = state => state.users.currentSession;
