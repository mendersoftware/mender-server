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
import { useFormContext } from 'react-hook-form';

import { Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';

import Time from '@northern.tech/common-ui/Time';
import dayjs from 'dayjs';
import pluralize from 'pluralize';

import type { DeploymentFormValues } from '../types';
import { deploymentFormSections, getPhaseStartTime } from '../utils';
import { BatchSizeInput, DelayInput } from './Input';
import type { DelayUnit, RolloutMode } from './constants';
import { delayDefaults, phaseDefaults, phaseLimits, rolloutModes } from './constants';
import type { ActivePhaseComponentProps, PhaseDefinition } from './utils';
import { getPhaseDeviceCount, getPhaseMessages, percentageToDevices } from './utils';

const uniformTableHeaders = ['Batch size', 'First phase begins', 'Delay before next phase'];

export const getUniformBatchDefault = (rolloutMode: RolloutMode, numberDevices: number): number =>
  rolloutMode === rolloutModes.percentage.key
    ? phaseDefaults.batchSize
    : Math.min(numberDevices || phaseLimits.fallbackDeviceCount, phaseLimits.maxDefaultBatchDevices);

const PhasesSummary = ({ deviceCount, delay, delayUnit, filter, isPercentageMode, batchSize }) => {
  let phasesCount = Math.ceil(deviceCount / batchSize);
  let perPhaseCount = batchSize;
  let remainder = deviceCount % batchSize;
  if (isPercentageMode) {
    phasesCount = Math.ceil(phaseLimits.fullBatchPercentage / batchSize);
    perPhaseCount = percentageToDevices(batchSize, deviceCount);
    remainder = percentageToDevices(phaseLimits.fullBatchPercentage % batchSize, deviceCount);
  }
  const delayDescriptor = `${delay}-${pluralize(delayUnit, 1)} delay between phases`;
  const totalDescriptor = `(${deviceCount} ${pluralize('device', deviceCount)} total)`;
  return (
    <div className="margin-top-small margin-left-small">
      <Typography variant="subtitle2">Summary</Typography>
      <Typography variant="body2" color="text.secondary">
        {filter
          ? `Deploy in phases of ${perPhaseCount.toLocaleString()} ${pluralize('device', perPhaseCount)}, with a ${delayDescriptor}`
          : remainder
            ? `${phasesCount - 1} ${pluralize('phase', phasesCount - 1)} with ${perPhaseCount.toLocaleString()} ${pluralize('device', perPhaseCount)}${phasesCount - 1 > 1 ? ' each' : ''} and a ${delayDescriptor}, plus a final phase with ${remainder} ${pluralize('device', remainder)} ${totalDescriptor}`
            : `${phasesCount} ${pluralize('phase', phasesCount)} with ${perPhaseCount.toLocaleString()} ${pluralize('device', perPhaseCount)} ${totalDescriptor}`}
      </Typography>
    </div>
  );
};

export const UniformPhaseSettings = ({ classes = {}, filter, deploymentDeviceCount }: ActivePhaseComponentProps) => {
  const { setValue, watch } = useFormContext<DeploymentFormValues>();
  const rolloutMode: RolloutMode = watch(deploymentFormSections.rolloutMode) || rolloutModes.percentage.key;
  const phases: Array<PhaseDefinition> = watch(deploymentFormSections.phases) || [];
  const configuredStartTime = watch(deploymentFormSections.startTime);
  const maxDevices = watch(deploymentFormSections.maxDevices);
  const isPercentageMode = rolloutMode === rolloutModes.percentage.key;

  const consideredDevices = maxDevices ? maxDevices : deploymentDeviceCount;
  const [definition = {}] = phases;
  const batchSize = definition.batchSize ?? getUniformBatchDefault(rolloutMode, consideredDevices);
  const delay = definition.delay ?? delayDefaults.delay;
  const delayUnit = (definition.delayUnit ?? delayDefaults.delayUnit) as DelayUnit;

  const deviceCount = isPercentageMode ? getPhaseDeviceCount(consideredDevices, batchSize, 0, false) : batchSize || 0;

  const updatePhaseDefinition = (changes: PhaseDefinition) => setValue(deploymentFormSections.phases, [{ batchSize, delay, delayUnit, ...changes }]);

  const handleBatchChange = (value: number) =>
    updatePhaseDefinition({ batchSize: isPercentageMode ? Math.min(phaseLimits.maxPerBatchPercentage, Math.max(1, value)) : Math.max(1, value) });

  const handleDelayChange = (value: number) => updatePhaseDefinition({ delay: Math.max(1, value) });

  const handleDelayUnitChange = ({ target: { value } }) => updatePhaseDefinition({ delayUnit: value });

  const messages = getPhaseMessages({
    batchSize,
    deploymentDeviceCount: consideredDevices,
    isDynamic: !!filter,
    maxDevices,
    remainder: Math.max(0, (isPercentageMode ? phaseLimits.fullBatchPercentage : consideredDevices) - batchSize),
    rolloutMode
  });
  const hasError = messages.some(({ severity }) => severity === 'error');
  const hasWarning = messages.some(({ severity }) => severity === 'warning');

  return (
    <div className="margin-bottom-small margin-top-small">
      <Table size="small">
        <TableHead>
          <TableRow>
            {uniformTableHeaders.map((content, index) => (
              <TableCell key={index}>{content}</TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          <TableRow className={hasError ? classes.rowError : hasWarning ? classes.rowWarning : ''}>
            <TableCell>
              <BatchSizeInput
                value={batchSize}
                onChange={value => handleBatchChange(value ?? 1)}
                isPercentageMode={isPercentageMode}
                deviceCount={deviceCount}
                max={isPercentageMode ? phaseLimits.maxPerBatchPercentage : maxDevices || deploymentDeviceCount}
                hasError={hasError}
                messages={messages}
              />
            </TableCell>
            <TableCell>
              <Time value={getPhaseStartTime([{ delay, delayUnit }], 0, configuredStartTime ?? dayjs())} />
            </TableCell>
            <TableCell>
              <DelayInput
                id="uniform-delay"
                delay={delay}
                delayUnit={delayUnit}
                onDelayChange={value => handleDelayChange(value ?? 1)}
                onDelayUnitChange={handleDelayUnitChange}
              />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
      <PhasesSummary
        batchSize={batchSize}
        filter={filter}
        isPercentageMode={isPercentageMode}
        deviceCount={consideredDevices}
        delay={delay}
        delayUnit={delayUnit}
      />
    </div>
  );
};
