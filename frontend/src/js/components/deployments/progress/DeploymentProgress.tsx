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
import { deploymentDisplayStates } from '@northern.tech/store/constants';
import type { Deployment } from '@northern.tech/store/deploymentsSlice';
import { getDeploymentState } from '@northern.tech/store/utils';

import DeploymentStats from '../DeploymentStatus';
import { DeploymentStatusNotification } from './DeploymentStatusNotification';
import { ListProgress, SimpleProgress } from './RolloutProgressBar';
import type { ProgressVariant } from './RolloutProgressBar';

/**
 * Smart router component that determines which progress display to show based on deployment state and variant.
 *
 * Variants:
 * - 'dashboard': Simple progress bar or deployment stats for finished deployments
 * - 'list': Full rollout progress with header, footer, and side panel
 */
interface DeploymentProgressProps {
  className?: string;
  deployment: Deployment;
  variant?: ProgressVariant;
}

export const DeploymentProgress = ({ className, deployment, variant }: DeploymentProgressProps) => {
  const status = getDeploymentState(deployment);

  if (status === 'queued') {
    return <DeploymentStatusNotification status={status} />;
  }
  if (status === deploymentDisplayStates.finished) {
    return <DeploymentStats deployment={deployment} />;
  }
  if (variant === 'dashboard') {
    return <SimpleProgress deployment={deployment} />;
  }
  return <ListProgress className={className} deployment={deployment} />;
};
