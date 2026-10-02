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
// material ui
import { Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import FileSize from '@northern.tech/common-ui/FileSize';
import { RelativeTime } from '@northern.tech/common-ui/Time';
import { TwoColumnData } from '@northern.tech/common-ui/TwoColumnData';
import { DEPLOYMENT_STATES, DEPLOYMENT_TYPES } from '@northern.tech/store/constants';
import type { IdAttribute } from '@northern.tech/store/constants';
import type { Deployment } from '@northern.tech/store/deploymentsSlice';
import type { Device } from '@northern.tech/store/devicesSlice';
import { useDeploymentDevice } from '@northern.tech/store/useDeploymentDevice';

import DeploymentStats from './DeploymentStatus';
import type { ColumnHeader } from './DeploymentsList';
import { getDeploymentTargetText } from './deployment-wizard/SoftwareDevices';
import { DeploymentProgress } from './progress/DeploymentProgress';
import { getDeploymentStartTime } from './progress/utils';

interface ColumnComponentProps {
  className: string;
  deployment: Deployment;
  devicesById: Record<string, unknown>;
  direction: string;
  idAttribute: IdAttribute | string;
  started: string;
  wrappingClass: string;
}

export const DeploymentDeviceCount = ({ className, deployment }: Pick<ColumnComponentProps, 'className' | 'deployment'>) => (
  <Typography variant="body2" className={className} key="DeploymentDeviceCount">
    {Math.max(deployment.device_count || 0, deployment.max_devices || 0)}
  </Typography>
);
export const DeploymentDeviceGroup = ({ deployment, devicesById, idAttribute, wrappingClass }: Partial<ColumnComponentProps>) => {
  const deploymentName = getDeploymentTargetText({ deployment, devicesById, idAttribute });
  return (
    <Typography variant="body2" className={wrappingClass} key="DeploymentDeviceGroup" title={deploymentName}>
      {deploymentName}
    </Typography>
  );
};
export const DeploymentEndTime = ({ className, deployment }: Pick<ColumnComponentProps, 'className' | 'deployment'>) => (
  <RelativeTime className={className} key="DeploymentEndTime" updateTime={deployment.finished} shouldCount="none" />
);
export const DeploymentPhases = ({ className, deployment }: Pick<ColumnComponentProps, 'className' | 'deployment'>) => (
  <Typography variant="body2" className={className} key="DeploymentPhases">
    {deployment.phases ? deployment.phases.length : '-'}
  </Typography>
);
export const DeploymentStatus = ({ deployment }: Pick<ColumnComponentProps, 'deployment'>) => (
  <DeploymentStats key="DeploymentStatus" deployment={deployment} />
);
export const DeploymentRelease = ({
  deployment: { artifact_name, type = DEPLOYMENT_TYPES.software },
  wrappingClass
}: Pick<ColumnComponentProps, 'deployment' | 'wrappingClass'>) => {
  const deploymentRelease = type === DEPLOYMENT_TYPES.configuration ? type : artifact_name;
  return (
    <Typography variant="body2" className={wrappingClass} key="DeploymentRelease" title={deploymentRelease}>
      {deploymentRelease}
    </Typography>
  );
};
export const DeploymentStartTime = ({ className, direction = 'both', started }: Pick<ColumnComponentProps, 'className' | 'direction' | 'started'>) => (
  <RelativeTime className={className} key="DeploymentStartTime" updateTime={started} shouldCount={direction} />
);

export const DeploymentSize = ({ deployment: { statistics } }: Pick<ColumnComponentProps, 'deployment'>) => (
  <Typography variant="body2" className="align-right" component="div">
    {statistics.total_size ? <FileSize fileSize={statistics.total_size} /> : '-'}
  </Typography>
);

const useStyles = makeStyles()(() => ({
  compactProgress: { minWidth: 270 },
  textWrapping: { whiteSpace: 'initial' }
}));

interface DeploymentItemCommonProps {
  className?: string;
  columnHeaders: ColumnHeader[];
  deployment: Deployment;
  devices: Record<string, Device>;
  idAttribute?: IdAttribute | string;
  openReport: (type: string, id: string) => void;
  type: string;
}

export interface DeploymentItemProps extends DeploymentItemCommonProps {
  isCompact?: boolean;
  isEnterprise?: boolean;
}

interface DeploymentItemCompactProps extends DeploymentItemCommonProps {
  started: string;
  wrappingClass: string;
}

export const DeploymentItemCompact = ({
  className = '',
  columnHeaders,
  deployment,
  devices,
  idAttribute,
  openReport,
  started,
  type,
  wrappingClass
}: DeploymentItemCompactProps) => {
  useDeploymentDevice(deployment.name);

  const { classes } = useStyles();

  // Find the progress column to render it separately
  const { renderer: ProgressColumn, props: progressProps, title: progressTitle } = columnHeaders.find(col => col.renderer === DeploymentProgress) || {};
  const otherColumns = columnHeaders.filter(col => col.renderer !== DeploymentProgress);

  const deploymentInfo = otherColumns.reduce((accu, column) => {
    const ColumnComponent = column.renderer;
    accu[column.title] = (
      <ColumnComponent
        className={column.class || ''}
        idAttribute={idAttribute}
        deployment={deployment}
        devicesById={devices}
        started={started}
        wrappingClass={wrappingClass}
        {...column.props}
      />
    );
    return accu;
  }, {});
  if (ProgressColumn) {
    deploymentInfo[progressTitle] = (
      <div className={classes.compactProgress}>
        <ProgressColumn deployment={deployment} {...progressProps} />
      </div>
    );
  }

  return (
    <div className={`padding-small relative clickable ${className}`} role="listitem" onClick={() => openReport(type, deployment.id)}>
      <TwoColumnData data={deploymentInfo} />
    </div>
  );
};

export const DeploymentItem = ({
  className = '',
  columnHeaders,
  deployment,
  devices,
  idAttribute,
  isCompact,
  isEnterprise,
  openReport,
  type
}: DeploymentItemProps) => {
  useDeploymentDevice(deployment.name);

  const { classes } = useStyles();

  const { created } = deployment;

  const started = (isEnterprise && getDeploymentStartTime(deployment)) || created;
  const wrappingClass = `text-overflow ${type === DEPLOYMENT_STATES.inprogress ? classes.textWrapping : ''}`;

  if (isCompact) {
    return (
      <DeploymentItemCompact
        className={className}
        columnHeaders={columnHeaders}
        deployment={deployment}
        devices={devices}
        key={deployment.id}
        idAttribute={idAttribute}
        openReport={openReport}
        started={started}
        type={type}
        wrappingClass={wrappingClass}
      />
    );
  }
  return (
    <div className={`padding-small relative clickable ${className}`} role="listitem" onClick={() => openReport(type, deployment.id)}>
      {columnHeaders.map(({ renderer: ColumnComponent, class: columnClass = '', props }, i) => (
        <ColumnComponent
          key={`deploy-item-${i}`}
          className={columnClass}
          idAttribute={idAttribute}
          deployment={deployment}
          devicesById={devices}
          started={started}
          wrappingClass={wrappingClass}
          {...props}
        />
      ))}
    </div>
  );
};

export default DeploymentItem;
