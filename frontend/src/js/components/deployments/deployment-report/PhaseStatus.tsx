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
import { LinearProgress, Typography } from '@mui/material';

import { DEPLOYMENT_STATES, deploymentDisplayStates } from '@northern.tech/store/constants';
import type { DeploymentPhase } from '@northern.tech/types/MenderTypes';

export const PhaseStatus = ({ phase }: { phase: DeploymentPhase }) => {
  const { device_count = 0, initial_batch_device_count = 0, status = DEPLOYMENT_STATES.pending } = phase;
  if (status !== DEPLOYMENT_STATES.inprogress) {
    return (
      <Typography className="padding-top-small padding-bottom-small" variant="body2">
        {deploymentDisplayStates[status]}
      </Typography>
    );
  }
  const progress = device_count ? Math.min((device_count / initial_batch_device_count) * 100, 100) : 0;
  return (
    <div className="padding-top-x-small padding-bottom-x-small">
      <div className="flexbox space-between">
        <Typography variant="body2">{deploymentDisplayStates[status]}</Typography>
        <Typography variant="body2">{`${device_count.toLocaleString()}/${initial_batch_device_count.toLocaleString()}`}</Typography>
      </div>
      <LinearProgress className="margin-top-x-small" variant="determinate" value={progress} />
    </div>
  );
};

export default PhaseStatus;
