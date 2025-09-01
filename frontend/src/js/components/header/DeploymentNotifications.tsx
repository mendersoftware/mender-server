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
import { Link } from 'react-router-dom';

// material ui
import { Refresh as RefreshIcon } from '@mui/icons-material';

import { DEPLOYMENT_ROUTES } from '@northern.tech/store/constants';

const DeploymentNotifications = ({ className = '', inprogress }) => (
  <Link to={DEPLOYMENT_ROUTES.active.route} className={className}>
    <RefreshIcon className="flip-horizontal margin-right-x-small" fontSize="small" />
    <div>{inprogress}</div>
  </Link>
);

export default DeploymentNotifications;
