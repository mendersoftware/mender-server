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
import { Alert, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { Link } from '@northern.tech/common-ui/Link';
import Time from '@northern.tech/common-ui/Time';
import pluralize from 'pluralize';

const useStyles = makeStyles()(theme => ({
  textSpacing: {
    marginLeft: theme.spacing(0.5),
    marginRight: theme.spacing(0.5)
  }
}));
export const LastConnection = ({ check_in_time }) => {
  const { classes } = useStyles();

  return check_in_time ? (
    <Alert severity="warning">
      Device has not connected to the server since <Time className={classes.textSpacing} value={check_in_time} />
    </Alert>
  ) : (
    <Alert severity="info">The device has never connected to the server</Alert>
  );
};

export const ServiceNotification = ({ alerts, onClick }) => (
  <Alert onClick={onClick} severity="error" className="clickable margin-top-x-small">
    {alerts.length} {pluralize('service', alerts.length)} reported issues. View the details in the <Link>monitoring section</Link> below
  </Alert>
);

export const DeviceOfflineHeaderNotification = ({ offlineThresholdSettings }) => (
  <Alert severity="warning" className="margin-top-small margin-bottom-small">
    <Typography variant="body2" className="margin-right-small">
      Device may be offline: last check-in over {offlineThresholdSettings.interval}{' '}
      {pluralize(offlineThresholdSettings.intervalUnit, offlineThresholdSettings.interval)} ago
    </Typography>
  </Alert>
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
