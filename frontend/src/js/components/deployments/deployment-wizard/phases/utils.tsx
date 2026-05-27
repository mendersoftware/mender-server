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
import type { ReactNode } from 'react';

import type { AlertProps } from '@mui/material';

import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import type { NewDeploymentPhaseTypeManagement as DeploymentPhase, Filter } from '@northern.tech/types/MenderTypes';
import dayjs from 'dayjs';
import durationPlugin from 'dayjs/plugin/duration';
import pluralize from 'pluralize';

import type { RolloutMode, RolloutPattern } from './constants';
import { delayDefaults, delayUnits, phaseDefaults, phaseLimits, rolloutModes, rolloutPatterns } from './constants';

dayjs.extend(durationPlugin);

export interface ActivePhaseComponentProps {
  classes?: Record<string, string>;
  deploymentDeviceCount: number;
  filter?: Filter;
}

// the form tracks phases in a single shape - one batch size, read as a percentage or a device count depending on the
// rollout mode, with only the phases the user actually sized; the closing phase always takes whatever remains and is derived if needed
export type PhaseDefinition = {
  batchSize?: number;
  delay?: number;
  delayUnit?: string;
};

export const delayToSeconds = (delay: number, unit: string): string => `${dayjs.duration(delay, unit as dayjs.ManipulateType).asSeconds()}s`;

const orderedDelayUnits = [delayUnits.days, delayUnits.hours, delayUnits.minutes];

// the interval is stored in seconds, so we look for the largest unit it converts to as a whole number to get back to something readable
export const parseInterval = (interval?: string): { delay: number; delayUnit: string } => {
  if (!interval) {
    return { ...delayDefaults };
  }
  const duration = dayjs.duration(parseInt(interval) || phaseDefaults.delay, 'seconds');
  const totals = {
    [delayUnits.days]: duration.asDays(),
    [delayUnits.hours]: duration.asHours(),
    [delayUnits.minutes]: duration.asMinutes()
  };
  const delayUnit = orderedDelayUnits.find(unit => totals[unit] >= 1 && Number.isInteger(totals[unit]));
  return delayUnit ? { delay: totals[delayUnit], delayUnit } : { delay: Math.max(1, Math.round(duration.asHours())), delayUnit: delayUnits.hours };
};

type SizedPhase = { batch_size?: number; batch_size_devices?: number };

const getRemainderPercent = (phases: SizedPhase[]) =>
  Math.max(
    0,
    phases.slice(0, -1).reduce((accu, { batch_size = 0 }) => accu - batch_size, phaseLimits.fullBatchPercentage)
  );

const getRemainderDevices = (phases: SizedPhase[], numberDevices: number): number =>
  Math.max(
    0,
    phases.slice(0, -1).reduce((accu, { batch_size_devices = 0 }) => accu - batch_size_devices, numberDevices)
  );

// deployments coming back from the api carry their closing phase, so the remainder skips it - unlike for the definitions in the form
export const getRemainder = ({
  phases,
  numberDevices,
  rolloutMode
}: {
  numberDevices: number;
  phases: DeploymentPhase[];
  rolloutMode: RolloutMode;
}): number => {
  if (rolloutMode === rolloutModes.percentage.key) {
    return getRemainderPercent(phases);
  }
  return getRemainderDevices(phases, numberDevices);
};

export const getDefinitionsRemainder = ({
  phases,
  numberDevices,
  rolloutMode
}: {
  numberDevices: number;
  phases: PhaseDefinition[];
  rolloutMode: RolloutMode;
}): number => {
  const capacity = rolloutMode === rolloutModes.percentage.key ? phaseLimits.fullBatchPercentage : numberDevices;
  return Math.max(
    0,
    phases.reduce((accu, { batchSize = 0 }) => accu - batchSize, capacity)
  );
};

export const getPhaseDeviceCount = (numberDevices = 1, batchSize: number, remainder: number, isLastPhase: boolean) => {
  const count = (numberDevices / phaseLimits.fullBatchPercentage) * (batchSize || remainder);
  return isLastPhase ? Math.ceil(count) : Math.floor(count);
};

export const percentageToDevices = (percentage: number, numberDevices: number): number =>
  numberDevices > 0 ? Math.max(1, Math.floor((numberDevices / phaseLimits.fullBatchPercentage) * percentage)) : 0;

export const devicesToPercentage = (devices: number, numberDevices: number): number =>
  numberDevices > 0
    ? Math.max(1, Math.min(phaseLimits.maxPerBatchPercentage, Math.round((devices / numberDevices) * phaseLimits.fullBatchPercentage)))
    : phaseDefaults.batchSize;

export const convertDefinitionsToMode = (phases: PhaseDefinition[], newMode: RolloutMode, numberDevices: number): PhaseDefinition[] =>
  phases.map(({ batchSize, ...phase }) => ({
    ...phase,
    batchSize: !batchSize
      ? batchSize
      : newMode === rolloutModes.device_count.key
        ? percentageToDevices(batchSize, numberDevices)
        : devicesToPercentage(batchSize, numberDevices)
  }));

export type PhaseMessage = {
  message: string | ReactNode;
  severity: AlertProps['severity'];
};

export interface PhaseMessagesProps {
  batchSize?: number;
  deploymentDeviceCount: number;
  isDynamic: boolean;
  isFinal?: boolean;
  maxDevices?: number;
  remainder: number;
  rolloutMode: RolloutMode;
}

const getPercentagePhaseMessages = ({ batchSize, isFinal, remainder, deploymentDeviceCount }: PhaseMessagesProps): PhaseMessage[] => {
  const messages: PhaseMessage[] = [];
  if (!isFinal && batchSize !== undefined && (batchSize < 1 || batchSize > phaseLimits.maxPerBatchPercentage)) {
    messages.push({ message: 'Please enter a value between 1% and 99%', severity: 'error' });
  }
  const effectiveSize = (isFinal ? remainder : batchSize) || 0;
  if (effectiveSize > 0 && Math.floor((deploymentDeviceCount / phaseLimits.fullBatchPercentage) * effectiveSize) < 1) {
    messages.push({ message: `${effectiveSize}% rounds down to 0 devices. Increase the percentage or switch to device count mode.`, severity: 'error' });
  }
  if (!effectiveSize) {
    messages.push({ message: 'Phases must have at least 1 device', severity: 'error' });
  }
  return messages;
};

const getDeviceCountPhaseMessages = ({ batchSize, isDynamic, isFinal, remainder, deploymentDeviceCount, maxDevices }: PhaseMessagesProps): PhaseMessage[] => {
  const messages: PhaseMessage[] = [];
  if (!!batchSize && batchSize > deploymentDeviceCount) {
    if (isDynamic) {
      messages.push({
        message: `Rollout size exceeds the current target group size. Any new devices added to the group will join this phase until it's full`,
        severity: 'warning'
      });
    } else {
      messages.push({ message: 'Rollout size exceeds total target group size', severity: 'error' });
    }
  }
  if (!isFinal && batchSize === 0) {
    messages.push({ message: 'Phases must have at least 1 device', severity: 'error' });
  }
  if (isFinal && remainder < 1 && !isDynamic) {
    messages.push({ message: 'Phases must have at least 1 device', severity: 'error' });
  }
  if (maxDevices && !!batchSize && batchSize > maxDevices && !isDynamic) {
    messages.push({ message: 'Rollout size cannot exceed the maximum number devices', severity: 'error' });
  }
  return messages;
};

export const getPhaseMessages = (props: PhaseMessagesProps): PhaseMessage[] =>
  props.rolloutMode === rolloutModes.percentage.key ? getPercentagePhaseMessages(props) : getDeviceCountPhaseMessages(props);

export const getPhasesMessage = ({
  filter,
  rolloutPattern,
  maxDevices
}: {
  filter?: Filter;
  maxDevices?: number;
  rolloutPattern: RolloutPattern;
}): PhaseMessage | undefined => {
  if (!filter) {
    return;
  }
  if (rolloutPattern === rolloutPatterns.uniform.key && !maxDevices) {
    return {
      message: 'This deployment targets a dynamic device group using a uniform rollout. The deployment remains active until you manually stop it.',
      severity: 'info'
    };
  }
  if (rolloutPattern !== rolloutPatterns.uniform.key && maxDevices) {
    return { message: `This deployment will stop at ${maxDevices} ${pluralize('device', maxDevices)} due to the device limit above`, severity: 'info' };
  }
  return {
    message: (
      <>
        This deployment targets a dynamic device group, so the final phase may adjust as devices change. The last phase stays active to keep all devices
        updated. <DocsTextLink id={DOCSTIPS.dynamicDeployments.id} typographyProps={{ variant: 'inherit' }} />
      </>
    ),
    severity: 'info'
  };
};

const deviceCountFormatter = new Intl.NumberFormat('en-US', { notation: 'compact', roundingMode: 'trunc' });

export const formatDeviceCount = (count: number): string => (Number.isFinite(count) && count >= 0 ? deviceCountFormatter.format(count) : '0');

export interface StandardizedPhase {
  batch_size?: number;
  batch_size_devices?: number;
  delay?: number;
  delayUnit?: string;
  device_count?: number;
  isUniform?: boolean;
  start_ts?: number;
}

const getStoredPhasesMode = (phases: StandardizedPhase[]): RolloutMode =>
  phases.some(phase => phase.hasOwnProperty(rolloutModes.device_count.batchKey) && !!phase.batch_size_devices)
    ? (rolloutModes.device_count.key as RolloutMode)
    : (rolloutModes.percentage.key as RolloutMode);

// stored phases carry their closing remainder phase, the form's definitions don't - so it gets dropped on the way in
export const parsePreviousPhases = (phases: StandardizedPhase[]): { pattern: RolloutPattern; phases: PhaseDefinition[]; rolloutMode: RolloutMode } => {
  const rolloutMode = getStoredPhasesMode(phases);
  const toPhaseDefinition = ({ batch_size, batch_size_devices, delay, delayUnit }: StandardizedPhase): PhaseDefinition => ({
    batchSize: rolloutMode === rolloutModes.percentage.key ? batch_size : batch_size_devices,
    ...(delay ? { delay, delayUnit: delayUnit || delayDefaults.delayUnit } : {})
  });
  if (phases[0]?.isUniform) {
    return { phases: [toPhaseDefinition(phases[0])], pattern: rolloutPatterns.uniform.key as RolloutPattern, rolloutMode };
  }
  return { phases: phases.slice(0, -1).map(toPhaseDefinition), pattern: rolloutPatterns.custom.key as RolloutPattern, rolloutMode };
};

type ReadablePhaseDescriptions = { phasesDescription: string; tooltip: string };

const toUniformPhasesDescription = (phases: StandardizedPhase[], numberDevices: number): ReadablePhaseDescriptions => {
  const isPercentageMode = getStoredPhasesMode(phases) === rolloutModes.percentage.key;
  const { delay, delayUnit, batch_size, batch_size_devices = 0 } = phases[0];
  const prefix = 'Uniform: ';
  let phasesDescription = '';
  if (isPercentageMode) {
    phasesDescription = `${batch_size}% per phase, ${delay} ${delayUnit || delayDefaults.delayUnit} intervals`;
    return { phasesDescription: `${prefix}${phasesDescription}`, tooltip: phasesDescription };
  }
  phasesDescription = `${Math.min(numberDevices, batch_size_devices)} devices per phase, ${delay} ${delayUnit || delayDefaults.delayUnit} intervals`;
  return { phasesDescription: `${prefix}${phasesDescription}`, tooltip: phasesDescription };
};

export const toPhaseDescription = (phases: StandardizedPhase[], numberDevices: number): ReadablePhaseDescriptions => {
  const isPercentageMode = getStoredPhasesMode(phases) === rolloutModes.percentage.key;
  const { isUniform } = phases.length ? phases[0] : {};
  if (isUniform) {
    return toUniformPhasesDescription(phases, numberDevices);
  }
  const prefix = `${phases.length} ${pluralize('phase', phases.length)}: `;
  if (isPercentageMode) {
    const remainder = getRemainderPercent(phases);
    const phasesDescription = phases
      .map((phase, _, source) => `${phase.batch_size || remainder || phaseLimits.fullBatchPercentage / source.length}%`)
      .join(', ');
    const tooltip = phases
      .map(({ delay, delayUnit, batch_size }, _, source) =>
        delay
          ? `${batch_size}% > ${delay} ${delayUnit || delayDefaults.delayUnit}`
          : `${batch_size || remainder || phaseLimits.fullBatchPercentage / source.length}%`
      )
      .join(' > ');
    return { phasesDescription: `${prefix}${phasesDescription}`, tooltip };
  }
  const remainder = getRemainderDevices(phases, numberDevices);
  const phasesDescription = phases.map(phase => phase.batch_size_devices || remainder).join(', ');
  const tooltip = phases
    .map(({ delay, delayUnit, batch_size_devices }) =>
      delay ? `${batch_size_devices} > ${delay} ${delayUnit || delayDefaults.delayUnit}` : batch_size_devices || remainder
    )
    .join(' > ');
  return { phasesDescription: `${prefix}${phasesDescription}`, tooltip };
};

interface PhaseInfoProps {
  index: number;
  isDynamic: boolean;
  maxDevices?: number;
  numberDevices: number;
  phases: PhaseDefinition[];
  rolloutMode: RolloutMode;
}

type PhaseInfo = {
  batchValue?: number;
  deviceCount: number;
  hasError: boolean;
  hasWarning: boolean;
  max: number;
  messages: PhaseMessage[];
};

// rows beyond the sized definitions describe the derived closing phase, which always takes the remainder
export const computePhaseInfo = ({ index, phases, isDynamic, numberDevices, rolloutMode, maxDevices }: PhaseInfoProps): PhaseInfo => {
  const isFinal = index >= phases.length;
  const isPercentageMode = rolloutMode === rolloutModes.percentage.key;
  const remainder = getDefinitionsRemainder({ phases, numberDevices, rolloutMode });
  const batchSize = isFinal ? undefined : phases[index].batchSize;
  const messages = getPhaseMessages({ batchSize, deploymentDeviceCount: numberDevices, isDynamic, isFinal, maxDevices, remainder, rolloutMode });

  const hasError = messages.some(({ severity }) => severity === 'error');
  const hasWarning = messages.some(({ severity }) => severity === 'warning');

  if (isPercentageMode) {
    return {
      batchValue: Math.max(0, isFinal ? remainder : (batchSize ?? 0)),
      deviceCount: getPhaseDeviceCount(numberDevices, isFinal ? 0 : (batchSize ?? 0), remainder, isFinal),
      hasError,
      hasWarning,
      max: remainder,
      messages
    };
  }
  const batchValue = isFinal ? remainder : batchSize;
  return {
    batchValue,
    deviceCount: batchValue || 0,
    hasError,
    hasWarning,
    max: Math.min(numberDevices, Number.MAX_SAFE_INTEGER),
    messages
  };
};
