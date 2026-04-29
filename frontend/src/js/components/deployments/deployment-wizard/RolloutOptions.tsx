// Copyright 2021 Northern.tech AS
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
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';

import { Checkbox, Collapse, FormControl, FormControlLabel, FormGroup } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTooltip } from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import { FormCheckbox } from '@northern.tech/common-ui/forms/FormCheckbox';
import { NumberInput } from '@northern.tech/common-ui/forms/NumberInput';
import { BENEFITS } from '@northern.tech/store/constants';
import { toggle } from '@northern.tech/utils/helpers';

import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';
import RolloutSteps from './RolloutSteps';
import { deploymentFormSections } from './utils';

const useStyles = makeStyles()(() => ({
  defaultBox: { marginTop: 0, marginBottom: -15 },
  heading: { marginBottom: 0 },
  retryInput: { maxWidth: 150, minWidth: 130 },
  wrapper: { minHeight: 300 }
}));

export const ForceDeploy = () => {
  const { control } = useFormContext();
  const { classes } = useStyles();

  return (
    <div>
      <FormCheckbox
        className={classes.heading}
        id={deploymentFormSections.forceDeploy}
        control={control}
        label={
          <div className="flexbox align-items-center">
            <b className="margin-right-small">Force update</b> (optional)
            <MenderHelpTooltip
              id={HELPTOOLTIPS.forceDeployment.id}
              disableFocusListener={false}
              disableHoverListener={false}
              disableTouchListener={false}
              style={{ marginLeft: 15 }}
            />
          </div>
        }
      />
    </div>
  );
};

export const RolloutOptions = ({ isEnterprise }) => {
  const { classes } = useStyles();
  const { watch, setValue } = useFormContext();

  const phases = watch(deploymentFormSections.phases) || [];
  const release = watch(deploymentFormSections.release) || {};

  const updateControlMap = watch(deploymentFormSections.update_control_map) || { states: {} };
  const { states = {} } = updateControlMap;
  const [isPaused, setIsPaused] = useState(!!Object.keys(states).length);

  const onStepChangeClick = step => {
    const { action } = step;
    setValue(deploymentFormSections.update_control_map, { states: { ...states, [step.state]: { action } } });
  };

  const onIsPausedClick = () => setIsPaused(toggle);

  return (
    <>
      <FormControlLabel
        className={classes.heading}
        control={<Checkbox color="primary" checked={isPaused} disabled={!isEnterprise} onChange={onIsPausedClick} size="small" />}
        label={
          <div className="flexbox align-items-center">
            <b className="margin-right-small">Add pauses between update steps</b> (optional)
            <InfoHintContainer>
              <EnterpriseNotification id={BENEFITS.pausedDeployments.id} />
              <DocsTooltip id={DOCSTIPS.pausedDeployments.id} />
            </InfoHintContainer>
          </div>
        }
      />
      <Collapse in={isPaused} className={classes.wrapper}>
        <RolloutSteps disabled={phases.length > 1 || !isEnterprise} onStepChange={onStepChangeClick} release={release} steps={states} />
      </Collapse>
    </>
  );
};

const maxDeploymentRetries = 100;

export const Retries = ({ canRetry, commonClasses, hasNewRetryDefault = false, onSaveRetriesSetting }) => {
  const { classes } = useStyles();

  const onSaveRetriesSettingClick = (_, checked) => onSaveRetriesSetting(checked);

  return (
    <>
      <div className="flexbox align-items-center margin-top-small margin-bottom-small">
        <b className={canRetry ? '' : commonClasses.disabled}>Select the number of times each device will attempt to apply the update</b>
        <InfoHintContainer>
          <EnterpriseNotification id={BENEFITS.retryDeployments.id} />
          <DocsTooltip id={DOCSTIPS.phasedDeployments.id} />
        </InfoHintContainer>
      </div>
      <FormControl disabled={!canRetry}>
        <FormGroup row>
          <NumberInput
            id={deploymentFormSections.retries}
            className={`margin-right ${classes.retryInput}`}
            disabled={!canRetry}
            min={1}
            max={maxDeploymentRetries}
          />
          <FormControlLabel
            className={classes.defaultBox}
            control={<Checkbox checked={hasNewRetryDefault} onChange={onSaveRetriesSettingClick} />}
            label="Save as default"
          />
        </FormGroup>
      </FormControl>
    </>
  );
};
