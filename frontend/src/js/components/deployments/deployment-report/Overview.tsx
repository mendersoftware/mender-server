// Copyright 2020 Northern.tech AS
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

import { Launch as LaunchIcon, ArrowDropDownCircleOutlined as ScrollDownIcon } from '@mui/icons-material';
import { Button, Chip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Time from '@northern.tech/common-ui/Time';
import { SynchronizedTwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { DEPLOYMENT_TYPES } from '@northern.tech/store/constants';

import { getDeploymentTargetText, getDevicesLink } from '../deployment-wizard/SoftwareDevices';

const useStyles = makeStyles()(theme => ({
  chip: {
    opacity: 0.5,
    fontSize: '0.675rem',
    height: 18
  },
  statusWrapper: {
    backgroundColor: theme.palette.background.lightgrey,
    ['&:after']: {
      borderRight: '20px solid',
      borderRightColor: theme.palette.background.lightgrey
    }
  }
}));

const defaultLinkProps = {
  className: 'flexbox centered',
  style: { fontWeight: '500' },
  target: '_blank',
  rel: 'noopener noreferrer'
};

export const DeploymentOverview = ({ creator, deployment, devicesById, idAttribute, onScheduleClick }) => {
  const { classes } = useStyles();
  const {
    artifact_name,
    created: creationTime = new Date().toISOString(),
    devices = {},
    filter,
    group,
    name = '',

    type = DEPLOYMENT_TYPES.software
  } = deployment;
  const isSoftwareDeployment = type === DEPLOYMENT_TYPES.software;

  const deploymentRelease = isSoftwareDeployment ? (
    <Link {...defaultLinkProps} to={`/releases/${encodeURIComponent(artifact_name)}`}>
      {artifact_name}
      <LaunchIcon className="margin-left-small" fontSize="small" />
    </Link>
  ) : (
    type
  );

  const devicesLink = getDevicesLink({
    devices: Object.values(devices),
    filters: filter?.filters,
    group: group || filter?.name,
    name
  });
  const targetDevices = (
    <Link {...defaultLinkProps} to={devicesLink}>
      {getDeploymentTargetText({ deployment, devicesById, idAttribute })}
      <LaunchIcon className="margin-left-small" fontSize="small" />
      <Chip className={`margin-left uppercased ${classes.chip}`} label={filter?.name ? 'dynamic' : 'static'} size="small" />
    </Link>
  );

  const createdBy = creator ? { 'Created by': creator } : {};
  const deploymentInfo = {
    'Release': deploymentRelease,
    'Target device(s)': targetDevices,
    'Category': isSoftwareDeployment ? 'Software update' : 'Configuration',
    ...createdBy,
    'Created at': <Time value={creationTime} />
  };

  return (
    <div className="report-container margin-top-large margin-bottom-large">
      <div>
        <SynchronizedTwoColumnData data={deploymentInfo} />
        <Button endIcon={<ScrollDownIcon fontSize="small" />} className="margin-top" onClick={onScheduleClick} variant="text">
          See schedule details
        </Button>
      </div>
    </div>
  );
};

export default DeploymentOverview;
