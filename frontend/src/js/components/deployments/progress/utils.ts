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
import type { Deployment } from '@northern.tech/store/deploymentsSlice';
import type { DeploymentPhase } from '@northern.tech/types/MenderTypes';

export type DisplayablePhase = DeploymentPhase & { isFinal: boolean; isGrowing: boolean };

export const getDeploymentStartTime = ({ created, phases, uniform_phases }: Deployment): string | undefined =>
  uniform_phases?.start_ts || phases?.[0]?.start_ts || created;

// the backend does not reliably return positional phase ids (some are uuids), so the phase index is used as the id instead
export const getDisplayablePhases = (deployment: Deployment): DisplayablePhase[] => {
  const { created, device_count, filter, max_devices, phases = [], status } = deployment;
  // uncapped deployments to dynamic groups keep the last phase open for devices joining the group later on
  const isDynamicUncapped = !!filter?.name && !max_devices;
  const sourcePhases: DeploymentPhase[] = phases.length
    ? phases
    : [{ device_count, initial_batch_device_count: device_count, status: status as DeploymentPhase['status'], start_ts: created }];
  return sourcePhases.map((phase, index) => {
    const isLast = index === sourcePhases.length - 1;
    return { ...phase, id: `${index + 1}`, isFinal: sourcePhases.length > 1 && isLast, isGrowing: isDynamicUncapped && isLast };
  });
};
