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
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

import { Autocomplete, Checkbox, Collapse, FormControl, FormControlLabel, FormGroup, FormHelperText, TextField } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { DOCSTIPS, DocsTooltip } from '@northern.tech/common-ui/DocsLink';
import EnterpriseNotification from '@northern.tech/common-ui/EnterpriseNotification';
import { InfoHintContainer } from '@northern.tech/common-ui/InfoHint';
import { BENEFITS, TIMEOUTS } from '@northern.tech/store/constants';
import { useDebounce } from '@northern.tech/utils/debouncehook';
import { toggle } from '@northern.tech/utils/helpers';

import { HELPTOOLTIPS } from '../../helptips/HelpTooltips';
import { MenderHelpTooltip } from '../../helptips/MenderTooltip';
import RolloutSteps from './RolloutSteps';

const useStyles = makeStyles()(() => ({
  defaultBox: { marginTop: 0, marginBottom: -15 },
  heading: { marginBottom: 0 },
  retryInput: { maxWidth: 150, minWidth: 130 },
  wrapper: { minHeight: 300 }
}));

export const ForceDeploy = ({ deploymentObject, setDeploymentSettings }) => {
  const [forceDeploy, setForceDeploy] = useState(deploymentObject.forceDeploy ?? false);
  const { classes } = useStyles();

  useEffect(() => {
    setDeploymentSettings({ forceDeploy });
  }, [forceDeploy, setDeploymentSettings]);

  return (
    <div>
      <FormControlLabel
        className={classes.heading}
        control={<Checkbox color="primary" checked={forceDeploy} onChange={() => setForceDeploy(toggle)} size="small" />}
        label={
          <div className="flexbox center-aligned">
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

export const RolloutOptions = ({ deploymentObject, isEnterprise, setDeploymentSettings }) => {
  const { phases = [], release = {} } = deploymentObject;
  const { classes } = useStyles();

  const { states = {} } = deploymentObject.update_control_map || {};
  const [isPaused, setIsPaused] = useState(!!Object.keys(states).length);

  const onStepChangeClick = step => {
    const { action } = step;
    setDeploymentSettings({ update_control_map: { states: { ...states, [step.state]: { action } } } });
  };

  const onIsPausedClick = () => setIsPaused(toggle);

  return (
    <>
      <FormControlLabel
        className={classes.heading}
        control={<Checkbox color="primary" checked={isPaused} disabled={!isEnterprise} onChange={onIsPausedClick} size="small" />}
        label={
          <div className="flexbox center-aligned">
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

export const Retries = ({ canRetry, commonClasses, deploymentObject, previousRetries, setDeploymentSettings }) => {
  const { retries } = deploymentObject;
  const { classes } = useStyles();

  const [currentAttempts, setCurrentAttempts] = useState(Number(retries ?? previousRetries ?? 0));
  const debouncedAttempts = useDebounce(currentAttempts, TIMEOUTS.debounceShort);

  useEffect(() => {
    setDeploymentSettings({ retries: Number(debouncedAttempts) - 1 });
  }, [debouncedAttempts, setDeploymentSettings]);

  const formatValue = value => {
    const newValue = Math.max(0, Math.min(value, 100));
    return newValue ? `${newValue}` : '';
  };

  const onInputChange = (e, value, reason) => {
    if (reason === 'clear') {
      return setDeploymentSettings({ retries: 0 });
    } else if ((reason === 'reset' && !e) || reason === 'blur') {
      return;
    }
    setCurrentAttempts(formatValue(value));
  };

  return (
    <>
      <div className="flexbox center-aligned margin-top-small margin-bottom-small">
        <b className={canRetry ? '' : commonClasses.disabled}>Select the number of times each device will attempt to apply the update</b>
        <InfoHintContainer>
          <EnterpriseNotification id={BENEFITS.retryDeployments.id} />
          <DocsTooltip id={DOCSTIPS.phasedDeployments.id} />
        </InfoHintContainer>
      </div>
      <FormControl disabled={!canRetry}>
        <FormGroup row>
          <Autocomplete
            autoSelect
            autoHighlight
            className={`margin-right ${classes.retryInput}`}
            freeSolo
            disabled={!canRetry}
            getOptionLabel={formatValue}
            handleHomeEndKeys
            id="deployment-retries-selection"
            options={[1, 2, 3, 4]}
            onInputChange={onInputChange}
            renderInput={params => (
              <TextField
                {...params}
                className={classes.retryInput}
                slotProps={{
                  htmlInput: { ...params.inputProps, value: formatValue(params.inputProps.value) },
                  input: { ...params.InputProps }
                }}
                type="number"
                placeholder={`${previousRetries + 1}`}
              />
            )}
            value={currentAttempts}
          />
        </FormGroup>
        <FormHelperText>
          Default retry attempts: {previousRetries}. You can change the organization-wide default in the <Link to="/settings/global-settings">settings</Link>.
        </FormHelperText>
      </FormControl>
    </>
  );
};
