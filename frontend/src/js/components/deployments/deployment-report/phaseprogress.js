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
import React, { useEffect, useState } from 'react';

import { CheckCircle, ErrorRounded, Pause, PlayArrow, Warning as WarningIcon } from '@mui/icons-material';
import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import pluralize from 'pluralize';

import inprogressImage from '../../../../assets/img/pending_status.png';
import { deploymentDisplayStates, deploymentSubstates, installationSubstatesMap, pauseMap } from '../../../constants/deploymentConstants';
import { getDeploymentState, groupDeploymentStats, statCollector } from '../../../helpers';
import Confirm from '../../common/confirm';
import { ProgressChartComponent } from '../progressChart';

const useStyles = makeStyles()(theme => ({
  active: { borderLeftColor: theme.palette.text.primary },
  borderColor: { borderLeftStyle: 'solid', borderLeftWidth: 1, height: '100%', zIndex: 1 },
  continueButton: { marginRight: theme.spacing(2) },
  failureIcon: { fontSize: 16, marginRight: 10 },
  inactive: { borderLeftColor: theme.palette.grey[500] },
  phaseInfo: { marginBottom: theme.spacing() },
  phaseIndex: { margin: theme.spacing(0.5) }
}));

const shortCircuitIndicators = [deploymentSubstates.alreadyInstalled, deploymentSubstates.noartifact];

const substateIconMap = {
  finished: { state: 'finished', icon: <CheckCircle fontSize="small" /> },
  inprogress: { state: 'inprogress', icon: <img src={inprogressImage} /> },
  failed: { state: 'failed', icon: <ErrorRounded fontSize="small" /> },
  paused: { state: 'paused', icon: <Pause fontSize="small" /> },
  pendingPause: { state: 'pendingPause', icon: <Pause fontSize="small" color="disabled" /> }
};

const stepTotalWidth = 100 / Object.keys(installationSubstatesMap).length;

const PhaseDelimiter = ({ classes, compact, index, phase = {} }) => {
  const { classes: localClasses } = useStyles();
  const { status, substate: phaseSubstate } = phase;
  const isActive = status === substateIconMap.inprogress.state;
  const substate = phaseSubstate.done;

  const offset = `${stepTotalWidth * (index + 1) - stepTotalWidth / 2}%`;
  return (
    <div className={`${classes.phaseDelimiter} ${compact ? 'compact' : ''}`} style={{ left: offset, width: `${stepTotalWidth}%` }}>
      <div className={`${localClasses.borderColor} ${isActive ? localClasses.active : localClasses.inactive}`} />
      {substateIconMap[status] ? substateIconMap[status].icon : <div />}
      {compact ? <div /> : <div className="capitalized slightly-smaller">{substate}</div>}
    </div>
  );
};

const determineSubstateStatus = (successes, failures, totalDeviceCount, pauseIndicator, hasPauseDefined) => {
  let status;
  if (successes === totalDeviceCount) {
    status = substateIconMap.finished.state;
  } else if (failures === totalDeviceCount) {
    status = substateIconMap.failed.state;
  } else if (pauseIndicator) {
    status = substateIconMap.paused.state;
  } else if (successes || failures) {
    status = substateIconMap.inprogress.state;
  } else if (hasPauseDefined) {
    status = substateIconMap.pendingPause.state;
  }
  return status;
};

const getDisplayablePhases = ({ pauseMap, deployment, stepWidth, substatesMap, totalDeviceCount }) => {
  const { statistics = {}, update_control_map = {} } = deployment;
  const { status: stats = {} } = statistics;
  const currentPauseState = Object.keys(pauseMap)
    .reverse()
    .find(key => stats[key] > 0);
  return Object.values(substatesMap).reduce(
    (accu, substate, index) => {
      let successes = statCollector(substate.successIndicators, stats);
      let failures = statCollector(substate.failureIndicators, stats);
      if (
        !currentPauseState ||
        index <= Object.keys(pauseMap).indexOf(currentPauseState) ||
        (index && accu.displayablePhases[index - 1].failures + accu.displayablePhases[index - 1].success === totalDeviceCount)
      ) {
        failures = accu.displayablePhases[index - 1]?.failures || failures;
        successes = successes + accu.shortCutSuccesses;
      }
      successes = Math.min(totalDeviceCount, successes);
      failures = Math.min(totalDeviceCount - successes, failures);
      const successWidth = (successes / totalDeviceCount) * 100 || 0;
      const failureWidth = (failures / totalDeviceCount) * 100 || 0;
      const { states = {} } = update_control_map;
      const hasPauseDefined = states[substate.pauseConfigurationIndicator]?.action === 'pause';
      const status = determineSubstateStatus(successes, failures, totalDeviceCount, !!stats[substate.pauseIndicator], hasPauseDefined);
      accu.displayablePhases.push({ substate, successes, failures, offset: stepWidth * index, width: stepWidth, successWidth, failureWidth, status });
      return accu;
    },
    { displayablePhases: [], shortCutSuccesses: statCollector(shortCircuitIndicators, stats) }
  ).displayablePhases;
};

const statusMap = {
  inprogress: <PlayArrow fontSize="inherit" />,
  paused: <Pause fontSize="inherit" />
};

const Header = ({ device_count, failures, totalDeviceCount, showDetails, status }) => {
  const { classes } = useStyles();
  return showDetails ? (
    <>
      <div className={`flexbox space-between ${classes.phaseInfo}`}>
        {statusMap[status] ? statusMap[status] : <div />}
        <div className="flexbox center-aligned">
          {!!failures && <WarningIcon className={classes.failureIcon} />}
          {`${failures} ${pluralize('failure', failures)}`}
        </div>
      </div>
      <div className={`muted slightly-smaller ${classes.phaseIndex}`}>Phase 1 of 1</div>
    </>
  ) : (
    <>
      Phase 1: {Math.round((device_count / totalDeviceCount || 0) * 100)}% ({device_count} {pluralize('device', device_count)})
    </>
  );
};

export const PhaseProgressDisplay = ({ className = '', deployment, showDetails = true, status, ...remainder }) => {
  const { failures } = groupDeploymentStats(deployment);

  const { device_count = 0, max_devices = 0 } = deployment;
  const totalDeviceCount = Math.max(device_count, max_devices);

  const displayablePhases = getDisplayablePhases({ deployment, pauseMap, stepWidth: stepTotalWidth, substatesMap: installationSubstatesMap, totalDeviceCount });

  return (
    <ProgressChartComponent
      className={`flexbox column stepped-progress ${showDetails ? '' : 'no-background'} ${className}`}
      phases={displayablePhases}
      PhaseDelimiter={PhaseDelimiter}
      Header={Header}
      headerProps={{ device_count, failures, showDetails, status, totalDeviceCount }}
      {...remainder}
    />
  );
};

const confirmationStyle = {
  justifyContent: 'flex-start',
  paddingLeft: 100
};

const PhaseLabel = ({ classes, phase }) => <div className={`capitalized progress-step-number ${classes.progressStepNumber}`}>{phase.substate.title}</div>;

export const PhaseProgress = ({ className = '', deployment = {}, onAbort, onUpdateControlChange }) => {
  const { classes } = useStyles();
  const [shouldContinue, setShouldContinue] = useState(false);
  const [shouldAbort, setShouldAbort] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const { id, statistics = {}, update_control_map = {} } = deployment;
  const { status: stats = {} } = statistics;
  const { states = {} } = update_control_map;
  const { failures: totalFailureCount, paused: totalPausedCount } = groupDeploymentStats(deployment);

  const status = getDeploymentState(deployment);
  const currentPauseState = Object.keys(pauseMap)
    .reverse()
    .find(key => stats[key] > 0);

  useEffect(() => {
    if (!isLoading) {
      return;
    }
    setIsLoading(false);
  }, [isLoading, status]);

  const onAbortClick = () => {
    setShouldAbort(false);
    onAbort(id);
  };

  const onContinueClick = () => {
    setIsLoading(true);
    setShouldContinue(false);
    onUpdateControlChange({ states: { [pauseMap[currentPauseState].followUp]: { action: 'continue' } } });
  };

  const isPaused = status === deploymentDisplayStates.paused;
  const canContinue = isPaused && states[pauseMap[currentPauseState].followUp];
  const disableContinuationButtons = isLoading || (canContinue && states[pauseMap[currentPauseState].followUp].action !== 'pause');
  return (
    <div className={`flexbox column ${className}`}>
      <PhaseProgressDisplay compact deployment={deployment} PhaseLabel={PhaseLabel} showDetails={false} status={status} />
      <div className="margin-top">
        Deployment is <span className="uppercased">{status}</span> with {totalFailureCount} {pluralize('failure', totalFailureCount)}
        {isPaused && !canContinue && ` - waiting for an action on the ${pluralize('device', totalPausedCount)} to continue`}
      </div>
      {canContinue && (
        <div className="margin-top margin-bottom relative">
          {shouldContinue && (
            <Confirm
              type="deploymentContinuation"
              classes="confirmation-overlay"
              action={onContinueClick}
              cancel={() => setShouldContinue(false)}
              style={confirmationStyle}
            />
          )}
          {shouldAbort && (
            <Confirm
              type="deploymentAbort"
              classes="confirmation-overlay"
              action={onAbortClick}
              cancel={() => setShouldAbort(false)}
              style={confirmationStyle}
            />
          )}
          <Button color="primary" disabled={disableContinuationButtons} onClick={setShouldContinue} variant="contained" className={classes.continueButton}>
            Continue
          </Button>
          <Button disabled={disableContinuationButtons} onClick={setShouldAbort}>
            Abort
          </Button>
        </div>
      )}
    </div>
  );
};

export default PhaseProgress;
