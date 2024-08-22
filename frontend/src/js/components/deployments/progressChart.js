// Copyright 2016 Northern.tech AS
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
import React, { useEffect, useRef, useState } from 'react';

import { RotateLeftOutlined, Warning as WarningIcon } from '@mui/icons-material';
import { Tooltip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { mdiDotsHorizontalCircleOutline as QueuedIcon, mdiSleep as SleepIcon } from '@mdi/js';
import moment from 'moment';
import momentDurationFormatSetup from 'moment-duration-format';
import pluralize from 'pluralize';

import { TIMEOUTS } from '../../constants/appConstants';
import { groupDeploymentStats } from '../../helpers';
import MaterialDesignIcon from '../common/materialdesignicon';
import Time from '../common/time';

momentDurationFormatSetup(moment);

const statusMap = {
  complete: {
    icon: <MaterialDesignIcon path={SleepIcon} />,
    description: () => 'Complete, awaiting new devices'
  },
  queued: {
    icon: <MaterialDesignIcon path={QueuedIcon} />,
    description: () => 'Queued to start'
  },
  paused: { icon: <RotateLeftOutlined fontSize="inherit" />, description: window => `Paused until next window ${window}` }
};

const useStyles = makeStyles()(theme => ({
  container: {
    backgroundColor: theme.palette.grey[400],
    padding: '10px 20px',
    borderRadius: theme.spacing(0.5),
    justifyContent: 'center',
    minHeight: 70,
    '.chart-container': { minHeight: 70 },
    '.progress-chart': { minHeight: 45 },
    '.compact .progress-step, .detailed .progress-step': { minHeight: 45 },
    '.progress-step, .detailed .progress-step-total': {
      position: 'absolute',
      borderRightStyle: 'none'
    },
    '.progress-step-total .progress-bar': { backgroundColor: theme.palette.grey[50] },
    '.progress-step-number': { alignSelf: 'flex-start', marginTop: theme.spacing(-0.5) },
    '&.minimal': { padding: 'initial' },
    '&.no-background': { background: 'none' },
    '&.stepped-progress .progress-step-total': { marginLeft: '-0.25%', width: '100.5%' },
    '&.stepped-progress .progress-step-total .progress-bar': {
      backgroundColor: theme.palette.background.default,
      border: `1px solid ${theme.palette.grey[800]}`,
      borderRadius: 2,
      height: 12
    },
    '&.stepped-progress .detailed .progress-step': { minHeight: 20 }
  },
  dualPanel: {
    display: 'grid',
    gridTemplateColumns: '2fr 1fr',
    gridColumnGap: theme.spacing(2),
    '.progress-chart.detailed': {
      minHeight: 20,
      alignItems: 'center'
    }
  },
  defaultDelimiter: { borderRight: '1px dashed', zIndex: 10 },
  phaseDelimiter: {
    display: 'grid',
    rowGap: 4,
    placeItems: 'center',
    position: 'absolute',
    gridTemplateRows: `20px 1.25rem min-content`,
    zIndex: 2,
    ['&.compact']: {
      gridTemplateRows: '45px 1.25rem min-content'
    }
  }
}));

export const ProgressChartComponent = ({
  className,
  compact,
  Footer,
  footerProps,
  Header,
  headerProps,
  minimal,
  PhaseDelimiter,
  PhaseLabel,
  phases,
  showDelimiter,
  Side,
  sideProps,
  ...remainder
}) => {
  const { classes } = useStyles();
  return (
    <div className={`relative ${classes.container} ${minimal ? 'minimal' : ''} ${className}`}>
      {!minimal && Header && <Header {...headerProps} />}
      <div className={!minimal && Side ? classes.dualPanel : 'chart-container'}>
        <div className={`progress-chart relative ${compact ? 'compact' : 'detailed'}`}>
          {phases.map((phase, index) => {
            const commonProps = { ...remainder, compact, index, phase, classes };
            return (
              <React.Fragment key={phase.id ?? `deployment-phase-${index}`}>
                <div className="progress-step" style={{ left: `${phase.offset}%`, width: `${phase.width}%` }}>
                  {!minimal && PhaseLabel && <PhaseLabel {...commonProps} />}
                  {!!phase.progressWidth && <div className="progress-bar" style={{ width: `${phase.progressWidth}%`, backgroundColor: '#aaa' }} />}
                  <div style={{ display: 'contents' }}>
                    <div className="progress-bar green" style={{ width: `${phase.successWidth}%` }} />
                    <div className="progress-bar warning" style={{ left: `${phase.successWidth}%`, width: `${phase.failureWidth}%` }} />
                  </div>
                </div>
                {(showDelimiter || PhaseDelimiter) && index !== phases.length - 1 && (
                  <>
                    {PhaseDelimiter ? (
                      <PhaseDelimiter {...commonProps} />
                    ) : (
                      <div className={`absolute ${classes.defaultDelimiter}`} style={{ left: `${phase.offset}%` }}></div>
                    )}
                  </>
                )}
              </React.Fragment>
            );
          })}
          <div className="progress-step relative flexbox progress-step-total">
            <div className="progress-bar"></div>
          </div>
        </div>
        {!minimal && Side && <Side compact={compact} {...remainder} {...sideProps} />}
      </div>
      {!minimal && Footer && <Footer {...footerProps} />}
    </div>
  );
};

// to display failures per phase we have to approximate the failure count per phase by keeping track of the failures we display in previous phases and
// deduct the phase failures from the remainder - so if we have a total of 5 failures reported and are in the 3rd phase, with each phase before reporting
// 3 successful deployments -> the 3rd phase should end up with 1 failure so far
export const getDisplayablePhases = ({ currentPhase, currentProgressCount, phases, totalDeviceCount, totalFailureCount, totalSuccessCount }) =>
  phases.reduce(
    (accu, phase, index) => {
      let displayablePhase = { ...phase };
      // ongoing phases might not have a device_count yet - so we calculate it
      let expectedDeviceCountInPhase = Math.floor((totalDeviceCount / 100) * displayablePhase.batch_size) || displayablePhase.batch_size;
      // for phases with more successes than phase.device_count or more failures than phase.device_count we have to guess what phase to put them in =>
      // because of that we have to limit per phase success/ failure counts to the phase.device_count and split the progress between those with a bias for success,
      // therefore we have to track the remaining width and work with it - until we get per phase success & failure information
      let leftoverDevices = expectedDeviceCountInPhase;
      const possiblePhaseSuccesses = Math.max(Math.min(displayablePhase.device_count, totalSuccessCount - accu.countedSuccesses), 0);
      leftoverDevices -= possiblePhaseSuccesses;
      const possiblePhaseFailures = Math.max(Math.min(leftoverDevices, totalFailureCount - accu.countedFailures), 0);
      leftoverDevices -= possiblePhaseFailures;
      const possiblePhaseProgress = Math.max(Math.min(leftoverDevices, currentProgressCount - accu.countedProgress), 0);
      // if there are too few devices in a phase to register, fallback to occured deployments, as those have definitely happened
      expectedDeviceCountInPhase = Math.max(expectedDeviceCountInPhase, possiblePhaseSuccesses + possiblePhaseProgress + possiblePhaseFailures, 0);
      displayablePhase.successWidth = (possiblePhaseSuccesses / expectedDeviceCountInPhase) * 100 || 0;
      displayablePhase.failureWidth = (possiblePhaseFailures / expectedDeviceCountInPhase) * 100 || 0;
      if (displayablePhase.id === currentPhase.id || leftoverDevices > 0) {
        displayablePhase.progressWidth = (possiblePhaseProgress / expectedDeviceCountInPhase) * 100;
        accu.countedProgress += possiblePhaseProgress;
      }
      displayablePhase.offset = accu.countedBatch;
      const remainingWidth = 100 - (100 / totalDeviceCount) * accu.countedBatch;
      displayablePhase.width = index === phases.length - 1 ? remainingWidth : displayablePhase.batch_size;
      accu.countedBatch += displayablePhase.batch_size;
      accu.countedFailures += possiblePhaseFailures;
      accu.countedSuccesses += possiblePhaseSuccesses;
      accu.displayablePhases.push(displayablePhase);
      return accu;
    },
    { countedBatch: 0, countedFailures: 0, countedSuccesses: 0, countedProgress: 0, displayablePhases: [] }
  ).displayablePhases;

export const DeploymentStatusNotification = ({ status }) => (
  <div className="flexbox center-aligned">
    {statusMap[status].icon}
    <span className="margin-left-small">{statusMap[status].description()}</span>
  </div>
);

const Header = ({ status }) => (
  <>
    {statusMap[status] && (
      <span className="flexbox center-aligned small muted">
        {statusMap[status].icon}
        <span className="margin-left-small">{statusMap[status].description()}</span>
      </span>
    )}
  </>
);

const Footer = ({ currentPhaseIndex, duration, nextPhaseStart, phasesCount }) => (
  <div className="flexbox space-between muted">
    <div>Devices in progress</div>
    {phasesCount > 1 && phasesCount > currentPhaseIndex + 1 ? (
      <div>
        <span>Time until next phase: </span>
        <Tooltip title={<Time value={nextPhaseStart.toDate()} />} placement="top">
          <span>{duration}</span>
        </Tooltip>
      </div>
    ) : (
      <div>{`Current phase: ${currentPhaseIndex + 1} of ${phasesCount}`}</div>
    )}
  </div>
);

const Side = ({ totalFailureCount }) => (
  <div className={`flexbox center-aligned ${totalFailureCount ? 'warning' : 'muted'}`} style={{ justifyContent: 'flex-end' }}>
    {!!totalFailureCount && <WarningIcon style={{ fontSize: 16, marginRight: 10 }} />}
    {`${totalFailureCount} ${pluralize('failure', totalFailureCount)}`}
  </div>
);

export const getDeploymentPhasesInfo = deployment => {
  const { created, device_count = 0, id, phases: deploymentPhases = [], max_devices = 0 } = deployment;
  const {
    inprogress: currentProgressCount,
    successes: totalSuccessCount,
    failures: totalFailureCount
  } = groupDeploymentStats(deployment, deploymentPhases.length < 2);
  const totalDeviceCount = Math.max(device_count, max_devices);

  let phases = deploymentPhases.length ? deploymentPhases : [{ id, device_count: totalSuccessCount, batch_size: 100, start_ts: created }];
  return {
    currentProgressCount,
    phases,
    reversedPhases: phases.slice().reverse(),
    totalDeviceCount,
    totalFailureCount,
    totalSuccessCount
  };
};

export const ProgressDisplay = ({ className = '', deployment, minimal = false, status }) => {
  const [time, setTime] = useState(new Date());
  const timer = useRef();

  useEffect(() => {
    timer.current = setInterval(updateTime, TIMEOUTS.oneSecond);
    return () => {
      clearInterval(timer.current);
    };
  }, []);

  const updateTime = () => setTime(new Date());

  const { reversedPhases, totalFailureCount, phases, ...remainder } = getDeploymentPhasesInfo(deployment);

  const currentPhase = reversedPhases.find(phase => new Date(phase.start_ts) < time) || phases[0];
  const currentPhaseIndex = phases.findIndex(phase => phase.id === currentPhase.id);
  const nextPhaseStart = phases.length > currentPhaseIndex + 1 ? moment(phases[currentPhaseIndex + 1].start_ts) : moment(time);

  const displayablePhases = getDisplayablePhases({
    currentPhase,
    totalFailureCount,
    phases,
    ...remainder
  });

  const momentaryTime = moment(time);
  const duration = moment.duration(nextPhaseStart.diff(momentaryTime)).format('d [days] hh [h] mm [m] ss [s]');

  return (
    <ProgressChartComponent
      className={className}
      Footer={Footer}
      footerProps={{ currentPhaseIndex, duration, nextPhaseStart, phasesCount: phases.length }}
      Header={Header}
      headerProps={{ status }}
      minimal={minimal}
      phases={displayablePhases}
      Side={Side}
      sideProps={{ totalFailureCount }}
    />
  );
};

export default ProgressDisplay;
