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
import { useCallback, useState } from 'react';
import { useFormContext } from 'react-hook-form';

import {
  Alert,
  Checkbox,
  Collapse,
  FormControl,
  FormControlLabel,
  ListSubheader,
  MenuItem,
  Radio,
  RadioGroup,
  Select,
  Tooltip,
  Typography
} from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import { ALL_DEVICES, BENEFITS } from '@northern.tech/store/constants';
import { isDarkMode } from '@northern.tech/store/utils';
import type { Filter } from '@northern.tech/types/MenderTypes';

import { CustomPhaseTable } from './phases/CustomPhases';
import { UniformPhaseSettings } from './phases/UniformPhases';
import type { RolloutPattern } from './phases/constants';
import { type RolloutMode, delayUnits, phaseDefaults, rolloutModes, rolloutPatterns as rolloutPatternDefinitions } from './phases/constants';
import { getPhasesMessage, toPhaseDescription } from './phases/utils';
import type { DeploymentFormValues } from './types';
import { deploymentFormSections, useDerivedData } from './utils';

const useStyles = makeStyles()(theme => ({
  container: {
    background: isDarkMode(theme.palette.mode) ? theme.palette.info.dark : theme.palette.info.light
  },
  patternSelection: { marginTop: theme.spacing(2), width: 400 }
}));

const rolloutPatterns = {
  [rolloutPatternDefinitions.custom.key]: { ...rolloutPatternDefinitions.custom, component: CustomPhaseTable },
  [rolloutPatternDefinitions.uniform.key]: { ...rolloutPatternDefinitions.uniform, component: UniformPhaseSettings }
};

const getDefaultPhasesForPattern = (
  rolloutMode: RolloutMode,
  patternValue: string,
  numberDevices: number,
  deploymentDeviceCount: number,
  filter: Filter,
  phaseStart: Record<string, unknown>
) => {
  if (rolloutMode === rolloutModes.device_count.key) {
    const defaultBatch = numberDevices > 0 ? Math.max(1, Math.min(numberDevices, 10)) : 10;
    if (patternValue === rolloutPatterns.custom.key) return [{ batch_size_devices: defaultBatch, delay: 2, delayUnit: delayUnits.hours, ...phaseStart }, {}];
    return null;
  }
  const minBatch = deploymentDeviceCount < 10 && !filter ? Math.ceil((1 / deploymentDeviceCount) * 100) : 10;
  if (patternValue === rolloutPatterns.custom.key) return [{ batch_size: minBatch, delay: 2, delayUnit: delayUnits.hours, ...phaseStart }, {}];
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
  const uniformPhases = watch(deploymentFormSections.uniform_phases);
  const configuredStartTime = watch(deploymentFormSections.startTime);
  const maxDevices = watch(deploymentFormSections.maxDevices);
  const group = watch(deploymentFormSections.group);

  const [usesPattern, setUsesPattern] = useState(phases.some(i => i));
  const { classes } = useStyles();

  const numberDevices = deploymentDeviceCount ? deploymentDeviceCount : deploymentDeviceIds ? deploymentDeviceIds.length : 0;
  const isEmptyGroup = numberDevices === 0 && !filter && group !== ALL_DEVICES;

  const activePattern = (
    uniformPhases?.batch_size || uniformPhases?.batch_size_devices ? rolloutPatterns.uniform.key : rolloutPatterns.custom.key
  ) as RolloutPattern;

  const handlePatternChange = ({ target: { value } }) => {
    const startTime = configuredStartTime ?? (phases.length ? phases[0].start_ts : undefined);
    const phaseStart = { start_ts: startTime };

    if (value === rolloutPatterns.uniform.key) {
      const isPercentageMode = rolloutMode === rolloutModes.percentage.key;
      setValue(deploymentFormSections.phases, []);
      const defaultBatch = isPercentageMode ? 10 : Math.min(numberDevices || 100, 2000);
      setValue(deploymentFormSections.uniform_phases, {
        ...(isPercentageMode ? { batch_size: defaultBatch } : { batch_size_devices: defaultBatch }),
        time_interval: `${phaseDefaults.delay}s`
      });
      return;
    }

    setValue(deploymentFormSections.uniform_phases, undefined);
    const defaultPhases = getDefaultPhasesForPattern(rolloutMode, value, numberDevices, deploymentDeviceCount, filter, phaseStart);
    setValue(deploymentFormSections.phases, defaultPhases ?? (Array.isArray(value) ? structuredClone(value) : [{ batch_size: 100 }]));
  };

  const onUsesPatternClick = useCallback(() => {
    if (usesPattern) {
      const currentPhases = getValues(deploymentFormSections.phases) || [];
      const singlePhase = currentPhases.length > 0 ? currentPhases.slice(0, 1) : [{ batch_size: 100 }];
      setValue(deploymentFormSections.phases, singlePhase);
      setValue(deploymentFormSections.uniform_phases, undefined);
    } else {
      const currentPhases = getValues(deploymentFormSections.phases) || [];
      if (currentPhases.length < 2) {
        const startTime = configuredStartTime ?? (currentPhases.length ? currentPhases[0].start_ts : undefined);
        const defaultPhases = getDefaultPhasesForPattern(rolloutMode, rolloutPatterns.custom.key, numberDevices, deploymentDeviceCount, filter, {
          start_ts: startTime
        });
        if (defaultPhases) {
          setValue(deploymentFormSections.phases, defaultPhases);
        }
      }
    }
    setUsesPattern(!usesPattern);
  }, [usesPattern, getValues, setValue, configuredStartTime, rolloutMode, numberDevices, deploymentDeviceCount, filter]);

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
      <FormControlLabel
        control={<Checkbox color="primary" checked={usesPattern} disabled={!isEnterprise || isEmptyGroup} onChange={onUsesPatternClick} size="small" />}
        label={
          <div className="flexbox align-items-center">
            Select a rollout pattern
            <DocsTextLink className="margin-left-x-small" id={DOCSTIPS.phasedDeployments.id} />
            <InfoHintContainer>
              <EnterpriseNotification id={BENEFITS.phasedDeployments.id} />
            </InfoHintContainer>
          </div>
        }
      />
      <Collapse className="margin-bottom-small" in={usesPattern}>
        <FormControl className={classes.patternSelection}>
          <Select onChange={handlePatternChange} value={activePattern} disabled={!isEnterprise}>
            {(numberDevices > 1 || filter) && [
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
