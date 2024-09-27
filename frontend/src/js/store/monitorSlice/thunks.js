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
import storeActions from '@northern.tech/store/actions';
import Api from '@northern.tech/store/api/general-api';
import { DEVICE_LIST_DEFAULTS, TIMEOUTS, alertChannels, headerNames } from '@northern.tech/store/constants';
import { getDeviceFilters, getSearchEndpoint } from '@northern.tech/store/selectors';
import { commonErrorFallback, commonErrorHandler } from '@northern.tech/store/store';
import { convertDeviceListStateToFilters } from '@northern.tech/store/utils';
import { createAsyncThunk } from '@reduxjs/toolkit';

import { actions, sliceName } from '.';
import { monitorApiUrlv1 } from './constants';

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

const cutoffLength = 75;
const ellipsis = '...';
const longTextTrimmer = text => (text.length >= cutoffLength + ellipsis.length ? `${text.substring(0, cutoffLength + ellipsis.length)}${ellipsis}` : text);

const sanitizeDeviceAlerts = alerts => alerts.map(alert => ({ ...alert, fullName: alert.name, name: longTextTrimmer(alert.name) }));

export const getDeviceAlerts = createAsyncThunk(`${sliceName}/getDeviceAlerts`, ({ id, config = {} }, { dispatch }) => {
  const { page = defaultPage, perPage = defaultPerPage, issuedBefore, issuedAfter, sortAscending = false } = config;
  const issued_after = issuedAfter ? `&issued_after=${issuedAfter}` : '';
  const issued_before = issuedBefore ? `&issued_before=${issuedBefore}` : '';
  return Api.get(`${monitorApiUrlv1}/devices/${id}/alerts?page=${page}&per_page=${perPage}${issued_after}${issued_before}&sort_ascending=${sortAscending}`)
    .catch(err => commonErrorHandler(err, `Retrieving device alerts for device ${id} failed:`, dispatch))
    .then(res =>
      Promise.all([
        dispatch(actions.receiveDeviceAlerts({ deviceId: id, alerts: sanitizeDeviceAlerts(res.data) })),
        dispatch(actions.setAlertListState({ total: Number(res.headers[headerNames.total]) }))
      ])
    );
});

export const getLatestDeviceAlerts = createAsyncThunk(`${sliceName}/getLatestDeviceAlerts`, ({ id, config = {} }, { dispatch }) => {
  const { page = defaultPage, perPage = 10 } = config;
  return Api.get(`${monitorApiUrlv1}/devices/${id}/alerts/latest?page=${page}&per_page=${perPage}`)
    .catch(err => commonErrorHandler(err, `Retrieving device alerts for device ${id} failed:`, dispatch))
    .then(res => Promise.resolve(dispatch(actions.receiveLatestDeviceAlerts({ deviceId: id, alerts: sanitizeDeviceAlerts(res.data) }))));
});

export const getIssueCountsByType = createAsyncThunk(`${sliceName}/getIssueCountsByType`, ({ type, options = {} }, { dispatch, getState }) => {
  const state = getState();
  const { filters = getDeviceFilters(state), group, status, ...remainder } = options;
  const { applicableFilters: nonMonitorFilters, filterTerms } = convertDeviceListStateToFilters({
    ...remainder,
    filters,
    group,
    offlineThreshold: state.app.offlineThreshold,
    selectedIssues: [type],
    status
  });
  return Api.post(getSearchEndpoint(getState()), {
    page: 1,
    per_page: 1,
    filters: filterTerms,
    attributes: [{ scope: 'identity', attribute: 'status' }]
  })
    .catch(err => commonErrorHandler(err, `Retrieving issue counts failed:`, dispatch, commonErrorFallback))
    .then(res => {
      const total = nonMonitorFilters.length ? state.monitor.issueCounts.byType[type].total : Number(res.headers[headerNames.total]);
      const filtered = nonMonitorFilters.length ? Number(res.headers[headerNames.total]) : total;
      if (total === state.monitor.issueCounts.byType[type].total && filtered === state.monitor.issueCounts.byType[type].filtered) {
        return Promise.resolve();
      }
      return Promise.resolve(dispatch(actions.receiveDeviceIssueCounts({ counts: { filtered, total }, issueType: type })));
    });
});

export const getDeviceMonitorConfig = createAsyncThunk(`${sliceName}/getDeviceMonitorConfig`, (id, { dispatch }) =>
  Api.get(`${monitorApiUrlv1}/devices/${id}/config`)
    .catch(err => commonErrorHandler(err, `Retrieving device monitor config for device ${id} failed:`, dispatch))
    .then(({ data }) => Promise.all([dispatch(storeActions.receivedDevice({ id, monitors: data }), Promise.resolve(data))]))
);

export const changeNotificationSetting = createAsyncThunk(
  `${sliceName}/changeNotificationSetting`,
  ({ enabled, channel = alertChannels.email }, { dispatch }) => {
    return Api.put(`${monitorApiUrlv1}/settings/global/channel/alerts/${channel}/status`, { enabled })
      .catch(err => commonErrorHandler(err, `${enabled ? 'En' : 'Dis'}abling  ${channel} alerts failed:`, dispatch))
      .then(() =>
        Promise.all([
          dispatch(actions.changeAlertChannel({ channel, enabled })),
          dispatch(storeActions.setSnackbar(`Successfully ${enabled ? 'en' : 'dis'}abled ${channel} alerts`, TIMEOUTS.fiveSeconds))
        ])
      );
  }
);
