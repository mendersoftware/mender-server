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
import { Warning as WarningIcon } from '@mui/icons-material';
import { LinearProgress, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { deploymentDisplayStates } from '@northern.tech/store/constants';
import type { Deployment } from '@northern.tech/store/deploymentsSlice';
import { getDeploymentState, groupDeploymentStats } from '@northern.tech/store/utils';
import pluralize from 'pluralize';
import { isUUID } from 'validator';

import { DeploymentStatusNotification } from './DeploymentStatusNotification';

export type ProgressVariant = 'dashboard' | 'list';

const useStyles = makeStyles()(theme => ({
  container: {
    backgroundColor: theme.palette.background.default,
    padding: '10px 20px',
    borderRadius: theme.spacing(0.5),
    justifyContent: 'space-evenly',
    minHeight: 70
  },
  dualPanel: {
    display: 'grid',
    gridTemplateColumns: '2fr 1fr',
    gridColumnGap: theme.spacing(2),
    alignItems: 'center'
  }
}));

const determinateStates = {
  [deploymentDisplayStates.finished]: { variant: 'determinate' as const, value: 100 },
  queued: { variant: 'determinate' as const, value: 0 },
  default: { variant: 'indeterminate' as const, value: undefined }
};

const Failures = ({ failures }: { failures: number }) => (
  <Typography variant="body2" className={`flexbox align-items-center ${failures ? 'warning' : ''}`} style={{ justifyContent: 'flex-end' }}>
    {!!failures && <WarningIcon style={{ fontSize: 16, marginRight: 10 }} />}
    {`${failures} ${pluralize('failure', failures)}`}
  </Typography>
);

export const SimpleProgress = ({ deployment }: { deployment: Deployment }) => {
  const { phases = [] } = deployment;
  const { failures } = groupDeploymentStats(deployment, false);
  const status = getDeploymentState(deployment);
  const phaseFailures = phases.reduce((accu, phase) => {
    const { failures = 0 } = phase as { failures?: number };
    return accu + failures;
  }, 0);
  const variantProps = determinateStates[status] ?? determinateStates.default;

  return <LinearProgress color={failures || phaseFailures ? 'secondary' : 'primary'} {...variantProps} />;
};

export const ListProgress = ({ className = '', deployment }: { className?: string; deployment: Deployment }) => {
  const { classes } = useStyles();
  const { phases = [], current_phase = 1, phase_count } = deployment;
  const status = getDeploymentState(deployment);
  const { failures } = groupDeploymentStats(deployment, false);

  let currentPhase = current_phase;
  if (isUUID(`${currentPhase}`) && phases.length) {
    currentPhase = phases.findIndex(({ id }) => id === currentPhase) + 1;
  }

  return (
    <div className={`relative flexbox column ${classes.container} ${className}`}>
      <DeploymentStatusNotification status={status} />
      <div className={classes.dualPanel}>
        <SimpleProgress deployment={deployment} />
        <Failures failures={failures} />
      </div>
      <div className="flexbox space-between">
        <Typography variant="caption">Devices in progress</Typography>
        <Typography variant="caption">{`Current phase: ${currentPhase} of ${phase_count || phases.length || 1}`}</Typography>
      </div>
    </div>
  );
};
