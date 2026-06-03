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
import { useEffect, useState } from 'react';
import type { UseFormWatch } from 'react-hook-form';

import { ALL_DEVICES } from '@northern.tech/store/constants';
import { getDeviceCountsByStatus, getDevicesById, getGroupData } from '@northern.tech/store/selectors';
import { useAppDispatch, useAppSelector } from '@northern.tech/store/store';
import { getGroupDevices } from '@northern.tech/store/thunks';
import type { Device, Filter } from '@northern.tech/types/MenderTypes';
import dayjs from 'dayjs';
import validator from 'validator';

import { rolloutModes } from './phases/constants';
import type { DeploymentFormValues } from './types';

export const deploymentFormSections: Record<keyof DeploymentFormValues, string> = {
  delta: 'delta',
  forceDeploy: 'forceDeploy',
  group: 'group',
  maxDevices: 'maxDevices',
  phases: 'phases',
  release: 'release',
  retries: 'retries',
  rolloutMode: 'rolloutMode',
  startTime: 'startTime',
  uniform_phases: 'uniform_phases',
  update_control_map: 'update_control_map'
};

export const getPhaseStartTime = (phases, index, startDate) => {
  const startingDate = typeof startDate === 'string' && validator.isISO8601(startDate) ? startDate : undefined;
  if (index < 1) {
    return startDate?.toISOString ? startDate.toISOString() : startingDate;
  } else if (phases[index].start_ts && typeof phases[index].start_ts === 'string' && validator.isISO8601(phases[index].start_ts)) {
    return phases[index].start_ts;
  }
  const newStartTime = phases.slice(0, index).reduce((accu, phase) => dayjs(accu).add(phase.delay, phase.delayUnit), startingDate);
  return newStartTime.toISOString();
};

export const buildPhasePayload = ({
  phases = [],
  rolloutMode,
  startTime,
  uniform_phases
}: Pick<DeploymentFormValues, 'phases' | 'rolloutMode' | 'startTime' | 'uniform_phases'>) => {
  if (uniform_phases) {
    return {
      phases: undefined,
      uniform_phases: startTime ? { ...uniform_phases, start_ts: startTime } : uniform_phases
    };
  }
  if (phases.length) {
    return {
      uniform_phases: undefined,
      phases: phases.map((phase, i, origPhases) => {
        const { batch_size, batch_size_devices, start_ts: _st, delay: _d, delayUnit: _du, ...rest } = phase;
        return {
          ...rest,
          start_ts: getPhaseStartTime(origPhases, i, startTime),
          ...(rolloutMode === rolloutModes.device_count.key ? { batch_size_devices } : { batch_size })
        };
      })
    };
  }
  if (startTime) {
    // if there is no existing phase, set phase and start time
    return { phases: [{ batch_size: 100, start_ts: startTime }], uniform_phases: undefined };
  }
  return { phases: undefined, uniform_phases: undefined };
};

export type DeploymentDerivedState = {
  deploymentDeviceCount: number;
  deploymentDeviceIds: string[];
  devices: Device[];
  filter: Filter | undefined;
};

export const useDerivedData = (watch: UseFormWatch<DeploymentFormValues>, initialDevices: Device[] = []): DeploymentDerivedState => {
  const { groups } = useAppSelector(getGroupData);
  const devicesById = useAppSelector(getDevicesById);
  const { accepted: acceptedDeviceCount } = useAppSelector(getDeviceCountsByStatus);
  const dispatch = useAppDispatch();
  const group = watch(deploymentFormSections.group);

  const filter: Filter | undefined = groups[group]?.id ? groups[group] : undefined;

  const [deploymentDeviceCount, setDeploymentDeviceCount] = useState(initialDevices.length);
  const [deploymentDeviceIds, setDeploymentDeviceIds] = useState(initialDevices.map(({ id }) => id));
  const [devices, setDevices] = useState(initialDevices);

  // Compute device count from group selection
  useEffect(() => {
    if (group === ALL_DEVICES) {
      setDeploymentDeviceCount(acceptedDeviceCount);
    } else if (groups[group]) {
      dispatch(getGroupDevices({ group, perPage: 1 }))
        .unwrap()
        .then(
          ({
            payload: {
              group: { total }
            }
          }) => setDeploymentDeviceCount(total)
        );
    } else if (!initialDevices.length) {
      setDeploymentDeviceCount(0);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [acceptedDeviceCount, group, dispatch, JSON.stringify(groups)]);

  // Enrich devices from Redux store when initial devices are provided
  useEffect(() => {
    if (!initialDevices.length) {
      return;
    }
    const deviceIds = initialDevices.map(({ id }) => id);
    const enrichedDevices = initialDevices.map(({ id }) => ({ id, ...(devicesById[id] ?? {}) }) as Device);
    setDeploymentDeviceIds(deviceIds);
    setDeploymentDeviceCount(deviceIds.length);
    setDevices(enrichedDevices);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(initialDevices), JSON.stringify(devicesById)]);

  return {
    deploymentDeviceCount,
    deploymentDeviceIds,
    devices,
    filter
  };
};
