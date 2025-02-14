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

import { ArrowForward } from '@mui/icons-material';
import { Chip } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import { TwoColumnData } from '@northern.tech/common-ui/ConfigurationObject';
import LinedHeader from '@northern.tech/common-ui/LinedHeader';
import Time from '@northern.tech/common-ui/Time';
import { DEPLOYMENT_STATES } from '@northern.tech/store/constants';
import { formatTime } from '@northern.tech/utils/helpers';
import dayjs from 'dayjs';
import durationDayJs from 'dayjs/plugin/duration';
import isSameOrAfter from 'dayjs/plugin/isSameOrAfter';
import pluralize from 'pluralize';

import { ProgressChartComponent, getDeploymentPhasesInfo, getDisplayablePhases } from '../ProgressChart';
import { defaultColumnDataProps } from '../Report';
import { getPhaseDeviceCount, getPhaseStartTime, getRemainderPercent } from '../deployment-wizard/PhaseSettings';
import PhaseProgress from './PhaseProgress';

const useStyles = makeStyles()(theme => ({
  currentPhaseInfo: { backgroundColor: theme.palette.grey[400] },
  phaseInfo: { maxWidth: maxPhaseWidth, borderRadius: 5, paddingTop: theme.spacing(), paddingBottom: theme.spacing(3) },
  phaseIndex: { alignSelf: 'flex-start', margin: theme.spacing(2), marginLeft: theme.spacing(2.5) },
  phasesOverviewArrow: { marginLeft: theme.spacing(4), marginRight: theme.spacing(4) }
}));

dayjs.extend(durationDayJs);
dayjs.extend(isSameOrAfter);

const maxPhaseWidth = 270;

const PhaseLabel = ({ index }) => <div className="capitalized progress-step-number muted">Phase {index + 1}</div>;

export const RolloutSchedule = ({ deployment, headerClass, innerRef, onAbort, onUpdateControlChange }) => {
  const { classes } = useStyles();
  const now = dayjs();
  const { created: creationTime = now.toISOString(), filter, finished, status, update_control_map } = deployment;

  const { phases, reversedPhases, totalDeviceCount, ...remainder } = getDeploymentPhasesInfo(deployment);

  const start_time = phases[0].start_ts || creationTime;
  const currentPhase = reversedPhases.find(phase => now.isAfter(phase.start_ts)) || phases[0];
  const currentPhaseIndex = phases.findIndex(phase => phase.id === currentPhase.id);
  const currentPhaseStartTime = getPhaseStartTime(phases, currentPhaseIndex, start_time);
  let currentPhaseTime = 'N/A';
  if (now.isSameOrAfter(currentPhaseStartTime)) {
    currentPhaseTime = currentPhaseIndex + 1;
  }
  const endTime = finished ? <Time value={formatTime(finished)} /> : filter ? 'N/A' : '-';

  const displayablePhases = getDisplayablePhases({ currentPhase, phases, totalDeviceCount, ...remainder });
  return (
    <>
      <LinedHeader className={`margin-top-large ${headerClass}`} heading="Schedule details" innerRef={innerRef} />
      {phases.length > 1 || !update_control_map ? (
        <>
          <div className="flexbox">
            <TwoColumnData
              {...defaultColumnDataProps}
              config={{
                'Start time': <Time value={formatTime(start_time)} />,
                'Current phase': currentPhaseTime
              }}
            />
            <ArrowForward className={classes.phasesOverviewArrow} />
            <TwoColumnData {...defaultColumnDataProps} config={{ 'End time': endTime }} />
          </div>
          <ProgressChartComponent className="margin-top no-background" phases={displayablePhases} PhaseLabel={PhaseLabel} />
        </>
      ) : (
        <PhaseProgress deployment={deployment} onAbort={onAbort} onUpdateControlChange={onUpdateControlChange} />
      )}
      <div className="deployment-phases-report margin-top margin-bottom" style={{ gridTemplateColumns: `repeat(auto-fit, ${maxPhaseWidth}px)` }}>
        {phases.map((phase, index) => {
          const batchSize = phase.batch_size || getRemainderPercent(phases);
          const deviceCount = getPhaseDeviceCount(totalDeviceCount, batchSize, batchSize, index === phases.length - 1);
          const deviceCountText = !filter ? ` (${deviceCount} ${pluralize('device', deviceCount)})` : '';
          const startTime = phase.start_ts ?? getPhaseStartTime(phases, index, start_time);
          const phaseObject = {
            'Phase start time': <Time value={startTime} />,
            'Batch size': <div className="muted">{`${batchSize}%${deviceCountText}`}</div>
          };
          let phaseTitle = status !== DEPLOYMENT_STATES.scheduled ? <div className="muted">Complete</div> : null;
          let isCurrentPhase = false;
          if (now.isBefore(startTime)) {
            const duration = dayjs.duration(dayjs(startTime).diff(now));
            phaseTitle = <div>{`Begins in ${duration.format('DD [days] HH [h] mm [m]')}`}</div>;
          } else if (status === DEPLOYMENT_STATES.inprogress && phase.id === currentPhase.id) {
            phaseTitle = <div className="muted">Current phase</div>;
            isCurrentPhase = true;
          }
          return (
            <div className={`flexbox column centered ${classes.phaseInfo} ${isCurrentPhase ? classes.currentPhaseInfo : ''}`} key={startTime}>
              {phaseTitle}
              <Chip className={classes.phaseIndex} size="small" label={`Phase ${index + 1}`} />
              <TwoColumnData {...defaultColumnDataProps} style={{ ...defaultColumnDataProps.style, alignSelf: 'initial' }} config={phaseObject} />
            </div>
          );
        })}
      </div>
    </>
  );
};

export default RolloutSchedule;
