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
import { useState } from 'react';

// material ui
import { Cancel as CancelIcon } from '@mui/icons-material';
import { Button, IconButton, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Confirm from '@northern.tech/common-ui/Confirm';
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

const useStyles = makeStyles()(theme => ({
  centered: { display: 'grid', placeSelf: 'center' },
  compactConfirm: { marginTop: theme.spacing(-2), marginLeft: theme.spacing(-2) },
  compactProgress: { minWidth: 270 },
  textWrapping: { whiteSpace: 'initial' }
}));

interface DeploymentItemCommonProps {
  canConfigure?: boolean;
  canDeploy?: boolean;
  className?: string;
  columnHeaders: ColumnHeader[];
  deployment: Deployment;
  devices: Record<string, Device>;
  idAttribute?: IdAttribute | string;
  openReport: (type: string, id: string) => void;
  type: string;
}

export interface DeploymentItemProps extends DeploymentItemCommonProps {
  abort?: (id: string) => void;
  isCompact?: boolean;
  isEnterprise?: boolean;
}

interface DeploymentItemCompactProps extends DeploymentItemCommonProps {
  abort: string | null;
  abortDeployment: (id: string) => void;
  started: string;
  toggleConfirm: (id: string) => void;
  wrappingClass: string;
}

export const DeploymentItemCompact = ({
  abortDeployment,
  abort,
  canConfigure,
  canDeploy,
  className = '',
  columnHeaders,
  deployment,
  devices,
  idAttribute,
  openReport,
  started,
  toggleConfirm,
  type,
  wrappingClass
}: DeploymentItemCompactProps) => {
  useDeploymentDevice(deployment.name);

  const { classes } = useStyles();

  const { id } = deployment;

  let confirmation;
  if (abort === id) {
    confirmation = <Confirm classes={classes.compactConfirm} cancel={() => toggleConfirm(id)} action={() => abortDeployment(id)} type="abort" />;
  }

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
  deploymentInfo[''] = (
    <Button onClick={() => openReport(type, deployment.id)} variant="outlined" size="small">
      View details
    </Button>
  );
  if ((canDeploy || (canConfigure && deployment.type === DEPLOYMENT_TYPES.configuration)) && type !== DEPLOYMENT_STATES.finished) {
    deploymentInfo[' '] = (
      <Tooltip title="Abort" placement="top-start">
        <IconButton onClick={() => toggleConfirm(id)} size="small">
          <CancelIcon fontSize="small" />
        </IconButton>
      </Tooltip>
    );
  }

  return (
    <div className={`padding-small relative ${className}`} role="listitem">
      {!!confirmation && confirmation}
      <TwoColumnData data={deploymentInfo} />
    </div>
  );
};

export const DeploymentItem = ({
  abort: abortDeployment,
  canConfigure,
  canDeploy,
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
  const [abort, setAbort] = useState(null);
  useDeploymentDevice(deployment.name);

  const { classes } = useStyles();

  const toggleConfirm = id => setTimeout(() => setAbort(current => (current ? null : id)), 150);

  const { created, id, phases } = deployment;

  let confirmation;
  if (abort === id) {
    confirmation = <Confirm cancel={() => toggleConfirm(id)} action={() => abortDeployment(id)} type="abort" />;
  }
  const started = isEnterprise && phases?.length >= 1 ? phases[0].start_ts || created : created;
  const wrappingClass = `text-overflow ${type === DEPLOYMENT_STATES.inprogress ? classes.textWrapping : ''}`;

  if (isCompact) {
    return (
      <DeploymentItemCompact
        abort={abort}
        abortDeployment={abortDeployment}
        canConfigure={canConfigure}
        canDeploy={canDeploy}
        className={className}
        columnHeaders={columnHeaders}
        deployment={deployment}
        devices={devices}
        key={deployment.id}
        idAttribute={idAttribute}
        openReport={openReport}
        started={started}
        toggleConfirm={toggleConfirm}
        type={type}
        wrappingClass={wrappingClass}
      />
    );
  }
  return (
    <div className={`padding-small relative ${className}`} role="listitem">
      {!!confirmation && confirmation}
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
      <Button className={`nowrap ${classes.centered}`} onClick={() => openReport(type, deployment.id)} variant="outlined">
        View details
      </Button>
      {(canDeploy || (canConfigure && deployment.type === DEPLOYMENT_TYPES.configuration)) && type !== DEPLOYMENT_STATES.finished && (
        <Tooltip title="Abort" placement="top-start">
          <IconButton className={classes.centered} onClick={() => toggleConfirm(id)} size="small">
            <CancelIcon fontSize="small" />
          </IconButton>
        </Tooltip>
      )}
    </div>
  );
};

export default DeploymentItem;
