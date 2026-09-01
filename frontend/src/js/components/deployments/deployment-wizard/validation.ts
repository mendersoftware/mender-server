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

import { maxDeploymentRetries } from './RolloutOptions';
import { rolloutModes, rolloutPatterns } from './phases/constants';
import type { PhaseDefinition, PhaseMessagesProps } from './phases/utils';
import { getDefinitionsRemainder, getPhaseMessages } from './phases/utils';
import type { DeploymentFormValues } from './types';
import type { DeploymentDerivedState } from './utils';

export const deploymentErrors = {
  emptyGroupSchedule: 'Cannot schedule deployment for an empty device group. Please select a different start time or choose a group with devices',
  group: 'Please select a device group to target',
  maxDevicesRequired: 'Number of devices is required',
  numberRange: 'Please enter a valid number.',
  release: 'Please select software to deploy'
};

export const disabledReasons = {
  allDevicesLimit: 'Cannot limit device count when targeting all devices',
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

const isTargetingEmptyGroup = ({ deploymentDeviceCount, group, isDeviceCountResolved }: TargetState) =>
  !!group && !!isDeviceCountResolved && !deploymentDeviceCount;

export const getDeviceLimitDisabledReason = ({ devices = [], filter, group }: TargetState) => {
  if (filter) {
    return '';
  }
  if (group === ALL_DEVICES) {
    return disabledReasons.allDevicesLimit;
  }
  if (group) {
    return disabledReasons.staticGroupLimit;
  }
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

type PhaseErrorContext = Pick<PhaseMessagesProps, 'deploymentDeviceCount' | 'isDynamic' | 'maxDevices' | 'rolloutMode'>;

const getDefinitionErrors = (phases: Array<PhaseDefinition | undefined>, remainder: number, context: PhaseErrorContext) =>
  phases.reduce(
    (accu, definition, phaseIndex) => [
      ...accu,
      ...getPhaseMessages({ batchSize: definition?.batchSize, isFinal: definition === undefined, remainder, ...context })
        .filter(({ message, severity }) => severity === 'error' && typeof message === 'string')
        .map(({ message }) => ({ message: message as string, phaseIndex }))
    ],
    [] as Array<{ message: string; phaseIndex: number }>
  );

const getPhaseFieldErrors = (
  { phases = [], rolloutPattern, usesPattern }: Pick<DeploymentFormValues, 'phases' | 'rolloutPattern' | 'usesPattern'>,
  context: PhaseErrorContext
): Record<string, { message: string; type: string }> => {
  if (!usesPattern || !phases.length) {
    return {};
  }
  // for a uniform rollout only the first single phase will be considered & validated
  if (rolloutPattern === rolloutPatterns.uniform.key) {
    const [definition] = phases;
    const remainder = getDefinitionsRemainder({ phases: [definition], numberDevices: context.deploymentDeviceCount, rolloutMode: context.rolloutMode });
    const [uniformError] = getDefinitionErrors([definition], remainder, context);
    return uniformError ? { phases: { message: uniformError.message, type: 'validate' } } : {};
  }
  const remainder = getDefinitionsRemainder({ phases, numberDevices: context.deploymentDeviceCount, rolloutMode: context.rolloutMode });
  // validation considers also the derived final phase values
  const [phaseError] = getDefinitionErrors([...phases, undefined], remainder, context);
  return phaseError ? { phases: { message: `Phase ${phaseError.phaseIndex + 1}: ${phaseError.message}`, type: 'validate' } } : {};
};

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
  if (values.shouldLimit && !values.maxDevices) {
    errors.maxDevices = { message: deploymentErrors.maxDevicesRequired, type: 'required' };
  } else if (values.shouldLimit && !(Number.isInteger(values.maxDevices) && Number(values.maxDevices) >= 1)) {
    errors.maxDevices = { message: deploymentErrors.numberRange, type: 'validate' };
  }
  if (values.retries && !(Number.isInteger(values.retries) && Number(values.retries) > 0 && Number(values.retries) <= maxDeploymentRetries)) {
    errors.retries = { message: deploymentErrors.numberRange, type: 'validate' };
  }
  Object.assign(
    errors,
    getPhaseFieldErrors(values, {
      deploymentDeviceCount,
      isDynamic: !!filter,
      maxDevices: values.maxDevices,
      rolloutMode: values.rolloutMode ?? rolloutModes.percentage.key
    })
  );
  return { errors, values: isEmpty(errors) ? values : {} };
};
