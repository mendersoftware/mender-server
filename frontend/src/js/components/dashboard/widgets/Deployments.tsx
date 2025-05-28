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
import { makeStyles } from 'tss-react/mui';

import Time from '@northern.tech/common-ui/Time';
import { DEPLOYMENT_STATES } from '@northern.tech/store/constants';
import { useDeploymentDevice } from '@northern.tech/store/useDeploymentDevice';

import { DeploymentDeviceGroup, DeploymentProgress } from '../../deployments/DeploymentItem';
import DeploymentStats from '../../deployments/DeploymentStatus';
import { DeploymentStatusNotification } from '../../deployments/ProgressChart';

const maxWidth = 500;

const useStyles = makeStyles()(theme => ({
  base: {
    alignItems: 'center',
    backgroundColor: theme.palette.grey[400],
    borderRadius: theme.spacing(),
    columnGap: theme.spacing(2),
    display: 'grid',
    gridTemplateColumns: '200px 1fr',
    maxWidth,
    padding: theme.spacing(),
    '> div:last-child': {
      width: '100%'
    }
  },
  [DEPLOYMENT_STATES.finished]: { background: theme.palette.background.paper, borderRadius: theme.spacing(0.5), padding: theme.spacing(0.5) },
  wrapper: { display: 'flex', flexDirection: 'column', maxWidth, '> time': { alignSelf: 'flex-end', marginRight: 6 } }
}));

const KeyInfo = ({ deployment, devicesById, idAttribute }) => (
  <div>
    <div>{deployment.artifact_name}</div>
    <DeploymentDeviceGroup deployment={deployment} devicesById={devicesById} idAttribute={idAttribute} wrappingClass="" />
  </div>
);

const deploymentStateComponentMap = {
  [DEPLOYMENT_STATES.pending]: () => <DeploymentStatusNotification status="queued" />,
  [DEPLOYMENT_STATES.inprogress]: ({ deployment }) => <DeploymentProgress deployment={deployment} minimal />,
  [DEPLOYMENT_STATES.finished]: ({ classes, deployment }) => (
    <div className={classes.finished}>
      <DeploymentStats deployment={deployment} />
    </div>
  )
};

const BaseDeploymentWidget = ({ deployment, devicesById, idAttribute, onClick, state }) => {
  const { classes } = useStyles();
  useDeploymentDevice(deployment.name);

  const Component = deploymentStateComponentMap[state];
  const onWidgetClick = () => onClick({ route: 'deployments', id: deployment.id, tab: state === DEPLOYMENT_STATES.finished ? state : undefined, open: true });

  return (
    <div className={`clickable ${classes.base}`} onClick={onWidgetClick}>
      <KeyInfo deployment={deployment} devicesById={devicesById} idAttribute={idAttribute} />
      <Component classes={classes} deployment={deployment} />
    </div>
  );
};

export const BaseDeploymentsWidget = ({ deployments, ...props }) =>
  deployments.map(deployment => <BaseDeploymentWidget deployment={deployment} key={deployment.id} {...props} />);

export const CompletedDeployments = ({ deployments, ...props }) => {
  const { classes } = useStyles();
  return deployments.map(deployment => (
    <div className={classes.wrapper} key={deployment.id}>
      <Time className="muted slightly-smaller" value={deployment.finished} />
      <BaseDeploymentWidget deployment={deployment} {...props} />
    </div>
  ));
};
