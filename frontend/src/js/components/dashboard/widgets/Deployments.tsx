// Copyright 2022 Northern.tech AS
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
import { Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Time from '@northern.tech/common-ui/Time';
import { DEPLOYMENT_STATES } from '@northern.tech/store/constants';
import { Deployment } from '@northern.tech/store/deploymentsSlice';
import { Device } from '@northern.tech/store/devicesSlice';
import { useDeploymentDevice } from '@northern.tech/store/useDeploymentDevice';

import { DeploymentDeviceGroup } from '../../deployments/DeploymentItem';
import { DeploymentProgress } from '../../deployments/progress/DeploymentProgress';

const maxWidth = 500;

const useStyles = makeStyles()(theme => ({
  base: {
    alignItems: 'center',
    backgroundColor: theme.palette.divider,
    borderRadius: theme.spacing(),
    columnGap: theme.spacing(2),
    display: 'grid',
    gridTemplateColumns: '200px 1fr',
    maxWidth,
    padding: theme.spacing(3),
    paddingLeft: theme.spacing(1.5),
    '> div:last-child': {
      width: '100%'
    }
  },
  [DEPLOYMENT_STATES.finished]: {
    background: theme.palette.background.paper,
    borderRadius: theme.spacing(0.5),
    padding: theme.spacing(0.5),
    '>.flexbox': {
      justifyContent: 'space-evenly'
    }
  },
  finishedWrapper: { marginTop: theme.spacing(-1.25) },
  wrapper: { display: 'flex', flexDirection: 'column', maxWidth, '> time': { alignSelf: 'flex-end' } }
}));

interface DeploymentWidgetProps {
  deployment: Deployment;
  devicesById: Record<string, Device>;
  idAttribute?: string;
  onClick: (params: { id: string; open: boolean; route: string; tab?: string }) => void;
  state: string;
}

const KeyInfo = ({ deployment, devicesById, idAttribute }: Omit<DeploymentWidgetProps, 'onClick' | 'state'>) => (
  <div>
    <Typography>{deployment.artifact_name}</Typography>
    <DeploymentDeviceGroup deployment={deployment} devicesById={devicesById} idAttribute={idAttribute} wrappingClass="" />
  </div>
);

const BaseDeploymentWidget = ({ deployment, devicesById, idAttribute, onClick, state }: DeploymentWidgetProps) => {
  const { classes } = useStyles();
  useDeploymentDevice(deployment.name);

  const onWidgetClick = () => onClick({ route: 'deployments', id: deployment.id, tab: state === DEPLOYMENT_STATES.finished ? state : undefined, open: true });

  return (
    <div className={`clickable ${classes.base}`} onClick={onWidgetClick}>
      <KeyInfo deployment={deployment} devicesById={devicesById} idAttribute={idAttribute} />
      <div className={state === DEPLOYMENT_STATES.finished ? classes.finished : ''}>
        <DeploymentProgress deployment={deployment} variant="dashboard" />
      </div>
    </div>
  );
};

export const BaseDeploymentsWidget = ({ deployments, ...props }: Omit<DeploymentWidgetProps, 'deployment'> & { deployments: Deployment[] }) =>
  deployments.map(deployment => <BaseDeploymentWidget deployment={deployment} key={deployment.id} {...props} />);

export const CompletedDeployments = ({ deployments, ...props }: Omit<DeploymentWidgetProps, 'deployment'> & { deployments: Deployment[] }) => {
  const { classes } = useStyles();
  return (
    <div className={classes.finishedWrapper}>
      {deployments.map(deployment => (
        <div className={`${classes.wrapper} margin-bottom-x-small`} key={deployment.id}>
          <Time className="muted slightly-smaller" value={deployment.finished} />
          <BaseDeploymentWidget deployment={deployment} {...props} />
        </div>
      ))}
    </div>
  );
};
