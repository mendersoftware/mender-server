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
import React from 'react';

import {
  CheckCircle as CheckIcon,
  Error as ErrorIcon,
  Help as HelpIcon,
  ArrowDropDownCircleOutlined as ScrollDownIcon,
  Warning as WarningIcon
} from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import pluralize from 'pluralize';

import Time from '../../common/time';

const errorIcon = <ErrorIcon className="red" />;
const successIcon = <CheckIcon className="green" />;
const questionIcon = <HelpIcon />;
const warningIcon = <WarningIcon />;

export const monitoringSeverities = {
  CRITICAL: 'CRITICAL',
  CRITICAL_FLAPPING: 'CRITICAL_FLAPPING',
  OK: 'OK',
  WARNING: 'WARNING',
  UNKNOWN: 'UNKNOWN'
};

export const severityMap = {
  [monitoringSeverities.CRITICAL]: { className: 'red', icon: errorIcon },
  [monitoringSeverities.CRITICAL_FLAPPING]: { className: '', icon: errorIcon },
  [monitoringSeverities.OK]: { className: '', icon: successIcon },
  [monitoringSeverities.UNKNOWN]: { className: '', icon: questionIcon },
  [monitoringSeverities.WARNING]: { className: '', icon: warningIcon }
};

const useStyles = makeStyles()(theme => ({
  deviceDetailNotification: {
    marginBottom: theme.spacing(),
    paddingBottom: theme.spacing(1.5),
    paddingTop: theme.spacing(1.5),
    '&.red, &.green': {
      color: theme.palette.text.primary
    },
    '&.bordered': {
      border: `1px solid ${theme.palette.grey[500]}`,
      background: `${theme.palette.grey[600]}15`,
      '&.red': {
        borderColor: theme.palette.error.main,
        background: theme.palette.error.light
      }
    },
    '> span': {
      marginRight: theme.spacing(2)
    }
  },
  textSpacing: {
    marginLeft: theme.spacing(0.5),
    marginRight: theme.spacing(0.5)
  },
  downButton: {
    marginBottom: theme.spacing(-0.5)
  }
}));

export const BaseNotification = ({ bordered = true, className = '', children, severity, onClick }) => {
  const { classes } = useStyles();
  const mappedSeverity = severityMap[severity] ?? severityMap.UNKNOWN;
  return (
    <div
      className={`flexbox center-aligned padding-small ${classes.deviceDetailNotification} ${bordered ? 'bordered' : ''} ${
        mappedSeverity.className
      } ${className} ${onClick ? 'clickable' : ''}`}
      onClick={onClick}
    >
      <span>{mappedSeverity.icon}</span>
      <div className="flexbox center-aligned">{children}</div>
    </div>
  );
};

export const LastConnection = ({ check_in_time }) => {
  const { classes } = useStyles();

  return check_in_time ? (
    <BaseNotification severity={monitoringSeverities.CRITICAL}>
      Device has not connected to the server since <Time className={classes.textSpacing} value={check_in_time} />
    </BaseNotification>
  ) : (
    <BaseNotification severity={monitoringSeverities.WARNING}>The device has never connected to the server</BaseNotification>
  );
};

export const ServiceNotification = ({ alerts, onClick }) => {
  const { classes } = useStyles();

  return (
    <BaseNotification onClick={onClick} severity={monitoringSeverities.CRITICAL_FLAPPING}>
      {alerts.length} {pluralize('service', alerts.length)} reported issues. View details in the <a className={classes.textSpacing}>monitoring section</a> below
      <a className={classes.textSpacing}>
        <ScrollDownIcon className={classes.downButton} fontSize="small" />
      </a>
    </BaseNotification>
  );
};

export const NoAlertsHeaderNotification = () => (
  <BaseNotification bordered={false} severity={monitoringSeverities.OK}>
    No reported issues
  </BaseNotification>
);

export const DeviceOfflineHeaderNotification = ({ offlineThresholdSettings }) => (
  <BaseNotification className="column-data" bordered={false} severity={monitoringSeverities.CRITICAL}>
    <div className="key muted margin-right-small">
      <b>Device offline</b>
    </div>
    Last check-in over {offlineThresholdSettings.interval} {pluralize(offlineThresholdSettings.intervalUnit, offlineThresholdSettings.interval)} ago
  </BaseNotification>
);

export const DeviceNotifications = ({ alerts, device, onClick }) => {
  const { check_in_time = '', check_in_time_exact = undefined, isOffline } = device;
  return (
    <>
      {isOffline && <LastConnection check_in_time={check_in_time_exact ?? check_in_time} />}
      {Boolean(alerts.length) && <ServiceNotification alerts={alerts} onClick={onClick} />}
    </>
  );
};

export default DeviceNotifications;
