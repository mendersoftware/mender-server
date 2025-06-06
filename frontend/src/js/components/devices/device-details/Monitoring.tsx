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
import { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import DocsLink from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import Pagination from '@northern.tech/common-ui/Pagination';
import Time from '@northern.tech/common-ui/Time';
import storeActions from '@northern.tech/store/actions';
import { BENEFITS, DEVICE_LIST_DEFAULTS } from '@northern.tech/store/constants';
import { getOfflineThresholdSettings, getTenantCapabilities } from '@northern.tech/store/selectors';
import { getDeviceAlerts } from '@northern.tech/store/thunks';

import MonitorDetailsDialog from '../dialogs/MonitorDetailsDialog';
import { DeviceConnectionNote } from './Connection';
import DeviceDataCollapse from './DeviceDataCollapse';
import { DeviceOfflineHeaderNotification, NoAlertsHeaderNotification, monitoringSeverities, severityMap } from './Notifications';

const { setAlertListState } = storeActions;

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

export const DeviceMonitorsMissingNote = () => (
  <DeviceConnectionNote>
    No alert monitor is currently configured for this device.
    <br />
    Please <DocsLink path="add-ons/monitor" title="see the documentation" /> for a description on how to configure different kinds of monitors.
  </DeviceConnectionNote>
);

const MonitoringAlert = ({ alert, className = '', onDetailsClick, style }) => {
  const { description, lines_before = [], lines_after = [], line_matching = '' } = alert.subject.details;
  const lines = [...lines_before, line_matching, ...lines_after].filter(i => i);
  return (
    <div className={`monitoring-alert column-data ${className}`} style={style}>
      {(severityMap[alert.level] ?? severityMap[monitoringSeverities.UNKNOWN]).icon}
      <div className="key muted">
        <b>{alert.name}</b>
      </div>
      <div>{alert.level}</div>
      <Time value={alert.timestamp} />
      {(lines.length || description) && <a onClick={() => onDetailsClick(alert)}>view {lines.length ? 'log' : 'details'}</a>}
    </div>
  );
};

const paginationCutoff = defaultPerPage;
export const DeviceMonitoring = ({ device, onDetailsClick }) => {
  const { hasMonitor } = useSelector(state => getTenantCapabilities(state));
  const { alerts = [], latest: latestAlerts = [] } = useSelector(state => state.monitor.alerts.byDeviceId[device.id]) ?? {};
  const alertListState = useSelector(state => state.monitor.alerts.alertList) ?? {};
  const offlineThresholdSettings = useSelector(getOfflineThresholdSettings);
  const dispatch = useDispatch();
  const { page: pageNo = defaultPage, perPage: pageLength = defaultPerPage, total: alertCount } = alertListState;

  useEffect(() => {
    if (hasMonitor) {
      dispatch(getDeviceAlerts({ id: device.id, config: alertListState }));
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [device.id, dispatch, hasMonitor, pageNo, pageLength]);

  const onChangePage = page => dispatch(setAlertListState({ page }));

  const onChangeRowsPerPage = perPage => dispatch(setAlertListState({ page: 1, perPage }));

  const { monitors = [], isOffline, updated_ts = '' } = device;
  const hasMonitorsDefined = !!(monitors.length || alerts.length || latestAlerts.length);

  return (
    <DeviceDataCollapse
      header={
        hasMonitorsDefined || isOffline ? (
          <>
            {hasMonitorsDefined && !latestAlerts.length && <NoAlertsHeaderNotification />}
            {latestAlerts.map(alert => (
              <MonitoringAlert className="margin-bottom-x-small" alert={alert} key={alert.id} onDetailsClick={onDetailsClick} />
            ))}
            {isOffline && <DeviceOfflineHeaderNotification offlineThresholdSettings={offlineThresholdSettings} />}
          </>
        ) : (
          hasMonitor && <DeviceMonitorsMissingNote />
        )
      }
      isAddOn
      title={
        <div className="flexbox center-aligned">
          <h4 className="margin-bottom-small margin-right">Monitoring</h4>
          {!!monitors.length && <Time className="muted" value={updated_ts} />}
          <EnterpriseNotification className="margin-left-small" id={BENEFITS.deviceMonitor.id} />
        </div>
      }
    >
      {alerts.length ? (
        <>
          <div>
            <h4 className="muted">Alert history</h4>
            {alerts.map(alert => (
              <MonitoringAlert alert={alert} key={alert.id} onDetailsClick={onDetailsClick} />
            ))}
          </div>
          <div className="flexbox margin-top">
            {alertCount > paginationCutoff && (
              <Pagination
                className="margin-top-none"
                count={alertCount}
                rowsPerPage={pageLength}
                onChangeRowsPerPage={onChangeRowsPerPage}
                page={pageNo}
                onChangePage={onChangePage}
              />
            )}
          </div>
        </>
      ) : (
        hasMonitorsDefined && (
          <p className="muted margin-left-large" style={{ fontSize: 'larger' }}>
            There are currently no issues reported
          </p>
        )
      )}
    </DeviceDataCollapse>
  );
};

export const MonitoringTab = ({ device }) => {
  const [monitorDetails, setMonitorDetails] = useState();

  return (
    <>
      <DeviceMonitoring device={device} onDetailsClick={setMonitorDetails} />
      <MonitorDetailsDialog alert={monitorDetails} onClose={() => setMonitorDetails()} />
    </>
  );
};

export default MonitoringTab;
