// Copyright 2025 Northern.tech AS
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

import { CheckCircle, ErrorRounded, Pause } from '@mui/icons-material';
import { Button } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Confirm from '@northern.tech/common-ui/Confirm';
import { deploymentDisplayStates, deploymentSubstates, installationSubstatesMap, pauseMap } from '@northern.tech/store/constants';
import { Deployment } from '@northern.tech/store/deploymentsSlice';
import { getDeploymentState, groupDeploymentStats, statCollector } from '@northern.tech/store/utils';
import pluralize from 'pluralize';

import inprogressImage from '../../../../assets/img/pending_status.png';

const stepTotalWidth = 100 / Object.keys(installationSubstatesMap).length;

export const substateIconMap = {
  finished: { state: 'finished', icon: <CheckCircle fontSize="small" /> },
  inprogress: { state: 'inprogress', icon: <img src={inprogressImage} /> },
  failed: { state: 'failed', icon: <ErrorRounded fontSize="small" /> },
  paused: { state: 'paused', icon: <Pause fontSize="small" /> },
  pendingPause: { state: 'pendingPause', icon: <Pause fontSize="small" color="disabled" /> }
};

const useStyles = makeStyles()(theme => ({
  container: {
    backgroundColor: theme.palette.background.default,
    padding: '10px 20px',
    borderRadius: theme.spacing(0.5),
    alignContent: 'center',
    minHeight: 70,
    '.progress-bar': { height: theme.spacing(0.5) },
    '.progress-chart': { minHeight: 45 },
    '.progress-step': { minHeight: 45 },
    '.progress-step, .progress-step-total': {
      position: 'absolute',
      borderRightStyle: 'none'
    },
    '.progress-step-total .progress-bar': { backgroundColor: theme.palette.grey[50] },
    '.progress-step-number': { alignSelf: 'flex-start', marginTop: theme.spacing(-1) },
    '&.no-background': { background: 'none' },
    '&.stepped-progress .progress-step-total': { marginLeft: '-0.25%', width: '100.5%' },
    '&.stepped-progress .progress-step-total .progress-bar': {
      backgroundColor: theme.palette.background.default,
      border: `1px solid ${theme.palette.grey[800]}`,
      borderRadius: 2,
      height: theme.spacing()
    },
    '&.stepped-progress .progress-step': { minHeight: 20 }
  },
  phaseDelimiter: {
    display: 'grid',
    rowGap: 4,
    placeItems: 'center',
    position: 'absolute',
    gridTemplateRows: '20px 1.25rem min-content',
    top: theme.spacing(1.5),
    zIndex: 2
  },
  active: { borderLeftColor: theme.palette.text.primary },
  borderColor: { borderLeftStyle: 'solid', borderLeftWidth: 1, height: '100%', zIndex: 1 },
  failureIcon: { fontSize: 16, marginRight: 10 },
  inactive: { borderLeftColor: theme.palette.grey[500] },
  phaseInfo: { marginBottom: theme.spacing() }
}));

const shortCircuitIndicators = [deploymentSubstates.alreadyInstalled, deploymentSubstates.noartifact];

interface SubstatePhase {
  failures: number;
  failureWidth: number;
  offset: number;
  status?: string;
  substate: {
    failureIndicators: string[];
    pauseConfigurationIndicator: string;
    pauseIndicator: string;
    successIndicators: string[];
    title: string;
  };
  successes: number;
  successWidth: number;
  width: number;
}

const determineSubstateStatus = (successes: number, failures: number, totalDeviceCount: number, pauseIndicator: boolean, hasPauseDefined: boolean) => {
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

export const getDisplayableSubstatePhases = ({ deployment, totalDeviceCount }: { deployment: Deployment; totalDeviceCount: number }): SubstatePhase[] => {
  const { statistics = {}, update_control_map = {} } = deployment;
  const { status: stats = {} } = statistics;
  const currentPauseState = Object.keys(pauseMap)
    .reverse()
    .find(key => stats[key] > 0);
  return Object.values(installationSubstatesMap).reduce(
    (accu, substate, index) => {
      let successes = statCollector(substate.successIndicators, stats);
      let failures = statCollector(substate.failureIndicators, stats);
      if (
        !currentPauseState ||
        index <= Object.keys(pauseMap).indexOf(currentPauseState) ||
        (index && accu.displayablePhases[index - 1].failures + accu.displayablePhases[index - 1].successes === totalDeviceCount)
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
      accu.displayablePhases.push({ substate, successes, failures, offset: stepTotalWidth * index, width: stepTotalWidth, successWidth, failureWidth, status });
      return accu;
    },
    { displayablePhases: [] as SubstatePhase[], shortCutSuccesses: statCollector(shortCircuitIndicators, stats) }
  ).displayablePhases;
};

const SubstateHeader = ({ device_count, totalDeviceCount }: { device_count: number; totalDeviceCount: number }) => (
  <>
    Phase 1: {Math.round((device_count / totalDeviceCount || 0) * 100)}% ({device_count} {pluralize('device', device_count)})
  </>
);

const SubstateDelimiter = ({ index, phase }: { index: number; phase: SubstatePhase }) => {
  const { classes } = useStyles();
  const { status } = phase;
  const isActive = status === substateIconMap.inprogress.state;
  const icon = substateIconMap[status as keyof typeof substateIconMap]?.icon;

  const offset = `${stepTotalWidth * (index + 1) - stepTotalWidth / 2}%`;
  return (
    <div className={classes.phaseDelimiter} style={{ left: offset, width: `${stepTotalWidth}%` }}>
      <div className={`${classes.borderColor} ${isActive ? classes.active : classes.inactive}`} />
      {icon ? icon : <div />}
    </div>
  );
};

const SubstatePhaseLabel = ({ phase }: { phase: SubstatePhase }) => <div className="capitalized progress-step-number">{phase.substate.title}</div>;

const SubstateProgressChart = ({ phases }: { phases: SubstatePhase[] }) => (
  <div className="flexbox relative margin-top-small">
    {phases.map((phase, index) => (
      <React.Fragment key={`substate-phase-${index}`}>
        <div className="progress-step" style={{ left: `${phase.offset}%`, width: `${phase.width}%` }}>
          <SubstatePhaseLabel phase={phase} />
          <div style={{ display: 'contents' }}>
            <div className="progress-bar green" style={{ width: `${phase.successWidth}%` }} />
            <div className="progress-bar warning" style={{ left: `${phase.successWidth}%`, width: `${phase.failureWidth}%` }} />
          </div>
        </div>
        {index !== phases.length - 1 && <SubstateDelimiter index={index} phase={phase} />}
      </React.Fragment>
    ))}
    <div className="progress-step relative flexbox progress-step-total">
      <div className="progress-bar" />
    </div>
  </div>
);

const confirmationStyle = {
  justifyContent: 'flex-start',
  paddingLeft: 100
};

interface SubstateProgressBarProps {
  className?: string;
  deployment: Deployment;
  onAbort: (id: string) => void;
  onUpdateControlChange: (update: { states: Record<string, { action: string }> }) => void;
}

export const SubstateProgressBar = ({ className = '', deployment, onAbort, onUpdateControlChange }: SubstateProgressBarProps) => {
  const { classes } = useStyles();
  const [shouldContinue, setShouldContinue] = useState(false);
  const [shouldAbort, setShouldAbort] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const { id, device_count, max_devices, statistics = {}, update_control_map = {} } = deployment;
  const { status: stats = {} } = statistics;
  const { states = {} } = update_control_map;
  const { failures: totalFailureCount, paused: totalPausedCount } = groupDeploymentStats(deployment);
  const totalDeviceCount = Math.max(device_count, max_devices);

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
    if (!currentPauseState || !pauseMap[currentPauseState]) {
      return;
    }
    setIsLoading(true);
    setShouldContinue(false);
    onUpdateControlChange({ states: { [pauseMap[currentPauseState as keyof typeof pauseMap].followUp]: { action: 'continue' } } });
  };

  const displayablePhases = getDisplayableSubstatePhases({ deployment, totalDeviceCount });

  const isPaused = status === deploymentDisplayStates.paused;
  const canContinue = isPaused && currentPauseState && states[pauseMap[currentPauseState as keyof typeof pauseMap].followUp];
  const disableContinuationButtons =
    isLoading || (canContinue && currentPauseState && states[pauseMap[currentPauseState as keyof typeof pauseMap].followUp]?.action !== 'pause');

  return (
    <div className={`flexbox column ${className}`}>
      <div className={`relative flexbox column ${classes.container}`}>
        <SubstateHeader device_count={device_count} totalDeviceCount={totalDeviceCount} />
        <SubstateProgressChart phases={displayablePhases} />
      </div>
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
          <Button disabled={disableContinuationButtons} onClick={() => setShouldContinue(true)} variant="contained" className="margin-right">
            Continue
          </Button>
          <Button disabled={disableContinuationButtons} onClick={() => setShouldAbort(true)}>
            Abort
          </Button>
        </div>
      )}
    </div>
  );
};
