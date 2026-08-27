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
import { useFieldArray, useFormContext } from 'react-hook-form';

import { Add as AddIcon, Close as CancelIcon, RepeatOutlined as RepeatIcon } from '@mui/icons-material';
import { Button, FormHelperText, IconButton, Table, TableBody, TableCell, TableHead, TableRow, Typography } from '@mui/material';

import Time from '@northern.tech/common-ui/Time';

import type { DeploymentFormValues } from '../types';
import { deploymentFormSections, getPhaseStartTime, useValidatedSetValue } from '../utils';
import { BatchSizeInput, DelayInput } from './Input';
import type { DelayUnit, RolloutMode } from './constants';
import { delayDefaults, phaseDefaults, rolloutModes } from './constants';
import type { ActivePhaseComponentProps, PhaseDefinition } from './utils';
import { computePhaseInfo, getDefinitionsRemainder } from './utils';

const evenSplitThreshold = 50;

const tableHeaders = ['Phases', 'Batch size', 'Phase begins', 'Delay before next phase', ''];

export const CustomPhaseTable = ({ classes = {}, filter, deploymentDeviceCount }: ActivePhaseComponentProps) => {
  const {
    control,
    formState: { errors },
    watch
  } = useFormContext<DeploymentFormValues>();
  const setValue = useValidatedSetValue();
  const { fields, append, insert, remove, replace } = useFieldArray({ control, name: deploymentFormSections.phases });

  const phases: Array<PhaseDefinition> = watch(deploymentFormSections.phases) || [];
  const rolloutMode: RolloutMode = watch(deploymentFormSections.rolloutMode) || rolloutModes.percentage.key;
  const maxDevices = watch(deploymentFormSections.maxDevices);
  const isPercentageMode = rolloutMode === rolloutModes.percentage.key;

  const startTime = watch(deploymentFormSections.startTime) ?? new Date();

  const remainder = getDefinitionsRemainder({ phases, numberDevices: deploymentDeviceCount, rolloutMode });

  const updateBatchSize = (value: number, index: number) => setValue(`${deploymentFormSections.phases}.${index}.batchSize`, Math.max(0, Math.round(value)));

  const updateDelay = (value: number, index: number) => setValue(`${deploymentFormSections.phases}.${index}.delay`, Math.max(1, value));

  const updateDelayUnit = (value: string, index: number) => setValue(`${deploymentFormSections.phases}.${index}.delayUnit`, value);

  const addPhase = () => {
    // make it default 10, unless remainder is <=10 in which case make it half remainder
    const defaultBatch =
      deploymentDeviceCount > 0
        ? Math.max(1, remainder > phaseDefaults.batchSize ? phaseDefaults.batchSize : Math.floor(remainder / 2))
        : phaseDefaults.batchSize;
    append({ batchSize: defaultBatch, ...delayDefaults });
  };

  const repeatPhase = (index: number) => {
    const source = phases[index];
    if (isPercentageMode && (source.batchSize ?? 0) >= evenSplitThreshold) {
      // distribute in 2 even phases instead of the regular handling - the duplicate becomes the derived closing phase
      replace([{ ...source, batchSize: evenSplitThreshold }]);
      return;
    }
    insert(index + 1, { batchSize: source.batchSize, delay: source.delay || delayDefaults.delay, delayUnit: source.delayUnit || delayDefaults.delayUnit });
  };

  // the final phase always takes whatever the sized phase definitions leave over
  const rows: Array<PhaseDefinition | undefined> = [...phases, undefined];
  const schedule = [...phases, {}];

  const mappedPhases = rows.map((phase, index) => {
    const { batchValue, deviceCount, hasError, hasWarning, max, messages } = computePhaseInfo({
      index,
      isDynamic: !!filter,
      phases,
      numberDevices: deploymentDeviceCount,
      rolloutMode,
      maxDevices
    });
    const isFinal = phase === undefined;

    return (
      <TableRow key={isFinal ? 'final' : fields[index]?.id || index} className={hasError ? classes.rowError : hasWarning ? classes.rowWarning : ''}>
        <TableCell className="nowrap">
          <Typography variant="body2">{`Phase ${index + 1}`}</Typography>
          {isFinal && rows.length > 1 && <Typography variant="caption">(Final step)</Typography>}
        </TableCell>
        <TableCell>
          <BatchSizeInput
            deviceCount={deviceCount}
            value={batchValue}
            onChange={value => updateBatchSize(value ?? 1, index)}
            isPercentageMode={isPercentageMode}
            hasError={hasError}
            max={max}
            disabled={isFinal}
            messages={messages}
          />
        </TableCell>
        <TableCell>
          <Time value={getPhaseStartTime(schedule, index, startTime)} />
        </TableCell>
        <TableCell>
          {!isFinal && phase.delay ? (
            <DelayInput
              id={`phase-delay-${index}`}
              delay={phase.delay}
              delayUnit={phase.delayUnit as DelayUnit}
              onDelayChange={value => updateDelay(value ?? 1, index)}
              onDelayUnitChange={({ target: { value } }) => updateDelayUnit(value as string, index)}
            />
          ) : (
            '-'
          )}
        </TableCell>
        <TableCell>
          {!isFinal ? (
            <div className="flexbox">
              <IconButton disabled={!remainder} onClick={() => repeatPhase(index)} title="Repeat phase">
                <RepeatIcon />
              </IconButton>
              <IconButton disabled={phases.length <= 1} onClick={() => remove(index)} title="Remove phase">
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
      <Button
        className="margin-bottom-x-small margin-top-small"
        color="info"
        disabled={!remainder}
        variant="outlined"
        startIcon={<AddIcon />}
        onClick={addPhase}
      >
        Add a phase
      </Button>
      {!!errors.phases && <FormHelperText error>{errors.phases.message}</FormHelperText>}
    </>
  );
};
