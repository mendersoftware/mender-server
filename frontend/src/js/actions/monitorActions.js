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
import Api, { apiUrl, headerNames } from '../api/general-api';
import { TIMEOUTS } from '../constants/appConstants';
import * as DeviceConstants from '../constants/deviceConstants';
import * as MonitorConstants from '../constants/monitorConstants';
import { getDeviceFilters } from '../selectors';
import { commonErrorFallback, commonErrorHandler, setSnackbar } from './appActions';
import { convertDeviceListStateToFilters, getSearchEndpoint } from './deviceActions';

export const monitorApiUrlv1 = `${apiUrl.v1}/devicemonitor`;

const { page: defaultPage, perPage: defaultPerPage } = DeviceConstants.DEVICE_LIST_DEFAULTS;

const cutoffLength = 75;
const ellipsis = '...';

const longTextTrimmer = text => (text.length >= cutoffLength + ellipsis.length ? `${text.substring(0, cutoffLength + ellipsis.length)}${ellipsis}` : text);

const sanitizeDeviceAlerts = alerts => alerts.map(alert => ({ ...alert, fullName: alert.name, name: longTextTrimmer(alert.name) }));

export const setAlertListState = selectionState => (dispatch, getState) =>
  Promise.resolve(
    dispatch({
      type: MonitorConstants.SET_ALERT_LIST_STATE,
      value: { ...getState().monitor.alerts.alertList, ...selectionState }
    })
  );

export const getDeviceAlerts =
  (id, config = {}) =>
  dispatch => {
    const { page = defaultPage, perPage = defaultPerPage, issuedBefore, issuedAfter, sortAscending = false } = config;
    const issued_after = issuedAfter ? `&issued_after=${issuedAfter}` : '';
    const issued_before = issuedBefore ? `&issued_before=${issuedBefore}` : '';
    return Api.get(`${monitorApiUrlv1}/devices/${id}/alerts?page=${page}&per_page=${perPage}${issued_after}${issued_before}&sort_ascending=${sortAscending}`)
      .catch(err => commonErrorHandler(err, `Retrieving device alerts for device ${id} failed:`, dispatch))
      .then(res =>
        Promise.all([
          dispatch({
            type: MonitorConstants.RECEIVE_DEVICE_ALERTS,
            deviceId: id,
            alerts: sanitizeDeviceAlerts(res.data)
          }),
          dispatch(setAlertListState({ total: Number(res.headers[headerNames.total]) }))
        ])
      );
  };

export const getLatestDeviceAlerts =
  (id, config = {}) =>
  dispatch => {
    const { page = defaultPage, perPage = 10 } = config;
    return Api.get(`${monitorApiUrlv1}/devices/${id}/alerts/latest?page=${page}&per_page=${perPage}`)
      .catch(err => commonErrorHandler(err, `Retrieving device alerts for device ${id} failed:`, dispatch))
      .then(res =>
        Promise.resolve(
          dispatch({
            type: MonitorConstants.RECEIVE_LATEST_DEVICE_ALERTS,
            deviceId: id,
            alerts: sanitizeDeviceAlerts(res.data)
          })
        )
      );
  };

export const getIssueCountsByType =
  (type, options = {}) =>
  (dispatch, getState) => {
    const state = getState();
    const { filters = getDeviceFilters(state), group, status, ...remainder } = options;
    const { applicableFilters: nonMonitorFilters, filterTerms } = convertDeviceListStateToFilters({
      ...remainder,
      filters,
      group,
      offlineThreshold: getState().app.offlineThreshold,
      selectedIssues: [type],
      status
    });
    return Api.post(getSearchEndpoint(state.app.features.hasReporting), {
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
        return Promise.resolve(
          dispatch({
            counts: { filtered, total },
            issueType: type,
            type: MonitorConstants.RECEIVE_DEVICE_ISSUE_COUNTS
          })
        );
      });
  };

export const getDeviceMonitorConfig = id => dispatch =>
  Api.get(`${monitorApiUrlv1}/devices/${id}/config`)
    .catch(err => commonErrorHandler(err, `Retrieving device monitor config for device ${id} failed:`, dispatch))
    .then(({ data }) => {
      let tasks = [
        dispatch({
          type: MonitorConstants.RECEIVE_DEVICE_MONITOR_CONFIG,
          device: { id, monitors: data }
        })
      ];
      tasks.push(Promise.resolve(data));
      return Promise.all(tasks);
    });

export const changeNotificationSetting =
  (enabled, channel = MonitorConstants.alertChannels.email) =>
  dispatch => {
    return Api.put(`${monitorApiUrlv1}/settings/global/channel/alerts/${channel}/status`, { enabled })
      .catch(err => commonErrorHandler(err, `${enabled ? 'En' : 'Dis'}abling  ${channel} alerts failed:`, dispatch))
      .then(() =>
        Promise.all([
          dispatch({
            type: MonitorConstants.CHANGE_ALERT_CHANNEL,
            channel,
            enabled
          }),
          dispatch(setSnackbar(`Successfully ${enabled ? 'en' : 'dis'}abled ${channel} alerts`, TIMEOUTS.fiveSeconds))
        ])
      );
  };
