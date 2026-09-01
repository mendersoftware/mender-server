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
import type { Filter } from '@northern.tech/types/MenderTypes';
import type { StandardizedPhase } from '@northern.tech/utils/helpers';

import { CustomPhaseTable } from './phases/CustomPhases';
import { getUniformBatchDefault } from './phases/UniformPhases';
import type { RolloutMode, RolloutPattern } from './phases/constants';
import { delayDefaults, phaseDefaults, phaseLimits, rolloutModes, rolloutPatterns as rolloutPatternDefinitions } from './phases/constants';
import type { PhaseDefinition } from './phases/utils';
import { convertDefinitionsToMode, getPhasesMessage, parsePreviousPhases, toPhaseDescription } from './phases/utils';
import type { DeploymentFormValues } from './types';
import { DisabledReasonHint, deploymentFormSections, useValidatedSetValue } from './utils';

const useStyles = makeStyles()(theme => ({
  container: {
    background: isDarkMode(theme.palette.mode) ? alpha(theme.palette.info.light, theme.palette.action.selectedOpacity) : theme.palette.action.hover
  },
  patternSelection: { marginTop: theme.spacing(2), width: 400 },
  rowError: { backgroundColor: alpha(theme.palette.error.main, theme.palette.action.selectedOpacity) },
  rowWarning: { backgroundColor: alpha(theme.palette.warning.main, theme.palette.action.selectedOpacity) }
}));

const rolloutPatterns = {
  [rolloutPatternDefinitions.custom.key]: { ...rolloutPatternDefinitions.custom, component: CustomPhaseTable }
};

const getDefaultPhaseDefinitions = (pattern: RolloutPattern, rolloutMode: RolloutMode, numberDevices: number): PhaseDefinition[] => {
  if (pattern === rolloutPatternDefinitions.uniform.key) {
    return [{ batchSize: getUniformBatchDefault(rolloutMode, numberDevices), ...delayDefaults }];
  }
  if (rolloutMode === rolloutModes.device_count.key) {
    return [{ batchSize: numberDevices > 0 ? Math.max(1, Math.min(numberDevices, phaseDefaults.batchSize)) : phaseDefaults.batchSize, ...delayDefaults }];
  }
  const minBatch =
    numberDevices > 0 && numberDevices < phaseDefaults.batchSize ? Math.ceil((1 / numberDevices) * phaseLimits.fullBatchPercentage) : phaseDefaults.batchSize;
  return [{ batchSize: minBatch, ...delayDefaults }];
};

interface RolloutPatternSelectionProps {
  deploymentDeviceCount?: number;
  disabledReason: string;
  filter?: Filter;
  isEnterprise: boolean;
  previousPhases?: Array<Array<StandardizedPhase>>;
}

export const RolloutPatternSelection = ({
  deploymentDeviceCount = 0,
  isEnterprise,
  disabledReason = '',
  filter,
  previousPhases = []
}: RolloutPatternSelectionProps) => {
  const { watch, getValues } = useFormContext<DeploymentFormValues>();
  const setValue = useValidatedSetValue();
  const rolloutMode: RolloutMode = watch(deploymentFormSections.rolloutMode) || rolloutModes.percentage.key;
  const usesPattern = watch(deploymentFormSections.usesPattern);
  const maxDevices = watch(deploymentFormSections.maxDevices);
  const { classes } = useStyles();

  const activePattern: RolloutPattern = watch(deploymentFormSections.rolloutPattern) || (rolloutPatternDefinitions.custom.key as RolloutPattern);

  const handlePatternChange = ({ target: { value } }) => {
    if (typeof value === 'string') {
      setValue(deploymentFormSections.rolloutPattern, value);
      setValue(deploymentFormSections.phases, getDefaultPhaseDefinitions(value as RolloutPattern, rolloutMode, deploymentDeviceCount));
      return;
    }
    // a recent pattern was picked - adopt its shape, converted to the currently selected mode where needed
    const { phases, pattern, rolloutMode: storedMode } = parsePreviousPhases(value);
    setValue(deploymentFormSections.rolloutPattern, rolloutPatterns[pattern] ? pattern : rolloutPatternDefinitions.custom.key);
    setValue(deploymentFormSections.phases, storedMode === rolloutMode ? phases : convertDefinitionsToMode(phases, rolloutMode, deploymentDeviceCount));
  };

  const onUsesPatternClick = useCallback(
    ({ target: { checked } }: React.MouseEvent<HTMLButtonElement> & { target: HTMLInputElement }) => {
      if (!checked) {
        setValue(deploymentFormSections.phases, []);
        return;
      }
      if (!getValues(deploymentFormSections.phases)?.length) {
        const pattern = getValues(deploymentFormSections.rolloutPattern) || (rolloutPatternDefinitions.custom.key as RolloutPattern);
        setValue(deploymentFormSections.phases, getDefaultPhaseDefinitions(pattern, rolloutMode, deploymentDeviceCount));
      }
    },
    [getValues, setValue, rolloutMode, deploymentDeviceCount]
  );

  // switching the mode only changes what the batch numbers mean, so they get converted in the same go - which keeps
  // the mode field the single signal for how to read them
  const handleModeChange = ({ target: { value } }: React.ChangeEvent<HTMLInputElement>) => {
    const newMode = value as RolloutMode;
    const conversionBasis = activePattern === rolloutPatternDefinitions.uniform.key && maxDevices ? maxDevices : deploymentDeviceCount;
    setValue(deploymentFormSections.rolloutMode, newMode);
    setValue(deploymentFormSections.phases, convertDefinitionsToMode(getValues(deploymentFormSections.phases) || [], newMode, conversionBasis));
  };

  const previousPhaseOptions =
    previousPhases.length > 0
      ? previousPhases.map((previousPhaseSetting, index) => {
          const { phasesDescription, tooltip } = toPhaseDescription(previousPhaseSetting, deploymentDeviceCount);
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

  const { component: ActivePatternComponent } = rolloutPatterns[activePattern] ?? rolloutPatterns[rolloutPatternDefinitions.custom.key];
  return (
    <>
      <FormCheckbox
        id={deploymentFormSections.usesPattern}
        disabled={!isEnterprise || deploymentDeviceCount === 0 || !!disabledReason}
        handleClick={onUsesPatternClick}
        label={
          <div className="flexbox align-items-center">
            Select a rollout pattern
            <InfoHintContainer>
              <EnterpriseNotification id={BENEFITS.phasedDeployments.id} />
              {isEnterprise && <DisabledReasonHint reason={disabledReason} />}
              <DocsTextLink id={DOCSTIPS.phasedDeployments.id} />
            </InfoHintContainer>
          </div>
        }
        slotProps={{ checkbox: { className: 'margin-left-small', size: 'small' } }}
      />
      <Collapse className={usesPattern ? 'margin-bottom-small' : ''} in={usesPattern}>
        {deploymentDeviceCount > 1 && (
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
          <ActivePatternComponent classes={classes} deploymentDeviceCount={deploymentDeviceCount} filter={filter} />
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
