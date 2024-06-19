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
import React from 'react';

import { Add as AddIcon, ArrowRight as ArrowRightIcon, PauseCircleOutline as PauseIcon } from '@mui/icons-material';
import { Chip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TIMEOUTS } from '../../../constants/appConstants';
import DocsLink from '../../common/docslink';
import InfoText from '../../common/infotext';
import MenderTooltip from '../../common/mendertooltip';
import { HELPTOOLTIPS, MenderHelpTooltip } from '../../helptips/helptooltips';

const useStyles = makeStyles()(theme => ({
  chip: { marginLeft: theme.spacing(-3), marginRight: theme.spacing(-3) },
  connector: {
    backgroundColor: theme.palette.text.secondary,
    height: 3,
    width: '100%',
    marginRight: theme.spacing(-1),
    '& svg': { color: theme.palette.text.secondary }
  },
  connectorArrowWrapper: { height: theme.spacing(4), paddingLeft: theme.spacing(), width: '100%' },
  connectorWrapper: { minWidth: 120 },
  pauseConnector: { borderLeft: `${theme.palette.grey[500]} dashed 1px`, height: theme.spacing(6), margin: 4, marginTop: -10 },
  stepChip: { minWidth: theme.spacing(11) }
}));

const stepActions = {
  continue: 'continue',
  force_continue: 'force_continue',
  pause: 'pause',
  fail: 'fail'
};

const defaultStep = { action: stepActions.continue };

const defaultSteps = {
  ArtifactDownload: { ...defaultStep, title: 'Download', tooltip: 'The Artifact is downloaded (streamed) to the inactive partition.' },
  ArtifactInstall_Enter: {
    ...defaultStep,
    title: 'Install',
    tooltip: (
      <>
        For <b>operating system updates</b>, this means switching the <i>inactive</i> partition on the device to be <i>active</i> next time the device reboots.
        This means that on the next reboot the device will boot the updated software, regardless if it was rebooted by Mender, an external process or due to
        power loss.
        <br />
        For <b>application updates</b>, it depends on the Update Module but in general refers to the <i>system changing</i> effects; e.g. writing a file to its
        location, running a script, installing or starting a container.
      </>
    ),
    state: 'ArtifactInstall_Enter'
  },
  ArtifactReboot_Enter: {
    ...defaultStep,
    title: 'Reboot',
    tooltip:
      'The device will reboot and the installed update will be active when the device boots up again. As changes are not yet committed, the update is not persistent and the device will still roll back again on the next reboot.',
    state: 'ArtifactReboot_Enter'
  },
  ArtifactCommit_Enter: {
    ...defaultStep,
    title: 'Commit',
    tooltip:
      'If the update passes integrity checks, Mender will mark the update as successful and continue running from this partition. The commit makes the update persistent.',
    state: 'ArtifactCommit_Enter'
  }
};

const menderDemoArtifactName = 'mender-demo-artifact';

export const RolloutStepConnector = ({ disabled, step, onStepChange, release = {} }) => {
  const { classes } = useStyles();
  const onTogglePauseClick = () => {
    onStepChange({ ...step, action: step.action === stepActions.pause ? stepActions.continue : stepActions.pause });
  };

  let stepModifier = { props: {}, toggleButton: undefined };
  if (onStepChange) {
    stepModifier.props = { onDelete: onTogglePauseClick };
    stepModifier.toggleButton = <Chip className={classes.chip} icon={<AddIcon />} label="Add a pause" color="primary" onClick={onTogglePauseClick} />;
  }

  const pauseChip = <Chip className={classes.chip} icon={<PauseIcon />} label="Pause" {...stepModifier.props} />;
  const stepPauseChip =
    step.state === defaultSteps.ArtifactReboot_Enter.state && release.name?.includes(menderDemoArtifactName) ? (
      <MenderTooltip
        arrow
        leaveDelay={TIMEOUTS.oneSecond}
        placement="top"
        title="The demo artifact you selected does not require a reboot and will not pause before starting with the next stage."
      >
        {pauseChip}
      </MenderTooltip>
    ) : (
      pauseChip
    );

  return (
    <div className={`flexbox column center-aligned ${classes.connectorWrapper}`}>
      <div className={`flexbox centered ${classes.connectorArrowWrapper}`}>
        <div className={classes.connector} />
        <ArrowRightIcon fontSize="small" />
      </div>
      {!disabled && (
        <>
          {(onStepChange || step.action === stepActions.pause) && <div className={classes.pauseConnector} />}
          {step.action === stepActions.pause ? stepPauseChip : stepModifier.toggleButton}
        </>
      )}
    </div>
  );
};

export const RolloutSteps = ({ disabled, onStepChange, release, steps = {} }) => {
  const { classes } = useStyles();
  const mappableSteps = Object.entries(defaultSteps).reduce((accu, [key, step]) => [...accu, { ...step, ...steps[key] }], []);

  return (
    <div className={`flexbox margin-top ${onStepChange ? 'margin-left-large margin-right-large' : ''}`}>
      {mappableSteps.map((step, index) => (
        <React.Fragment key={step.title}>
          {index !== 0 && <RolloutStepConnector disabled={disabled} step={step} onStepChange={onStepChange} release={release} />}
          <MenderTooltip title={step.tooltip} arrow>
            <Chip className={classes.stepChip} disabled={disabled} label={step.title} variant="outlined" />
          </MenderTooltip>
        </React.Fragment>
      ))}
    </div>
  );
};

export const RolloutStepsContainer = ({ className = '', disabled, onStepChange, release, steps }) => (
  <div className={className}>
    <div className={disabled ? 'muted' : ''}>
      <RolloutSteps disabled={disabled} onStepChange={onStepChange} release={release} steps={steps} />
      {onStepChange && !disabled && (
        <InfoText>
          A &apos;pause&apos; means each device will pause its update after completing the previous step, and wait for approval before continuing.
          <br />
          You can grant approval by clicking &quot;continue&quot; in the deployment progress UI. <DocsLink path="" title="Learn more" />
        </InfoText>
      )}
    </div>
    {disabled && (
      <div className="flexbox margin-left-small" style={{ marginTop: -30 }}>
        <MenderHelpTooltip id={HELPTOOLTIPS.phasedPausedDeployments.id} />
      </div>
    )}
  </div>
);

export default RolloutStepsContainer;
