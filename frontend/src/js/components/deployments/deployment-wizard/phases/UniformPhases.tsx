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
import { useEffect, useRef } from 'react';
import { useFormContext } from 'react-hook-form';

import { Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';

import Time from '@northern.tech/common-ui/Time';
import type { Filter } from '@northern.tech/types/MenderTypes';
import dayjs from 'dayjs';
import durationPlugin from 'dayjs/plugin/duration';
import type { DurationUnitType } from 'dayjs/plugin/duration';
import pluralize from 'pluralize';

import type { DeploymentFormValues } from '../types';
import { deploymentFormSections, getPhaseStartTime } from '../utils';
import { BatchSizeInput, DelayInput } from './Input';
import type { RolloutMode } from './constants';
import { delayUnits, phaseDefaults, rolloutModes } from './constants';
import { computePhaseInfo, devicesToPercentage, getPhaseDeviceCount, percentageToDevices, useRowStyles } from './utils';

dayjs.extend(durationPlugin);

const delayToSeconds = (delay: number, unit: string): string => `${dayjs.duration(delay, unit as dayjs.ManipulateType).asSeconds()}s`;

const parseInterval = (interval?: string): { delay: number; delayUnit: string } => {
  if (!interval) return { delay: 2, delayUnit: delayUnits.hours };
  const seconds = dayjs.duration(parseInt(interval) || phaseDefaults.delay, 'seconds');
  return [delayUnits.days, delayUnits.hours, delayUnits.minutes].reduce(
    (accu, unit) => {
      const durationPerUnit = seconds.get(unit as DurationUnitType);
      if (durationPerUnit >= 1 && Number.isInteger(durationPerUnit)) {
        return { delay: durationPerUnit, delayUnit: unit };
      }
      return accu;
    },
    { delay: Math.max(1, Math.round(seconds.asHours())), delayUnit: delayUnits.hours }
  );
};

const uniformTableHeaders = ['Batch size', 'First phase begins', 'Delay before next phase'];

const PhasesSummary = ({ deviceCount, delay, delayUnit, filter, isPercentageMode, batchSize }) => {
  let phasesCount = Math.ceil(deviceCount / batchSize);
  let perPhaseCount = batchSize;
  let remainder = deviceCount % batchSize;
  if (isPercentageMode) {
    phasesCount = Math.ceil(100 / batchSize);
    perPhaseCount = percentageToDevices(batchSize, deviceCount);
    remainder = percentageToDevices(100 % batchSize, deviceCount);
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

export const UniformPhaseSettings = ({ filter, deploymentDeviceCount }: { deploymentDeviceCount: number; filter?: Filter }) => {
  const { watch, setValue, getValues } = useFormContext<DeploymentFormValues>();
  const rolloutMode: RolloutMode = watch(deploymentFormSections.rolloutMode) || rolloutModes.percentage.key;
  const uniformPhases = watch(deploymentFormSections.uniform_phases);
  const configuredStartTime = watch(deploymentFormSections.startTime);
  const maxDevices = watch(deploymentFormSections.maxDevices);
  const isPercentageMode = rolloutMode === rolloutModes.percentage.key;

  const consideredDevices = maxDevices ? maxDevices : deploymentDeviceCount;
  const batchSize = uniformPhases?.batch_size ?? (isPercentageMode ? phaseDefaults.batchSize : undefined);
  const batchDevices = uniformPhases?.batch_size_devices ?? (isPercentageMode ? undefined : Math.min(consideredDevices || 100, 2000));
  const { delay, delayUnit } = parseInterval(uniformPhases?.time_interval);

  const currentBatch = isPercentageMode ? batchSize : batchDevices;
  const deviceCount = isPercentageMode ? getPhaseDeviceCount(consideredDevices, batchSize, 0, false) : batchDevices || 0;

  const { classes } = useRowStyles();

  const prevModeRef = useRef(rolloutMode);
  useEffect(() => {
    if (prevModeRef.current === rolloutMode) {
      return;
    }
    prevModeRef.current = rolloutMode;
    const current = getValues(deploymentFormSections.uniform_phases);
    const interval = current?.time_interval || `${phaseDefaults.delay}s`;
    if (rolloutMode === rolloutModes.device_count.key) {
      const percentage = current?.batch_size || phaseDefaults.batchSize;
      setValue(deploymentFormSections.uniform_phases, { batch_size_devices: percentageToDevices(percentage, consideredDevices), time_interval: interval });
    } else {
      const devices = current?.batch_size_devices || Math.min(consideredDevices || 100, 2000);
      setValue(deploymentFormSections.uniform_phases, { batch_size: devicesToPercentage(devices, consideredDevices), time_interval: interval });
    }
  }, [rolloutMode, consideredDevices, setValue, getValues]);

  const updateUniformPhases = (newBatch?: number, newBatchDevices?: number, newDelay?: number, newUnit?: string) => {
    const nextDelay = newDelay ?? delay;
    const unit = newUnit ?? delayUnit;
    setValue(deploymentFormSections.uniform_phases, {
      ...(isPercentageMode ? { batch_size: newBatch ?? batchSize } : { batch_size_devices: newBatchDevices ?? batchDevices }),
      time_interval: delayToSeconds(nextDelay, unit)
    });
  };

  const handleBatchChange = (value: number) => {
    if (isPercentageMode) {
      updateUniformPhases(Math.min(99, Math.max(1, value)), undefined);
    } else {
      updateUniformPhases(undefined, Math.max(1, value));
    }
  };

  const handleDelayChange = (value: number) => updateUniformPhases(undefined, undefined, Math.max(1, value));

  const handleDelayUnitChange = ({ target: { value } }) => updateUniformPhases(undefined, undefined, undefined, value);

  const { hasError, hasWarning, max, messages } = computePhaseInfo({
    index: 0,
    phases: [{}],
    isDynamic: !!filter,
    numberDevices: deviceCount,
    rolloutMode,
    maxDevices
  });

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
                value={currentBatch}
                onChange={value => handleBatchChange(value ?? 1)}
                isPercentageMode={isPercentageMode}
                deviceCount={deviceCount}
                max={isPercentageMode ? 99 : maxDevices ? max : deploymentDeviceCount}
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
        batchSize={currentBatch}
        filter={filter}
        isPercentageMode={isPercentageMode}
        deviceCount={consideredDevices}
        delay={delay}
        delayUnit={delayUnit}
      />
    </div>
  );
};
