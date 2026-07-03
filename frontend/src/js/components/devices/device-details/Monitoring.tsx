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

import { CheckCircle as CheckIcon, Cancel as ErrorIcon, Help as HelpIcon, Warning as WarningIcon } from '@mui/icons-material';
import { Table, TableBody, TableCell, TableRow, Typography, tableCellClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { ContentSection } from '@northern.tech/common-ui/ContentSection';
import DocsLink from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { Link } from '@northern.tech/common-ui/Link';
import Pagination from '@northern.tech/common-ui/Pagination';
import Time from '@northern.tech/common-ui/Time';
import storeActions from '@northern.tech/store/actions';
import { BENEFITS, DEVICE_LIST_DEFAULTS } from '@northern.tech/store/constants';
import { getOfflineThresholdSettings, getTenantCapabilities } from '@northern.tech/store/selectors';
import { getDeviceAlerts } from '@northern.tech/store/thunks';

import MonitorDetailsDialog from '../dialogs/MonitorDetailsDialog';
import { DeviceOfflineHeaderNotification } from './Notifications';

const errorIcon = <ErrorIcon className="red" />;
const successIcon = <CheckIcon className="green" />;
const questionIcon = <HelpIcon />;
const warningIcon = <WarningIcon />;

const monitoringSeverities = {
  CRITICAL: 'CRITICAL',
  CRITICAL_FLAPPING: 'CRITICAL_FLAPPING',
  OK: 'OK',
  WARNING: 'WARNING',
  UNKNOWN: 'UNKNOWN'
};

const severityMap = {
  [monitoringSeverities.CRITICAL]: { className: 'red', icon: errorIcon, label: 'Critical' },
  [monitoringSeverities.CRITICAL_FLAPPING]: { className: '', icon: errorIcon, label: 'Critical flapping' },
  [monitoringSeverities.OK]: { className: '', icon: successIcon, label: 'OK' },
  [monitoringSeverities.UNKNOWN]: { className: '', icon: questionIcon, label: 'Unknown' },
  [monitoringSeverities.WARNING]: { className: '', icon: warningIcon, label: 'Warning' }
};

const useStyles = makeStyles()(() => ({
  table: {
    width: 'auto',
    maxWidth: 900,
    [`.${tableCellClasses.root}`]: { borderBottom: 'none', whiteSpace: 'nowrap' }
  },
  mutedIcon: { opacity: 0.5 }
}));

const { setAlertListState } = storeActions;

const { page: defaultPage, perPage: defaultPerPage } = DEVICE_LIST_DEFAULTS;

export const DeviceMonitorsMissingNote = () => (
  <Typography className="align-center full-width margin-top-large">
    No alert monitor is currently configured for this device.
    <br />
    Please <DocsLink path="add-ons/monitor" title="see the documentation" /> for a description on how to configure different kinds of monitors.
  </Typography>
);

const columns = [
  { key: 'icon', Component: ({ alert }) => (severityMap[alert.level] ?? severityMap[monitoringSeverities.UNKNOWN]).icon },
  { key: 'name', Component: ({ alert }) => alert.name, style: { width: '100%', whiteSpace: 'normal' } },
  { key: 'level', Component: ({ alert }) => (severityMap[alert.level] ?? severityMap[monitoringSeverities.UNKNOWN]).label },
  { key: 'time', Component: ({ alert }) => <Time value={alert.timestamp} /> },
  {
    key: 'details',
    Component: ({ alert, onDetailsClick }) => {
      const { description, lines_before = [], lines_after = [], line_matching = '' } = alert.subject.details ?? {};
      const lines = [...lines_before, line_matching, ...lines_after].filter(i => i);
      return (lines.length || description) && <Link onClick={() => onDetailsClick(alert)}>view {lines.length ? 'log' : 'details'}</Link>;
    }
  }
];

const MonitoringAlerts = ({ alerts, onDetailsClick, muted = false }) => {
  const { classes } = useStyles();
  return (
    <Table className={classes.table} size="small">
      <TableBody>
        {alerts.map(alert => (
          <TableRow key={alert.id}>
            {columns.map(({ key, Component, style }) => (
              <TableCell key={key} className={muted && key === 'icon' ? classes.mutedIcon : undefined} style={style}>
                <Component alert={alert} onDetailsClick={onDetailsClick} />
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
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
    <ContentSection
      isAddOn
      postTitle={
        <>
          {!!monitors.length && (
            <>
              Latest update: <Time value={updated_ts} />{' '}
            </>
          )}
          <EnterpriseNotification id={BENEFITS.deviceMonitor.id} />
        </>
      }
      title="Monitoring"
    >
      {hasMonitorsDefined || isOffline ? (
        <>
          {hasMonitorsDefined && !latestAlerts.length && (
            <div className="flexbox align-items-center">
              <CheckIcon className="green" />
              <Typography variant="subtitle1" className="margin-left-x-small">
                No reported issues
              </Typography>
            </div>
          )}
          {!!latestAlerts.length && <MonitoringAlerts alerts={latestAlerts} onDetailsClick={onDetailsClick} />}
          {isOffline && <DeviceOfflineHeaderNotification offlineThresholdSettings={offlineThresholdSettings} />}
        </>
      ) : (
        hasMonitor && <DeviceMonitorsMissingNote />
      )}
      {alerts.length ? (
        <>
          <div className="margin-top-large">
            <Typography className="margin-bottom-small" variant="subtitle1">
              Alert history
            </Typography>
            <MonitoringAlerts alerts={alerts} onDetailsClick={onDetailsClick} muted />
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
        hasMonitorsDefined && <Typography className="margin-left-large">There are currently no issues reported</Typography>
      )}
    </ContentSection>
  );
};

export const MonitoringTab = ({ device }) => {
  const [monitorDetails, setMonitorDetails] = useState(null);

  return (
    <>
      <DeviceMonitoring device={device} onDetailsClick={setMonitorDetails} />
      <MonitorDetailsDialog alert={monitorDetails} onClose={() => setMonitorDetails(null)} />
    </>
  );
};

export default MonitoringTab;
