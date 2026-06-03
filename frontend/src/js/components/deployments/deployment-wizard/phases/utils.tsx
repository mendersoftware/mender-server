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
import { alpha } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import type { NewDeploymentPhaseTypeManagement as DeploymentPhase, Filter } from '@northern.tech/types/MenderTypes';
import pluralize from 'pluralize';

import type { RolloutMode, RolloutPattern } from './constants';
import { delayDefaults, rolloutModes, rolloutPatterns } from './constants';

export const useRowStyles = makeStyles()(theme => ({
  rowError: { backgroundColor: alpha(theme.palette.error.main, theme.palette.action.selectedOpacity) },
  rowWarning: { backgroundColor: alpha(theme.palette.warning.main, theme.palette.action.selectedOpacity) }
}));

const getRemainderPercent = (phases: DeploymentPhase[]) => {
  const percentage = phases.reduce((accu, phase, index, source) => {
    if (index === source.length - 1) {
      return accu;
    }
    return phase.batch_size ? accu - phase.batch_size : accu;
  }, 100);
  return Math.max(0, percentage);
};

const getRemainderDevices = (phases: DeploymentPhase[], numberDevices: number): number => {
  const count =
    numberDevices -
    phases.reduce((accu, phase, index, source) => {
      if (index === source.length - 1) {
        return accu;
      }
      return accu + (phase.batch_size_devices || 0);
    }, 0);
  return Math.max(0, count);
};

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

export const getPhaseDeviceCount = (numberDevices = 1, batchSize: number, remainder: number, isLastPhase: boolean) => {
  const count = (numberDevices / 100) * (batchSize || remainder);
  return isLastPhase ? Math.ceil(count) : Math.floor(count);
};

export const percentageToDevices = (percentage: number, numberDevices: number): number =>
  numberDevices > 0 ? Math.max(1, Math.floor((numberDevices / 100) * percentage)) : 0;

export const devicesToPercentage = (devices: number, numberDevices: number): number =>
  numberDevices > 0 ? Math.max(1, Math.min(99, Math.round((devices / numberDevices) * 100))) : 10;

export type PhaseMessage = {
  message: string | ReactNode;
  severity: AlertProps['severity'];
};

interface GetPhaseMessagesBaseProps {
  deploymentDeviceCount: number;
  isLast: boolean;
  maxDevices?: number;
  phase: DeploymentPhase;
  remainder: number;
}

const getPercentagePhaseMessages = ({ phase, isLast, remainder, deploymentDeviceCount }: GetPhaseMessagesBaseProps): PhaseMessage[] => {
  const messages: PhaseMessage[] = [];
  const { batch_size: batchSize } = phase;
  if (batchSize != null && !isLast && (batchSize < 1 || batchSize > 99)) {
    messages.push({ message: 'Please enter a value between 1% and 99%', severity: 'error' });
  }
  const effectiveSize = batchSize || remainder;
  if (effectiveSize > 0 && Math.floor((deploymentDeviceCount / 100) * effectiveSize) < 1) {
    messages.push({ message: `${effectiveSize}% rounds down to 0 devices. Increase the percentage or switch to device count mode.`, severity: 'error' });
  }
  if (!effectiveSize) {
    messages.push({ message: 'Phases must have at least 1 device', severity: 'error' });
  }
  return messages;
};

const getDeviceCountPhaseMessages = ({
  phase,
  isDynamic = false,
  isLast,
  remainder,
  deploymentDeviceCount,
  phasesLength,
  maxDevices
}: GetPhaseMessagesBaseProps & {
  isDynamic: boolean;
  phasesLength: number;
}): PhaseMessage[] => {
  const messages: PhaseMessage[] = [];
  const { batch_size_devices: batchDevices } = phase;
  if (batchDevices > deploymentDeviceCount) {
    if (isDynamic) {
      messages.push({
        message: `Rollout size exceeds the current target group size. Any new devices added to the group will join this phase until it's full`,
        severity: 'warning'
      });
    } else {
      messages.push({ message: 'Rollout size exceeds total target group size', severity: 'error' });
    }
  }

  if (!isLast && batchDevices === 0) {
    messages.push({ message: 'Phases must have at least 1 device', severity: 'error' });
  }
  if (isLast && remainder < 1 && phasesLength) {
    messages.push({ message: 'Phases must have at least 1 device', severity: 'error' });
  }
  if (maxDevices && batchDevices !== null && batchDevices > maxDevices && !isDynamic) {
    messages.push({ message: 'Rollout size cannot exceed the maximum number devices', severity: 'error' });
  }
  return messages;
};

export const getPhaseMessages = ({
  isDynamic,
  phases,
  phaseIndex,
  deploymentDeviceCount,
  rolloutMode,
  maxDevices
}: {
  deploymentDeviceCount: number;
  isDynamic: boolean;
  maxDevices?: number;
  phaseIndex: number;
  phases: DeploymentPhase[];
  rolloutMode: RolloutMode;
}): PhaseMessage[] => {
  if (!phases?.length) {
    return [];
  }
  const isPercentage = rolloutMode === rolloutModes.percentage.key;
  const remainder = isPercentage ? getRemainderPercent(phases) : getRemainderDevices(phases, deploymentDeviceCount);

  const isLast = phaseIndex === phases.length - 1;
  const phaseMessages = isPercentage
    ? getPercentagePhaseMessages({ phase: phases[phaseIndex], isLast, remainder, deploymentDeviceCount })
    : getDeviceCountPhaseMessages({ phase: phases[phaseIndex], isDynamic, isLast, remainder, deploymentDeviceCount, phasesLength: phases.length, maxDevices });
  return phaseMessages;
};

export const getPhasesMessage = ({
  filter,
  rolloutPattern,
  maxDevices
}: {
  filter?: Filter;
  maxDevices: number;
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

const deviceCountThresholds = {
  million: 1_000_000,
  tenThousand: 10_000,
  oneThousand: 1_000
};

const toFixedWithoutRounding = (number: number) => (Math.trunc(number * 10) / 10).toFixed(1).replace(/\.0$/, '');

export const formatDeviceCount = (count: number): string => {
  if (!Number.isFinite(count) || count < 0) return '0';
  if (count >= deviceCountThresholds.million) {
    const number = count / deviceCountThresholds.million;
    return number < 10 ? `${toFixedWithoutRounding(number)}M` : `${Math.floor(number)}M`;
  }
  if (count >= deviceCountThresholds.tenThousand) return `${Math.floor(count / deviceCountThresholds.oneThousand)}K`;
  if (count >= deviceCountThresholds.oneThousand) {
    const number = count / deviceCountThresholds.oneThousand;
    return `${toFixedWithoutRounding(number)}K`;
  }
  return count.toLocaleString();
};

export interface StandardizedPhase {
  batch_size?: number;
  batch_size_devices?: number;
  delay?: number;
  delayUnit?: string;
  device_count?: number;
  isUniform?: boolean;
  start_ts?: number;
}

export type UiDeploymentPhase = DeploymentPhase & StandardizedPhase;

type ReadablePhaseDescriptions = { phasesDescription: string; tooltip: string };

const toUniformPhasesDescription = (phases: StandardizedPhase[], numberDevices: number): ReadablePhaseDescriptions => {
  const isPercentageMode = phases.some(phase => phase.hasOwnProperty(rolloutModes.percentage.batchKey));
  const { delay, delayUnit, batch_size, batch_size_devices } = phases[0];
  const prefix = 'Uniform: ';
  let phasesDescription = '';
  if (isPercentageMode) {
    phasesDescription = `${batch_size}% per phase, ${delay}${delayUnit || delayDefaults.delayUnit} intervals`;
    return { phasesDescription: `${prefix}${phasesDescription}`, tooltip: phasesDescription };
  }
  phasesDescription = `${Math.min(numberDevices, batch_size_devices!)} devices per phase, ${delay}${delayUnit || delayDefaults.delayUnit} intervals`;
  return { phasesDescription: `${prefix}${phasesDescription}`, tooltip: phasesDescription };
};

export const toPhaseDescription = (phases: StandardizedPhase[], numberDevices: number): ReadablePhaseDescriptions => {
  const isPercentageMode = phases.some(phase => phase.hasOwnProperty(rolloutModes.percentage.batchKey));
  const { isUniform } = phases.length ? phases[0] : {};
  if (isUniform) {
    return toUniformPhasesDescription(phases, numberDevices);
  }
  const prefix = `${phases.length} ${pluralize('phase', phases.length)}: `;
  if (isPercentageMode) {
    const remainder = getRemainderPercent(phases);
    const phasesDescription = phases.map((phase, _, source) => `${phase.batch_size || remainder || 100 / source.length}%`).join(', ');
    const tooltip = phases
      .map(({ delay, delayUnit, batch_size }, _, source) =>
        delay ? `${batch_size}% > ${delay} ${delayUnit || delayDefaults.delayUnit} >` : `${batch_size || remainder || 100 / source.length}%`
      )
      .join(', ');
    return { phasesDescription: `${prefix}${phasesDescription}`, tooltip };
  }
  const remainder = getRemainderDevices(phases, numberDevices);
  const phasesDescription = phases.map(phase => phase.batch_size_devices || remainder).join(', ');
  const tooltip = phases
    .map(({ delay, delayUnit, batch_size_devices }) =>
      delay ? `${batch_size_devices} > ${delay} ${delayUnit || delayDefaults.delayUnit} >` : batch_size_devices || remainder
    )
    .join(', ');
  return { phasesDescription: `${prefix}${phasesDescription}`, tooltip };
};

interface PhaseInfoProps {
  index: number;
  isDynamic: boolean;
  maxDevices: number;
  numberDevices: number;
  phases: Array<DeploymentPhase>;
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

export const computePhaseInfo = ({ index, phases, isDynamic, numberDevices, rolloutMode, maxDevices }: PhaseInfoProps): PhaseInfo => {
  const phase = phases[index];
  const isLast = index === phases.length - 1;
  const isPercentageMode = rolloutMode === rolloutModes.percentage.key;
  const messages = getPhaseMessages({ isDynamic, phases, phaseIndex: index, deploymentDeviceCount: numberDevices, rolloutMode, maxDevices });

  const { hasError, hasWarning } = messages.reduce(
    (accu, { severity }) => ({ hasError: accu.hasError || severity === 'error', hasWarning: accu.hasWarning || severity === 'warning' }),
    { hasError: false, hasWarning: false }
  );

  const remainder = getRemainder({ phases, numberDevices, rolloutMode });

  if (isPercentageMode) {
    const batchValue = Math.max(0, isLast ? remainder : phase.batch_size);
    const deviceCount = getPhaseDeviceCount(numberDevices, phase.batch_size, remainder, isLast);
    return {
      batchValue,
      deviceCount,
      hasError,
      hasWarning,
      max: remainder,
      messages
    };
  }
  const batchValue = isLast ? remainder : phase.batch_size_devices;
  const max = numberDevices > 0 ? numberDevices : Number.MAX_SAFE_INTEGER;
  const deviceCount = batchValue || (isLast ? remainder : 0);
  return {
    batchValue,
    deviceCount,
    hasError,
    hasWarning,
    max,
    messages
  };
};
