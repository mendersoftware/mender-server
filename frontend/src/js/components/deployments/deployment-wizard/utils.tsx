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
import { useCallback, useEffect, useState } from 'react';
import type { UseFormWatch } from 'react-hook-form';
import { useFormContext } from 'react-hook-form';

import { HelpOutlineOutlined as HelpIcon } from '@mui/icons-material';
import { Tooltip } from '@mui/material';

import { ALL_DEVICES } from '@northern.tech/store/constants';
import { getDeviceCountsByStatus, getDevicesById, getGroupData } from '@northern.tech/store/selectors';
import { useAppDispatch, useAppSelector } from '@northern.tech/store/store';
import { getGroupDevices } from '@northern.tech/store/thunks';
import type { Device, Filter } from '@northern.tech/types/MenderTypes';
import dayjs from 'dayjs';
import validator from 'validator';

import { delayDefaults, phaseLimits, rolloutModes, rolloutPatterns } from './phases/constants';
import { delayToSeconds } from './phases/utils';
import type { DeploymentFormValues } from './types';

export const deploymentFormSections: { [K in keyof DeploymentFormValues]-?: K } = {
  delta: 'delta',
  forceDeploy: 'forceDeploy',
  group: 'group',
  isPaused: 'isPaused',
  maxDevices: 'maxDevices',
  phases: 'phases',
  release: 'release',
  retries: 'retries',
  rolloutMode: 'rolloutMode',
  rolloutPattern: 'rolloutPattern',
  shouldLimit: 'shouldLimit',
  startTime: 'startTime',
  update_control_map: 'update_control_map',
  usesPattern: 'usesPattern'
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

// the form tracks a single list of phase definitions with one batch size each - only here they split back into the
// api's custom phases (the definitions + a closing remainder phase) or the repeating uniform phase definition
export const buildPhasePayload = ({
  phases = [],
  rolloutMode,
  rolloutPattern,
  startTime
}: Pick<DeploymentFormValues, 'phases' | 'rolloutMode' | 'rolloutPattern' | 'startTime'>) => {
  const batchKey = rolloutMode === rolloutModes.device_count.key ? rolloutModes.device_count.batchKey : rolloutModes.percentage.batchKey;
  if (phases.length && rolloutPattern === rolloutPatterns.uniform.key) {
    const [{ batchSize, delay = delayDefaults.delay, delayUnit = delayDefaults.delayUnit }] = phases;
    return {
      phases: undefined,
      uniform_phases: { [batchKey]: batchSize, time_interval: delayToSeconds(delay, delayUnit), ...(startTime ? { start_ts: startTime } : {}) }
    };
  }
  if (phases.length) {
    const withFinalPhase = [...phases, {}];
    return {
      uniform_phases: undefined,
      phases: withFinalPhase.map(({ batchSize }, i) => ({
        start_ts: getPhaseStartTime(withFinalPhase, i, startTime),
        ...(i < phases.length ? { [batchKey]: batchSize } : {})
      }))
    };
  }
  if (startTime) {
    // if there is no phased rollout, a single full size phase carries the start time
    return { phases: [{ batch_size: phaseLimits.fullBatchPercentage, start_ts: startTime }], uniform_phases: undefined };
  }
  return { phases: undefined, uniform_phases: undefined };
};

// most of the form is written through setValue, which doesn't re-run the validation unless it is told to - and it has
// to, so that an error the user just resolved goes away right away instead of lingering until the next submit attempt
export const useValidatedSetValue = () => {
  const {
    formState: { isSubmitted },
    setValue
  } = useFormContext();
  return useCallback((name, value) => setValue(name, value, { shouldValidate: isSubmitted }), [isSubmitted, setValue]);
};

export type DeploymentDerivedState = {
  deploymentDeviceCount: number;
  deploymentDeviceIds: string[];
  devices: Device[];
  filter: Filter | undefined;
  isDeviceCountResolved: boolean;
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
  const [isDeviceCountResolved, setIsDeviceCountResolved] = useState(!!initialDevices.length);

  // Compute device count from group selection
  useEffect(() => {
    if (group === ALL_DEVICES) {
      setDeploymentDeviceCount(acceptedDeviceCount);
      setIsDeviceCountResolved(true);
    } else if (groups[group]) {
      setIsDeviceCountResolved(false);
      dispatch(getGroupDevices({ group, perPage: 1 }))
        .unwrap()
        .then(result => {
          const total = result?.payload?.group?.total ?? 0;
          setDeploymentDeviceCount(total);
          setIsDeviceCountResolved(true);
        })
        .catch(() => {
          setDeploymentDeviceCount(0);
          setIsDeviceCountResolved(true);
        });
    } else if (!initialDevices.length) {
      setDeploymentDeviceCount(0);
      setIsDeviceCountResolved(false);
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
    setIsDeviceCountResolved(true);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(initialDevices), JSON.stringify(devicesById)]);

  return {
    deploymentDeviceCount,
    deploymentDeviceIds,
    devices,
    filter,
    isDeviceCountResolved
  };
};

export const DisabledReasonHint = ({ reason }: { reason?: string }) =>
  reason ? (
    <Tooltip arrow placement="top" title={reason}>
      <HelpIcon color="action" fontSize="small" />
    </Tooltip>
  ) : null;
