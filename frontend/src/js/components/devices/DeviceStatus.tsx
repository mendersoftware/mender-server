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
import { Error as ErrorIcon, ReportProblemOutlined } from '@mui/icons-material';
import { Avatar, Chip, Tooltip, chipClasses } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DEVICE_STATES } from '@northern.tech/store/constants';
import pluralize from 'pluralize';

const NumberIcon = ({ className, value }) => <Avatar className={className}>{value}</Avatar>;

const WarningIcon = <ReportProblemOutlined fontSize="small" />;

const statusTypes = {
  default: { color: 'text.primary', icon: <ErrorIcon />, label: '', notification: { default: '' } },
  authRequests: {
    color: 'text.primary',
    icon: <NumberIcon className="" value={null} />,
    label: '',
    notification: {
      [DEVICE_STATES.accepted]: `This device has a new auth request. This can happen if the device's public key changes. Click on the row to see more details`,
      [DEVICE_STATES.pending]: `This device has a new auth request. Inspect its identity details, then check it to accept it.`
    }
  },
  monitor: {
    color: 'error',
    icon: WarningIcon,
    label: 'monitoring',
    notification: {
      default: `This device has reported an issue. Click on the row to see more details`
    }
  },
  offline: {
    color: 'warning',
    icon: WarningIcon,
    label: 'offline',
    notification: { default: 'This device has not communicated with the Mender backend for a while. Click on the row to see more details' }
  }
};

const useStyles = makeStyles()(theme => ({
  numberIcon: {
    width: 18,
    height: 18,
    backgroundColor: theme.palette.divider,
    [`&.${chipClasses.icon}`]: {
      fontSize: 'x-small'
    }
  }
}));

const DeviceStatus = ({ device: { auth_sets = [], isOffline, monitor = {}, status: deviceStatus } }) => {
  const { classes } = useStyles();
  let color = statusTypes.default.color;
  let label = statusTypes.default.label;
  let icon = statusTypes.default.icon;
  let notification = statusTypes.default.notification.default;

  const pendingAuthSetsCount = auth_sets.filter(item => item.status === DEVICE_STATES.pending).length;
  if (pendingAuthSetsCount) {
    icon = <NumberIcon className={classes.numberIcon} value={pendingAuthSetsCount} />;
    notification = statusTypes.authRequests.notification[deviceStatus] ?? statusTypes.authRequests.notification[DEVICE_STATES.accepted];
    label = `new ${pluralize('request', pendingAuthSetsCount)}`;
  } else if (Object.values(monitor).some(i => i)) {
    color = statusTypes.monitor.color;
    icon = statusTypes.monitor.icon;
    label = statusTypes.monitor.label;
    notification = statusTypes.monitor.notification.default;
  } else if (isOffline) {
    color = statusTypes.offline.color;
    icon = statusTypes.offline.icon;
    label = statusTypes.offline.label;
    notification = statusTypes.offline.notification.default;
  }
  return label ? (
    <Tooltip arrow title={notification} placement="bottom">
      <Chip className="margin-right-small capitalized" size="small" color={color} icon={icon} label={label} variant="outlined" />
    </Tooltip>
  ) : (
    <div className="margin-right-small capitalized">{deviceStatus}</div>
  );
};

export default DeviceStatus;
