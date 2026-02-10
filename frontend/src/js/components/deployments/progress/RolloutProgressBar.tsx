// Copyright 2026 Northern.tech AS
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

import { Warning as WarningIcon } from '@mui/icons-material';
import { LinearProgress, Tooltip, Typography } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import Time from '@northern.tech/common-ui/Time';
import { deploymentDisplayStates } from '@northern.tech/store/constants';
import { Deployment } from '@northern.tech/store/deploymentsSlice';
import { getDeploymentState, groupDeploymentStats } from '@northern.tech/store/utils';
import pluralize from 'pluralize';

import { DeploymentStatusNotification } from './DeploymentStatusNotification';
import { usePhaseProgress } from './usePhaseProgress';

export type ProgressVariant = 'dashboard' | 'list' | 'report';

const useStyles = makeStyles()(theme => ({
  container: {
    backgroundColor: theme.palette.background.default,
    padding: '10px 20px',
    borderRadius: theme.spacing(0.5),
    alignContent: 'center',
    minHeight: 70,
    '.progress-bar': { height: theme.spacing(0.5) },
    '.chart-container': { minHeight: 70 },
    '.progress-chart': { minHeight: 45 },
    '.progress-step': { minHeight: 45 },
    '.progress-step, .progress-step-total': {
      position: 'absolute',
      borderRightStyle: 'none'
    },
    '.progress-step-total .progress-bar': { backgroundColor: theme.palette.grey[50] },
    '.progress-step-number': { alignSelf: 'flex-start', marginTop: theme.spacing(-0.5) },
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
  inProgressBar: { backgroundColor: theme.palette.grey[400] },
  dualPanel: {
    display: 'grid',
    gridTemplateColumns: '2fr 1fr',
    gridColumnGap: theme.spacing(2),
    alignItems: 'center',
    '.progress-chart.detailed': {
      minHeight: 20
    }
  },
  defaultDelimiter: { borderRight: '1px dashed', zIndex: 10 }
}));

interface FooterProps {
  currentPhaseIndex: number;
  duration: string;
  nextPhaseStart: Date;
  phasesCount: number;
}

const Footer = ({ currentPhaseIndex, duration, nextPhaseStart, phasesCount }: FooterProps) => (
  <div className="flexbox space-between">
    <Typography variant="caption">Devices in progress</Typography>
    {phasesCount > 1 && phasesCount > currentPhaseIndex + 1 ? (
      <Typography variant="caption">
        <span>Time until next phase: </span>
        <Tooltip title={<Time value={nextPhaseStart} />} placement="top">
          <span>{duration}</span>
        </Tooltip>
      </Typography>
    ) : (
      <Typography variant="caption">{`Current phase: ${currentPhaseIndex + 1} of ${phasesCount}`}</Typography>
    )}
  </div>
);

const Side = ({ totalFailureCount }: { totalFailureCount: number }) => (
  <Typography variant="body2" className={`flexbox center-aligned ${totalFailureCount ? 'warning' : ''}`} style={{ justifyContent: 'flex-end' }}>
    {!!totalFailureCount && <WarningIcon style={{ fontSize: 16, marginRight: 10 }} />}
    {`${totalFailureCount} ${pluralize('failure', totalFailureCount)}`}
  </Typography>
);

const PhaseLabel = ({ index }: { index: number }) => <div className="capitalized progress-step-number muted">Phase {index + 1}</div>;

interface DisplayablePhase {
  failureWidth: number;
  id?: string;
  offset: number;
  progressWidth?: number;
  successWidth: number;
  width: number;
}

const MultiPhaseProgress = ({ phases, showPhaseLabel }: { phases: DisplayablePhase[]; showPhaseLabel: boolean }) => {
  const { classes } = useStyles();
  return (
    <div className="progress-chart relative detailed">
      {phases.map((phase, index) => (
        <React.Fragment key={phase.id ?? `deployment-phase-${index}`}>
          <div className="progress-step" style={{ left: `${phase.offset}%`, width: `${phase.width}%` }}>
            {showPhaseLabel && <PhaseLabel index={index} />}
            {!!phase.progressWidth && <div className={`progress-bar ${classes.inProgressBar}`} style={{ width: `${phase.progressWidth}%` }} />}
            <div style={{ display: 'contents' }}>
              <div className="progress-bar green" style={{ width: `${phase.successWidth}%` }} />
              <div className="progress-bar warning" style={{ left: `${phase.successWidth}%`, width: `${phase.failureWidth}%` }} />
            </div>
          </div>
          {index !== phases.length - 1 && <div className={`absolute ${classes.defaultDelimiter}`} style={{ left: `${phase.offset}%` }} />}
        </React.Fragment>
      ))}
      <div className="progress-step relative flexbox progress-step-total">
        <div className="progress-bar" />
      </div>
    </div>
  );
};

const determinateStates = {
  [deploymentDisplayStates.finished]: { variant: 'determinate' as const, value: 100 },
  queued: { variant: 'determinate' as const, value: 0 },
  default: { variant: 'indeterminate' as const, value: undefined }
};

export const SimpleProgress = ({ deployment }: { deployment: Deployment }) => {
  const { phases = [] } = deployment;
  const { failures } = groupDeploymentStats(deployment, false);
  const status = getDeploymentState(deployment);
  const phaseFailures = phases.reduce((accu, phase) => {
    const { failures = 0 } = phase as { failures?: number };
    return accu + failures;
  }, 0);
  const variantProps = determinateStates[status] ?? determinateStates.default;
  return <LinearProgress color={failures || phaseFailures ? 'secondary' : 'primary'} {...variantProps} />;
};

interface VariantProps {
  showFooter: boolean;
  showHeader: boolean;
  showPhaseLabel: boolean;
  showSide: boolean;
}

const variantPropsMap: { default: VariantProps } & Record<ProgressVariant, Partial<VariantProps>> = {
  default: { showFooter: false, showHeader: false, showPhaseLabel: false, showSide: false },
  dashboard: {},
  list: { showFooter: true, showHeader: true, showSide: true },
  report: { showPhaseLabel: true }
};

export const RolloutProgressBar = ({ className = '', deployment, variant }: { className?: string; deployment: Deployment; variant: ProgressVariant }) => {
  const { classes } = useStyles();
  const { showHeader, showFooter, showPhaseLabel, showSide } = { ...variantPropsMap.default, ...variantPropsMap[variant] };

  const { currentPhaseIndex, displayablePhases, duration, nextPhaseStart, phases, totalFailureCount } = usePhaseProgress(deployment);
  const status = getDeploymentState(deployment);
  const isMultiPhase = displayablePhases.length > 1;

  const progressContent = isMultiPhase ? (
    <MultiPhaseProgress phases={displayablePhases} showPhaseLabel={showPhaseLabel} />
  ) : (
    <>
      {showPhaseLabel && <PhaseLabel index={0} />}
      <SimpleProgress deployment={deployment} />
    </>
  );

  return (
    <div className={`relative flexbox column ${classes.container} ${className}`}>
      {showHeader && <DeploymentStatusNotification status={status} />}
      <div className={showSide ? classes.dualPanel : 'chart-container'}>
        {progressContent}
        {showSide && <Side totalFailureCount={totalFailureCount} />}
      </div>
      {showFooter && <Footer currentPhaseIndex={currentPhaseIndex} duration={duration} nextPhaseStart={nextPhaseStart.toDate()} phasesCount={phases.length} />}
    </div>
  );
};
