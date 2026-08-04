// Copyright 2019 Northern.tech AS
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
import { useCallback } from 'react';
import { useFormContext } from 'react-hook-form';

import { Alert, Collapse, FormControl, FormControlLabel, ListSubheader, MenuItem, Radio, RadioGroup, Select, Tooltip, Typography, alpha } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import { FormCheckbox } from '@northern.tech/common-ui/forms/FormCheckbox';
import { BENEFITS } from '@northern.tech/store/constants';
import { isDarkMode } from '@northern.tech/store/utils';

import { CustomPhaseTable, convertPhasesToMode } from './phases/CustomPhases';
import type { RolloutMode, RolloutPattern } from './phases/constants';
import { delayDefaults, delayUnits, phaseDefaults, phaseLimits, rolloutModes, rolloutPatterns as rolloutPatternDefinitions } from './phases/constants';
import { getPhasesMessage, toPhaseDescription } from './phases/utils';
import type { DeploymentFormValues } from './types';
import { deploymentFormSections, useDerivedData } from './utils';

const useStyles = makeStyles()(theme => ({
  container: {
    background: isDarkMode(theme.palette.mode) ? alpha(theme.palette.info.light, theme.palette.action.selectedOpacity) : theme.palette.action.hover
  },
  patternSelection: { marginTop: theme.spacing(2), width: 400 }
}));

const rolloutPatterns = {
  [rolloutPatternDefinitions.custom.key]: { ...rolloutPatternDefinitions.custom, component: CustomPhaseTable }
};

const getDefaultPhasesForPattern = (
  rolloutMode: RolloutMode,
  patternValue: string | Array<Record<string, unknown>>,
  numberDevices: number,
  deploymentDeviceCount: number,
  phaseStart: Record<string, unknown>
) => {
  if (rolloutMode === rolloutModes.device_count.key) {
    const defaultBatch = numberDevices > 0 ? Math.max(1, Math.min(numberDevices, phaseDefaults.batchSize)) : phaseDefaults.batchSize;
    if (patternValue === rolloutPatterns.custom.key)
      return [{ batch_size_devices: defaultBatch, delay: delayDefaults.delay, delayUnit: delayUnits.hours, ...phaseStart }, {}];
    return null;
  }
  const minBatch =
    deploymentDeviceCount < phaseDefaults.batchSize ? Math.ceil((1 / deploymentDeviceCount) * phaseLimits.fullBatchPercentage) : phaseDefaults.batchSize;
  if (patternValue === rolloutPatterns.custom.key)
    return [{ batch_size: minBatch, delay: delayDefaults.delay, delayUnit: delayUnits.hours, ...phaseStart }, {}];
  return null;
};

interface RolloutPatternSelectionProps {
  isEnterprise: boolean;
  previousPhases?: Array<Array<Record<string, unknown>>>;
}

export const RolloutPatternSelection = ({ isEnterprise, previousPhases = [] }: RolloutPatternSelectionProps) => {
  const { watch, setValue, getValues } = useFormContext<DeploymentFormValues>();
  const { deploymentDeviceCount, deploymentDeviceIds, filter } = useDerivedData(watch);
  const phases = watch(deploymentFormSections.phases) || [];
  const rolloutMode: RolloutMode = watch(deploymentFormSections.rolloutMode) || rolloutModes.percentage.key;
  const usesPattern = watch(deploymentFormSections.usesPattern);
  const configuredStartTime = watch(deploymentFormSections.startTime);
  const maxDevices = watch(deploymentFormSections.maxDevices);
  const { classes } = useStyles();

  const numberDevices = deploymentDeviceCount ? deploymentDeviceCount : deploymentDeviceIds ? deploymentDeviceIds.length : 0;

  const activePattern = rolloutPatterns.custom.key as RolloutPattern;

  const handlePatternChange = ({ target: { value } }) => {
    const startTime = configuredStartTime ?? (phases.length ? phases[0].start_ts : undefined);
    const phaseStart = { start_ts: startTime };

    const defaultPhases = getDefaultPhasesForPattern(rolloutMode, value, numberDevices, deploymentDeviceCount, filter, phaseStart);
    let nextPhases = [{ batch_size: phaseLimits.fullBatchPercentage }];
    if (Array.isArray(value)) {
      if (value[0].batch_size) {
        nextPhases = rolloutMode === rolloutModes.device_count.key ? convertPhasesToMode(value, rolloutMode, deploymentDeviceCount) : value;
      } else {
        nextPhases = rolloutMode === rolloutModes.percentage.key ? convertPhasesToMode(value, rolloutMode, deploymentDeviceCount) : value;
      }
    }
    setValue(deploymentFormSections.phases, defaultPhases ?? nextPhases);
  };

  const onUsesPatternClick = useCallback(
    ({ target: { checked } }: React.MouseEvent<HTMLButtonElement> & { target: HTMLInputElement }) => {
      const currentPhases = getValues(deploymentFormSections.phases) || [];
      if (!checked) {
        const singlePhase = currentPhases.length > 0 ? currentPhases.slice(0, 1) : [{ batch_size: phaseLimits.fullBatchPercentage }];
        setValue(deploymentFormSections.phases, singlePhase);
      } else if (currentPhases.length < 2) {
        const startTime = configuredStartTime ?? (currentPhases.length ? currentPhases[0].start_ts : undefined);
        const defaultPhases = getDefaultPhasesForPattern(rolloutMode, rolloutPatterns.custom.key, numberDevices, deploymentDeviceCount, filter, {
          start_ts: startTime
        });
        if (defaultPhases) {
          setValue(deploymentFormSections.phases, defaultPhases);
        }
      }
    },
    [getValues, setValue, configuredStartTime, rolloutMode, numberDevices, deploymentDeviceCount, filter]
  );

  const handleModeChange = ({ target: { value } }: React.ChangeEvent<HTMLInputElement>) => setValue(deploymentFormSections.rolloutMode, value);

  const previousPhaseOptions =
    previousPhases.length > 0
      ? previousPhases.map((previousPhaseSetting, index) => {
          const { phasesDescription, tooltip } = toPhaseDescription(previousPhaseSetting, numberDevices);
          return (
            <MenuItem key={`previousPhaseSetting-${index}`} value={previousPhaseSetting}>
              <Tooltip title={tooltip} placement="left">
                <div className="full-width">{phasesDescription}</div>
              </Tooltip>
            </MenuItem>
          );
        })
      : [
          <MenuItem key="noPreviousPhaseSetting" disabled={true} style={{ opacity: '0.4' }}>
            No recent patterns
          </MenuItem>
        ];

  const phasesNotification = getPhasesMessage({ filter, rolloutPattern: activePattern, maxDevices });

  const { component: ActivePatternComponent } = rolloutPatterns[activePattern];
  return (
    <>
      <FormCheckbox
        id={deploymentFormSections.usesPattern}
        disabled={!isEnterprise || numberDevices === 0}
        handleClick={onUsesPatternClick}
        label={
          <div className="flexbox align-items-center">
            Select a rollout pattern
            <InfoHintContainer>
              <EnterpriseNotification id={BENEFITS.phasedDeployments.id} />
              <DocsTextLink id={DOCSTIPS.phasedDeployments.id} />
            </InfoHintContainer>
          </div>
        }
        slotProps={{ checkbox: { className: 'margin-left-small', size: 'small' } }}
      />
      <Collapse className={usesPattern ? 'margin-bottom-small' : ''} in={usesPattern}>
        {numberDevices > 1 && (
          <FormControl className={classes.patternSelection}>
            <Select onChange={handlePatternChange} value={activePattern} disabled={!isEnterprise}>
              {[
                ...Object.values(rolloutPatterns).map(({ key, tip, title }) => (
                  <MenuItem key={key} divider value={key}>
                    <Tooltip title={tip} placement="left">
                      <div className="full-width">{title}</div>
                    </Tooltip>
                  </MenuItem>
                )),
                <ListSubheader key="phaseSettingsSubheader">Recent patterns</ListSubheader>,
                ...previousPhaseOptions
              ]}
            </Select>
          </FormControl>
        )}
        <div className={`margin-top-x-small padding-left-small padding-right-small padding-top-x-small padding-bottom-x-small ${classes.container}`}>
          <RadioGroup className="flexbox align-items-center margin-bottom-small margin-top-x-small" row value={rolloutMode} onChange={handleModeChange}>
            <Typography className="margin-right-x-small">Rollout phases:</Typography>
            {Object.values(rolloutModes).map(({ key, title }) => (
              <FormControlLabel key={key} value={key} control={<Radio size="small" />} label={title} />
            ))}
          </RadioGroup>
          <ActivePatternComponent deploymentDeviceCount={numberDevices} filter={filter} />
          {phasesNotification && (
            <Alert className="margin-top-x-small margin-bottom-x-small" severity={phasesNotification.severity}>
              {phasesNotification.message}
            </Alert>
          )}
        </div>
      </Collapse>
    </>
  );
};
