// Copyright 2017 Northern.tech AS
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
import { Link as RouterLink } from 'react-router';

// material ui
import { SyncOutlined as RefreshIcon } from '@mui/icons-material';
import { Button, Tooltip } from '@mui/material';

import { DEPLOYMENT_ROUTES } from '@northern.tech/store/constants';

const DeploymentNotifications = ({ className = '', inprogress }) => (
  <Tooltip title="Deployments in progress">
    <Button className={className} color="inherit" component={RouterLink} startIcon={<RefreshIcon fontSize="small" />} to={DEPLOYMENT_ROUTES.active.route}>
      {inprogress}
    </Button>
  </Tooltip>
);

export default DeploymentNotifications;
