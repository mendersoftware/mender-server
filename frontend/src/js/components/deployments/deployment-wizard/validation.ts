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
import type { Resolver } from 'react-hook-form';

import { ALL_DEVICES } from '@northern.tech/utils/constants';
import { isEmpty } from '@northern.tech/utils/helpers';

import type { DeploymentFormValues } from './types';
import type { DeploymentDerivedState } from './utils';

export const deploymentErrors = {
  emptyGroupSchedule: 'Cannot schedule deployment for an empty device group. Please select a different start time or choose a group with devices',
  group: 'Please select a device group to target',
  maxDevices: 'Number of devices is required',
  phases: 'Each phase has to contain at least 1 device and the phases may not exceed 100% in total',
  release: 'Please select software to deploy'
};

export const disabledReasons = {
  deviceTargetLimit: 'Cannot limit device count when targeting individual devices',
  emptyGroupPattern: 'Rollout pattern is not available for empty device groups',
  emptyGroupPauses: 'Pauses is not available for empty device groups',
  pausedPattern: 'Cannot select rollout pattern when pauses are enabled',
  patternPauses: 'Cannot add pauses when using a rollout pattern',
  staticGroupLimit: 'Cannot limit device count when targeting a static group'
};

type TargetState = Pick<DeploymentDerivedState, 'deploymentDeviceCount' | 'filter'> & {
  devices?: DeploymentDerivedState['devices'];
  group?: string | null;
  isDeviceCountResolved?: boolean;
};

// the advanced options start out available - they only become unavailable once the selected target makes them
// meaningless, or once one of the mutually exclusive rollout options is in use; while the device count of a target is
// still being retrieved it can't rule anything out yet
const isTargetingEmptyGroup = ({ deploymentDeviceCount, group, isDeviceCountResolved }: TargetState) =>
  !!group && !!isDeviceCountResolved && !deploymentDeviceCount;

export const getDeviceLimitDisabledReason = ({ devices = [], filter, group }: TargetState) => {
  if (filter) {
    return '';
  }
  if (group && group !== ALL_DEVICES) {
    return disabledReasons.staticGroupLimit;
  }
  // the api only supports limiting deployments to dynamic groups, a limit on directly targeted devices would
  // silently be ignored
  return devices.length ? disabledReasons.deviceTargetLimit : '';
};

export const getRolloutPatternDisabledReason = ({ isPaused, ...target }: TargetState & { isPaused?: boolean }) => {
  if (isTargetingEmptyGroup(target)) {
    return disabledReasons.emptyGroupPattern;
  }
  return isPaused ? disabledReasons.pausedPattern : '';
};

export const getPausesDisabledReason = ({ usesPattern, ...target }: TargetState & { usesPattern?: boolean }) => {
  if (isTargetingEmptyGroup(target)) {
    return disabledReasons.emptyGroupPauses;
  }
  return usesPattern ? disabledReasons.patternPauses : '';
};

export type DeploymentResolverContext = Pick<DeploymentDerivedState, 'deploymentDeviceCount' | 'devices' | 'filter'> & { group: string | null };

// the target device count & the preselected devices live outside of the form, so the validation has to run as a
// resolver with them handed in as context instead of as rules on the individual fields
export const deploymentResolver: Resolver<DeploymentFormValues, DeploymentResolverContext> = (values, context) => {
  const { deploymentDeviceCount = 0, devices = [], filter, group } = context ?? {};
  const errors: Record<string, { message: string; type: string }> = {};
  if (!values.group && !devices.length) {
    errors.group = { message: deploymentErrors.group, type: 'required' };
  }
  if (!values.release) {
    errors.release = { message: deploymentErrors.release, type: 'required' };
  }
  if (values.startTime && !deploymentDeviceCount && group && !filter) {
    errors.startTime = { message: deploymentErrors.emptyGroupSchedule, type: 'validate' };
  }
  if (values.shouldLimit && !(Number(values.maxDevices) >= 1)) {
    errors.maxDevices = { message: deploymentErrors.maxDevices, type: 'required' };
  }
  return { errors, values: isEmpty(errors) ? values : {} };
};
