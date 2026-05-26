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

import { Add as AddIcon, Close as CancelIcon, RepeatOutlined as RepeatIcon } from '@mui/icons-material';
import { Button, IconButton, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';

import Time from '@northern.tech/common-ui/Time';
import type { Filter } from '@northern.tech/types/MenderTypes';

import type { DeploymentFormValues } from '../types';
import { deploymentFormSections, getPhaseStartTime } from '../utils';
import { BatchSizeInput, DelayInput } from './Input';
import type { RolloutMode } from './constants';
import { delayDefaults, delayUnits, phaseDefaults, rolloutModes } from './constants';
import type { UiDeploymentPhase } from './utils';
import { computePhaseInfo, devicesToPercentage, getRemainder, percentageToDevices, useRowStyles } from './utils';

const convertPhasesToMode = (phases, newMode: RolloutMode, numberDevices: number) =>
  phases.map((phase, index, source) => {
    const isLast = index === source.length - 1;
    if (newMode === rolloutModes.device_count.key) {
      const { batch_size, ...rest } = phase;
      if (!batch_size && isLast) return rest;
      return { ...rest, batch_size_devices: percentageToDevices(batch_size || 0, numberDevices) };
    }
    const { batch_size_devices, ...rest } = phase;
    if (!batch_size_devices && isLast) return rest;
    return { ...rest, batch_size: devicesToPercentage(batch_size_devices || 0, numberDevices) };
  });

const applyBatchSizeUpdate = (phases, value: number, index: number, isPercentageMode: boolean) => {
  const newPhases = [...phases];
  const rounded = Math.max(0, Math.round(value));
  if (isPercentageMode) {
    newPhases[index] = { ...newPhases[index], batch_size: rounded };
  } else {
    newPhases[index] = { ...newPhases[index], batch_size_devices: rounded };
  }
  return newPhases;
};

const tableHeaders = ['Phases', 'Batch size', 'Phase begins', 'Delay before next phase', ''];

export const CustomPhaseTable = ({ filter, deploymentDeviceCount }: { deploymentDeviceCount: number; filter?: Filter }) => {
  const { watch, setValue, getValues } = useFormContext<DeploymentFormValues>();

  const phases: Array<UiDeploymentPhase> = watch(deploymentFormSections.phases) || [];
  const rolloutMode: RolloutMode = watch(deploymentFormSections.rolloutMode) || rolloutModes.percentage.key;
  const maxDevices = watch(deploymentFormSections.maxDevices);
  const isPercentageMode = rolloutMode === rolloutModes.percentage.key;
  const batchKey = isPercentageMode ? rolloutModes.percentage.batchKey : rolloutModes.device_count.batchKey;

  const configuredStartTime = watch(deploymentFormSections.startTime);
  const startTime = configuredStartTime ?? (phases.length ? phases[0].start_ts || new Date() : new Date());

  const { classes } = useRowStyles();

  const prevModeRef = useRef(rolloutMode);
  useEffect(() => {
    if (prevModeRef.current === rolloutMode) {
      return;
    }
    prevModeRef.current = rolloutMode;
    const currentPhases = getValues(deploymentFormSections.phases);
    setValue(deploymentFormSections.phases, convertPhasesToMode(currentPhases, rolloutMode, deploymentDeviceCount));
  }, [rolloutMode, deploymentDeviceCount, setValue, getValues]);

  const updateDelay = (value, index) => {
    const newPhases = [...phases];
    newPhases[index] = { ...newPhases[index], delay: Math.max(1, value) };
    setValue(deploymentFormSections.phases, newPhases);
  };

  const updateBatchSize = (value, index) => setValue(deploymentFormSections.phases, applyBatchSizeUpdate(phases, value, index, isPercentageMode));

  const addPhase = () => {
    const newPhases = [...phases];
    const remainder = getRemainder({ phases: newPhases, numberDevices: deploymentDeviceCount, rolloutMode });
    if (isPercentageMode) {
      newPhases[newPhases.length - 1] = {
        ...newPhases[newPhases.length - 1],
        batch_size: remainder > phaseDefaults.batchSize ? phaseDefaults.batchSize : Math.floor(remainder / 2),
        delay: newPhases[newPhases.length - 1].delay || 2,
        delayUnit: newPhases[newPhases.length - 1].delayUnit || delayUnits.hours
      };
    } else {
      const defaultBatch =
        deploymentDeviceCount > 0
          ? Math.max(1, remainder > phaseDefaults.batchSize ? phaseDefaults.batchSize : Math.floor(remainder / 2))
          : phaseDefaults.batchSize;
      newPhases[newPhases.length - 1] = {
        ...newPhases[newPhases.length - 1],
        batch_size_devices: defaultBatch,
        delay: newPhases[newPhases.length - 1].delay || 2,
        delayUnit: newPhases[newPhases.length - 1].delayUnit || delayUnits.hours
      };
    }
    newPhases.push({});
    setValue(deploymentFormSections.phases, newPhases);
  };

  const removePhase = index => {
    const newPhases = [...phases];
    newPhases.splice(index, 1);
    const { [batchKey]: _removed, delay, ...newPhase } = newPhases[newPhases.length - 1];
    if (newPhases.length > 1) {
      newPhase.delay = delay;
    }
    newPhases[newPhases.length - 1] = newPhase;
    setValue(deploymentFormSections.phases, newPhases);
  };

  const repeatPhase = (index: number) => {
    const newPhases = [...phases];
    const source = newPhases[index];
    const duplicate = { [batchKey]: source[batchKey], delay: source.delay || delayDefaults.delay, delayUnit: source.delayUnit || delayDefaults.delayUnit };
    if (isPercentageMode && source[batchKey] >= 50) {
      // distribute in 2 even phases instead of the regular handling
      setValue(deploymentFormSections.phases, [
        { ...source, [batchKey]: 50 },
        { ...duplicate, [batchKey]: 50 }
      ]);
      return;
    }
    newPhases.splice(index + 1, 0, duplicate);
    setValue(deploymentFormSections.phases, newPhases);
  };

  const handleDelayToggle = (value, index) => {
    const newPhases = [...phases];
    newPhases[index] = { ...newPhases[index], delayUnit: value };
    setValue(deploymentFormSections.phases, newPhases);
  };

  const mappedPhases = phases.map((phase, index) => {
    const { batchValue, deviceCount, hasError, hasWarning, max, messages } = computePhaseInfo({
      index,
      isDynamic: !!filter,
      phases,
      numberDevices: deploymentDeviceCount,
      rolloutMode,
      maxDevices
    });
    const isLast = index === phases.length - 1;

    return (
      <TableRow key={index} className={hasError ? classes.rowError : hasWarning ? classes.rowWarning : ''}>
        <TableCell className="nowrap">
          <Typography variant="body2">{`Phase ${index + 1}`}</Typography>
          {isLast && phases.length > 1 && <Typography variant="caption">(Final step)</Typography>}
        </TableCell>
        <TableCell>
          <BatchSizeInput
            deviceCount={deviceCount}
            value={batchValue}
            onChange={value => updateBatchSize(value ?? 1, index)}
            isPercentageMode={isPercentageMode}
            hasError={hasError}
            max={max}
            disabled={isLast && deviceCount >= 1}
            messages={messages}
          />
        </TableCell>
        <TableCell>
          <Time value={getPhaseStartTime(phases, index, startTime)} />
        </TableCell>
        <TableCell>
          {phase.delay && !isLast ? (
            <DelayInput
              id={`phase-delay-${index}`}
              delay={phase.delay}
              delayUnit={phase.delayUnit}
              onDelayChange={value => updateDelay(value ?? 1, index)}
              onDelayUnitChange={({ target: { value } }) => handleDelayToggle(value, index)}
            />
          ) : (
            '-'
          )}
        </TableCell>
        <TableCell>
          {!isLast && phases.length > 1 ? (
            <div className="flexbox">
              <IconButton onClick={() => repeatPhase(index)} title="Repeat phase">
                <RepeatIcon />
              </IconButton>
              <IconButton onClick={() => removePhase(index)} title="Remove phase">
                <CancelIcon />
              </IconButton>
            </div>
          ) : null}
        </TableCell>
      </TableRow>
    );
  });

  return (
    <>
      <Table size="small">
        <TableHead>
          <TableRow>
            {tableHeaders.map((content, index) => (
              <TableCell key={index}>{content}</TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>{mappedPhases}</TableBody>
      </Table>
      <Button className="margin-bottom-x-small margin-top-small" color="info" variant="outlined" startIcon={<AddIcon />} onClick={addPhase}>
        Add a phase
      </Button>
    </>
  );
};
