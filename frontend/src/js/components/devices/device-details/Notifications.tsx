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
import { ArrowDropDownCircleOutlined as ScrollDownIcon } from '@mui/icons-material';
import { Alert } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Time from '@northern.tech/common-ui/Time';
import pluralize from 'pluralize';

const useStyles = makeStyles()(theme => ({
  textSpacing: {
    marginLeft: theme.spacing(0.5),
    marginRight: theme.spacing(0.5)
  },
  downButton: {
    marginBottom: theme.spacing(-0.5)
  }
}));

export const BaseNotification = ({ children, severity, onClick }) => (
  <Alert severity={severity} onClick={onClick} className={onClick ? 'clickable' : ''}>
    <div className="flexbox center-aligned">{children}</div>
  </Alert>
);

export const LastConnection = ({ check_in_time }) => {
  const { classes } = useStyles();

  return check_in_time ? (
    <BaseNotification severity="warning">
      Device has not connected to the server since <Time className={classes.textSpacing} value={check_in_time} />
    </BaseNotification>
  ) : (
    <BaseNotification severity="info">The device has never connected to the server</BaseNotification>
  );
};

export const ServiceNotification = ({ alerts, onClick }) => {
  const { classes } = useStyles();

  return (
    <BaseNotification onClick={onClick} severity="error">
      {alerts.length} {pluralize('service', alerts.length)} reported issues. View details in the <a className={classes.textSpacing}>monitoring section</a> below
      <a className={classes.textSpacing}>
        <ScrollDownIcon className={classes.downButton} fontSize="small" />
      </a>
    </BaseNotification>
  );
};

export const NoAlertsHeaderNotification = () => <BaseNotification severity="success">No reported issues</BaseNotification>;

export const DeviceOfflineHeaderNotification = ({ offlineThresholdSettings }) => (
  <BaseNotification severity="error">
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
