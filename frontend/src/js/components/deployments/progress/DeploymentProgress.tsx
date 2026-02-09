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
import { Deployment } from '@northern.tech/store/deploymentsSlice';
import { getDeploymentState } from '@northern.tech/store/utils';

import DeploymentStats from '../DeploymentStatus';
import { DeploymentStatusNotification } from './DeploymentStatusNotification';
import { ProgressVariant, RolloutProgressBar, SimpleProgress } from './RolloutProgressBar';
import { SubstateProgressBar } from './SubstateProgressBar';

interface ReportProgressProps {
  className?: string;
  deployment: Deployment;
  onAbort: (id: string) => void;
  onUpdateControlChange: (update: { states: Record<string, { action: string }> }) => void;
}

const ReportProgress = ({ className, deployment, onAbort, onUpdateControlChange }: ReportProgressProps) => {
  const { phases = [], update_control_map } = deployment;
  if (phases.length > 1 || !update_control_map) {
    return <RolloutProgressBar className={className} deployment={deployment} variant="report" />;
  }
  return <SubstateProgressBar className={className} deployment={deployment} onAbort={onAbort} onUpdateControlChange={onUpdateControlChange} />;
};

/**
 * Smart router component that determines which progress display to show based on deployment state and variant.
 *
 * Variants:
 * - 'dashboard': Simple progress bar or deployment stats for finished deployments
 * - 'list': Full rollout progress with header, footer, and side panel
 * - 'report': Detailed progress view, either rollout phases or substate tracking
 */

interface DeploymentProgressProps {
  className?: string;
  deployment: Deployment;
  onAbort?: (id: string) => void;
  onUpdateControlChange?: (update: { states: Record<string, { action: string }> }) => void;
  variant: ProgressVariant;
}

export const DeploymentProgress = ({ className, deployment, variant, onAbort, onUpdateControlChange }: DeploymentProgressProps) => {
  if (variant === 'report' && onAbort && onUpdateControlChange) {
    return <ReportProgress className={className} deployment={deployment} onAbort={onAbort} onUpdateControlChange={onUpdateControlChange} />;
  }

  const status = getDeploymentState(deployment);

  if (status === 'queued') {
    return <DeploymentStatusNotification status={status} />;
  }
  if (status === deploymentDisplayStates.finished) {
    return <DeploymentStats deployment={deployment} />;
  }
  if (variant === 'list') {
    return <RolloutProgressBar className={className} deployment={deployment} variant="list" />;
  }

  return <SimpleProgress deployment={deployment} />;
};
