// Copyright 2016 Northern.tech AS
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
import { Cancel as CancelIcon, Error as FailedIcon, Pending as PendingIcon, Timelapse as ProgressIcon, CheckCircle as SuccessIcon } from '@mui/icons-material';
import { SvgIconOwnProps, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import type { Deployment } from '@northern.tech/store/deploymentsSlice';
import { groupDeploymentStats } from '@northern.tech/store/utils';

const phases: Record<string, { color: SvgIconOwnProps['color']; icon: typeof CancelIcon; title: string }> = {
  skipped: { title: 'Skipped', icon: CancelIcon, color: 'action' },
  pending: { title: 'Pending', icon: PendingIcon, color: 'action' },
  inprogress: { title: 'In progress', icon: ProgressIcon, color: 'action' },
  successes: { title: 'Successful', icon: SuccessIcon, color: 'success' },
  failures: { title: 'Failed', icon: FailedIcon, color: 'error' }
};

const useStyles = makeStyles()(theme => ({
  resultsStatus: {
    columnGap: theme.spacing(),
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, 32px)',
    '> div': {
      columnGap: theme.spacing(0.5)
    }
  }
}));

export const DeploymentStats = ({ deployment = {} as Deployment }: { deployment?: Deployment }) => {
  const { classes } = useStyles();
  const phaseStats = groupDeploymentStats(deployment, true);
  return (
    <div className={`flexbox ${classes.resultsStatus}`}>
      {Object.entries(phases).map(([key, { icon: Icon, color, title }]) => (
        <Tooltip key={key} title={title}>
          <div className="flexbox centered">
            <Icon color={!phaseStats[key] ? 'disabled' : color} />
            <Typography>{phaseStats[key]}</Typography>
          </div>
        </Tooltip>
      ))}
    </div>
  );
};

export default DeploymentStats;
