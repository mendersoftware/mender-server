// Copyright 2026 Northern.tech AS
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
import { PendingOutlined as QueuedIcon, RotateLeftOutlined } from '@mui/icons-material';
import { Typography } from '@mui/material';

import { mdiSleep as SleepIcon } from '@mdi/js';
import MaterialDesignIcon from '@northern.tech/common-ui/MaterialDesignIcon';

export const statusMap = {
  complete: {
    icon: <MaterialDesignIcon path={SleepIcon} />,
    description: () => 'Complete, awaiting new devices'
  },
  queued: {
    icon: <QueuedIcon />,
    description: () => 'Queued to start'
  },
  paused: { icon: <RotateLeftOutlined fontSize="inherit" />, description: window => `Paused until next window ${window}` }
};

export const DeploymentStatusNotification = ({ status }: { status: string }) =>
  statusMap[status] ? (
    <Typography variant="body2" className="flexbox align-items-center">
      {statusMap[status].icon}
      <span className="margin-left-small">{statusMap[status].description()}</span>
    </Typography>
  ) : null;

export default DeploymentStatusNotification;
