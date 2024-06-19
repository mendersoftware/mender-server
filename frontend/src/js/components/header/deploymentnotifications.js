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
import React from 'react';
import { Link } from 'react-router-dom';

// material ui
import { Refresh as RefreshIcon } from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import { DEPLOYMENT_ROUTES } from '../../constants/deploymentConstants';

const useStyles = makeStyles()(theme => ({
  icon: { color: theme.palette.grey[500], margin: '0 7px 0 10px', top: '5px', fontSize: '20px' }
}));

const DeploymentNotifications = props => {
  const { classes } = useStyles();
  return (
    <Link to={DEPLOYMENT_ROUTES.active.route} className="header-section">
      <span>{props.inprogress}</span>
      <RefreshIcon className={`flip-horizontal ${classes.icon}`} />
    </Link>
  );
};

export default DeploymentNotifications;
