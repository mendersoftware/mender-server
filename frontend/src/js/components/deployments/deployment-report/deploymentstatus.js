// Copyright 2021 Northern.tech AS
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

import { Pause as PauseIcon, ArrowDropDownCircleOutlined as ScrollDownIcon } from '@mui/icons-material';
import { makeStyles } from 'tss-react/mui';

import { deploymentDisplayStates, pauseMap } from '../../../constants/deploymentConstants';
import { groupDeploymentStats } from '../../../helpers';
import { TwoColumnData } from '../../common/configurationobject';
import { defaultColumnDataProps } from '../report';

const useStyles = makeStyles()(theme => ({
  progressStatus: {
    backgroundColor: theme.palette.background.light,
    padding: `${theme.spacing(2)} ${theme.spacing(8)} ${theme.spacing(2)}`
  },
  scrollDown: { marginLeft: theme.spacing() }
}));

export const DeploymentPhaseNotification = ({ className = '', deployment = {}, onReviewClick }) => {
  const { classes } = useStyles();
  const { paused } = groupDeploymentStats(deployment);
  if (paused === 0) {
    return null;
  }
  return (
    <div className={`${classes.progressStatus} flexbox center-aligned margin-bottom clickable ${className}`} onClick={onReviewClick} style={{ padding: 15 }}>
      <PauseIcon />
      <div className="muted">
        Deployment is <span className="uppercased">paused</span>. <a>Review its status</a> to continue, retry or abort the deployment{' '}
      </div>
      <ScrollDownIcon fontSize="small" className={`link-color ${classes.scrollDown}`} />
    </div>
  );
};

export const DeploymentStatus = ({ className = '', deployment = {} }) => {
  const { classes } = useStyles();
  const { finished, max_devices, retries = 1, status = 'pending', statistics = {} } = deployment;
  const { status: stats = {} } = statistics;
  const phaseStats = groupDeploymentStats(deployment, true);

  let statusDescription = (
    <>
      {deploymentDisplayStates[status]}
      {status === 'pending' ? ' (awaiting devices)' : ''}
    </>
  );
  if (finished) {
    statusDescription = <div>Finished {!!phaseStats.failure && <span className="failures">with failures</span>}</div>;
  } else if (status === 'paused' && phaseStats.paused > 0) {
    // based on the order of the possible pause states we find the furthest possible and use that as the current pause state - if applicable
    const currentPauseState = Object.keys(pauseMap)
      .reverse()
      .find(key => stats[key] > 0);
    statusDescription = (
      <>
        {deploymentDisplayStates[status]} ({pauseMap[currentPauseState].title})
      </>
    );
  }

  const statsBasedDeviceCount = Object.values(phaseStats).reduce((sum, count) => sum + count, 0);
  // eslint-disable-next-line no-unused-vars
  const { failure, finished: finishedDeployment, scheduled, success, ...phasesWithStats } = deploymentDisplayStates;

  return (
    <div className={`${classes.progressStatus} flexbox space-between centered margin-bottom ${className}`}>
      <div className="flexbox column">
        <div className="muted">Status</div>
        <h4 className="margin-bottom-none muted">{statusDescription}</h4>
      </div>
      <div className="flexbox space-between align-right" style={{ minWidth: '40%' }}>
        <div className="flexbox column">
          <div className="muted margin-bottom-small"># devices</div>
          <div>{statsBasedDeviceCount}</div>
        </div>
        {Object.entries(phasesWithStats).map(([key, phase]) => (
          <div key={key} className="flexbox column">
            <div className="muted margin-bottom-small">{phase}</div>
            <div className="status">{phaseStats[key].toLocaleString()}</div>
          </div>
        ))}
      </div>
      <TwoColumnData
        {...defaultColumnDataProps}
        config={{ 'Max attempts per device': retries, 'Maximum number of devices': max_devices || 'N/A' }}
        style={{ ...defaultColumnDataProps.style, gridTemplateColumns: 'max-content 1fr' }}
      />
    </div>
  );
};

export default DeploymentStatus;
