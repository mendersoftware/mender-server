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
import { useFormContext } from 'react-hook-form';

import { HelpOutlineOutlined as HelpIcon } from '@mui/icons-material';
import { Alert, Collapse, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTextLink } from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import Link from '@northern.tech/common-ui/Link';
import { FormCheckbox } from '@northern.tech/common-ui/forms/FormCheckbox';
import { NumberInput } from '@northern.tech/common-ui/forms/NumberInput';
import { BENEFITS } from '@northern.tech/store/constants';

import RolloutSteps from './RolloutSteps';
import { deploymentFormSections, useValidatedSetValue } from './utils';

const useStyles = makeStyles()(() => ({
  wrapper: { minHeight: 300 }
}));

export const ForceDeploy = () => {
  const { control } = useFormContext();

  return (
    <div>
      <FormCheckbox
        id={deploymentFormSections.forceDeploy}
        control={control}
        label="Force update if the software is already installed"
        slotProps={{ checkbox: { className: 'margin-left-small', size: 'small' } }}
      />
    </div>
  );
};

export const RolloutOptions = ({ isEnterprise }) => {
  const { classes } = useStyles();
  const { watch } = useFormContext();
  const setValue = useValidatedSetValue();

  const phases = watch(deploymentFormSections.phases) || [];
  const release = watch(deploymentFormSections.release) || {};

  const updateControlMap = watch(deploymentFormSections.update_control_map) || { states: {} };
  const { states = {} } = updateControlMap;
  const isPaused = watch(deploymentFormSections.isPaused);

  const onStepChangeClick = step => {
    const { action } = step;
    setValue(deploymentFormSections.update_control_map, { states: { ...states, [step.state]: { action } } });
  };

  return (
    <>
      <FormCheckbox
        id={deploymentFormSections.isPaused}
        disabled={!isEnterprise}
        label={
          <div className="flexbox align-items-center">
            Add pauses between update steps
            <InfoHintContainer>
              <EnterpriseNotification id={BENEFITS.pausedDeployments.id} />
              <DocsTextLink id={DOCSTIPS.pausedDeployments.id} />
            </InfoHintContainer>
          </div>
        }
        slotProps={{ checkbox: { className: 'margin-left-small', size: 'small' } }}
      />
      <Collapse in={isPaused} className={classes.wrapper}>
        <Alert severity="warning" className="margin-top-small margin-bottom-small" variant="outlined">
          This feature was removed in Mender Client 4.0. To manage phased deployments for newer devices, we recommend using rollout patterns instead.
        </Alert>
        <RolloutSteps disabled={phases.length > 1 || !isEnterprise} onStepChange={onStepChangeClick} release={release} steps={states} />
      </Collapse>
    </>
  );
};

export const maxDeploymentRetries = 100;

export const Retries = ({ canManageUsers, canRetry, commonClasses, defaultRetries }) => (
  <>
    <div className="flexbox align-items-center">
      <Typography className={canRetry ? '' : commonClasses.disabled} variant="subtitle1">
        Set the number of times each device will attempt this update
      </Typography>
      <InfoHintContainer>
        <EnterpriseNotification id={BENEFITS.retryDeployments.id} />
      </InfoHintContainer>
    </div>
    <div className="flexbox align-items-center margin-top-x-small margin-bottom-small">
      {/* input validation needs to be handled via validation.ts due to the cross form checks */}
      <NumberInput id={deploymentFormSections.retries} disabled={!canRetry} min={0} max={maxDeploymentRetries} showSteps size="small" width={120} />
      <Tooltip arrow placement="top" title={`Default is ${defaultRetries + 1}. This can be changed in the global settings`}>
        <HelpIcon className="margin-left-x-small margin-right-x-small" color="action" />
      </Tooltip>
      {canManageUsers && (
        <Link to="/settings/global-settings" target="_blank">
          Change global default
        </Link>
      )}
    </div>
  </>
);
