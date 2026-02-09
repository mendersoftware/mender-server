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
import { useEffect, useRef, useState } from 'react';

import { DEPLOYMENT_STATES, TIMEOUTS } from '@northern.tech/store/constants';
import { Deployment } from '@northern.tech/store/deploymentsSlice';
import { groupDeploymentStats } from '@northern.tech/store/utils';
import { DeploymentPhase } from '@northern.tech/types/MenderTypes';
import dayjs from 'dayjs';
import durationDayJs from 'dayjs/plugin/duration';

dayjs.extend(durationDayJs);

interface RolloutPhasesParams {
  currentPhase: { id: string };
  currentProgressCount: number;
  phases: DeploymentPhase[];
  totalDeviceCount: number;
  totalFailureCount: number;
  totalSuccessCount: number;
}

// to display failures per phase we have to approximate the failure count per phase by keeping track of the failures we display in previous phases and
// deduct the phase failures from the remainder - so if we have a total of 5 failures reported and are in the 3rd phase, with each phase before reporting
// 3 successful deployments -> the 3rd phase should end up with 1 failure so far
export const getDisplayableRolloutPhases = ({
  currentPhase,
  currentProgressCount,
  phases,
  totalDeviceCount,
  totalFailureCount,
  totalSuccessCount
}: RolloutPhasesParams) =>
  phases.reduce(
    (accu, phase, index) => {
      const displayablePhase = { ...phase };
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
      const remainingWidth = 100 - accu.countedBatch; // countedBatch should be the summarized percentages of the phases so far
      displayablePhase.width = index === phases.length - 1 ? remainingWidth : displayablePhase.batch_size;
      accu.countedBatch += displayablePhase.batch_size;
      accu.countedFailures += possiblePhaseFailures;
      accu.countedSuccesses += possiblePhaseSuccesses;
      accu.displayablePhases.push(displayablePhase);
      return accu;
    },
    { countedBatch: 0, countedFailures: 0, countedSuccesses: 0, countedProgress: 0, displayablePhases: [] }
  ).displayablePhases;

export const getDeploymentPhasesInfo = (deployment: Deployment) => {
  const { created, device_count = 0, id, phases: deploymentPhases = [], max_devices = 0 } = deployment;
  const {
    inprogress: currentProgressCount,
    successes: totalSuccessCount,
    failures: totalFailureCount
  } = groupDeploymentStats(deployment, deploymentPhases.length < 2);
  const totalDeviceCount = Math.max(device_count, max_devices);

  const phases = deploymentPhases.length ? deploymentPhases : [{ id, device_count: totalSuccessCount, batch_size: 100, start_ts: created }];
  return {
    currentProgressCount,
    phases,
    reversedPhases: phases.slice().reverse(),
    totalDeviceCount,
    totalFailureCount,
    totalSuccessCount
  };
};

export const usePhaseProgress = (deployment: Deployment) => {
  const [time, setTime] = useState(dayjs());
  const timer = useRef<ReturnType<typeof setInterval>>();
  const { status } = deployment;

  useEffect(() => {
    if (status === DEPLOYMENT_STATES.finished) {
      clearInterval(timer.current);
    }
  }, [status]);

  useEffect(() => {
    timer.current = setInterval(() => setTime(dayjs()), TIMEOUTS.oneSecond);
    return () => {
      clearInterval(timer.current);
    };
  }, []);

  const { reversedPhases, totalFailureCount, phases, ...remainder } = getDeploymentPhasesInfo(deployment);
  const currentPhase = reversedPhases.find(phase => dayjs(phase.start_ts) < time) || phases[0];
  const currentPhaseIndex = phases.findIndex(phase => phase.id === currentPhase.id);
  const nextPhaseStart = phases.length > currentPhaseIndex + 1 ? dayjs(phases[currentPhaseIndex + 1].start_ts) : dayjs(time);
  const duration = dayjs.duration(nextPhaseStart.diff(time)).format('DD [days] HH [h] mm [m] ss [s]');

  const displayablePhases = getDisplayableRolloutPhases({
    currentPhase,
    totalFailureCount,
    phases,
    ...remainder
  });

  return {
    currentPhase,
    currentPhaseIndex,
    displayablePhases,
    duration,
    nextPhaseStart,
    phases,
    totalFailureCount
  };
};
