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
import React from 'react';

import { Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import errorImage from '../../../assets/img/error_status.png';
import pendingImage from '../../../assets/img/pending_status.png';
import inprogressImage from '../../../assets/img/progress_status.png';
import skippedImage from '../../../assets/img/skipped_status.png';
import successImage from '../../../assets/img/success_status.png';
import { groupDeploymentStats } from '../../helpers';

const phases = {
  skipped: { title: 'Skipped', image: skippedImage },
  pending: { title: 'Pending', image: pendingImage },
  inprogress: { title: 'In progress', image: inprogressImage },
  successes: { title: 'Successful', image: successImage },
  failures: { title: 'Failed', image: errorImage }
};

const useStyles = makeStyles()(theme => ({
  resultsStatus: {
    columnGap: theme.spacing(),
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, 32px)',
    '> div': {
      columnGap: theme.spacing(0.5)
    },
    '.disabled': {
      opacity: '0.1'
    }
  }
}));

export const DeploymentStats = ({ deployment = {} }) => {
  const { classes } = useStyles();
  const phaseStats = groupDeploymentStats(deployment, true);
  return (
    <div className={`flexbox ${classes.resultsStatus} `}>
      {Object.entries(phases).map(([key, phase]) => (
        <Tooltip key={key} title={phase.title}>
          <div className={`flexbox centered ${phaseStats[key] ? '' : 'disabled'}`}>
            <img src={phase.image} />
            <div className="status">{phaseStats[key]}</div>
          </div>
        </Tooltip>
      ))}
    </div>
  );
};

export default DeploymentStats;
