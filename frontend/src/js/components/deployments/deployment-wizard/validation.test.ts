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
import { defaultState } from '@/testUtils';

import { defaultValues } from '../CreateDeployment';
import { rolloutModes, rolloutPatterns } from './phases/constants';
import { deploymentErrors, deploymentResolver } from './validation';

const deploymentCreationTime = defaultState.deployments.byId.d1.created;
const filter = { id: 'filterId', name: 'testGroupDynamic' };

describe('deploymentResolver function', () => {
  const release = { name: 'test-release' };
  const context = { deploymentDeviceCount: 5, devices: [] };

  it('requires a target & software', async () => {
    const { errors, values } = deploymentResolver(defaultValues, { deploymentDeviceCount: 0, devices: [] });
    expect(errors.group?.message).toEqual(deploymentErrors.group);
    expect(errors.release?.message).toEqual(deploymentErrors.release);
    expect(values).toEqual({});
  });
  it('accepts preselected devices in place of a group', async () => {
    const { errors } = deploymentResolver({ ...defaultValues, release }, { deploymentDeviceCount: 1, devices: [defaultState.devices.byId.a1] });
    expect(errors).toEqual({});
  });
  it('passes a fully specified deployment', async () => {
    const { errors, values } = deploymentResolver({ ...defaultValues, group: 'testGroup', release }, context);
    expect(errors).toEqual({});
    expect(values).toEqual({ ...defaultValues, group: 'testGroup', release });
  });
  it('rejects scheduling a deployment for an empty group', async () => {
    const values = { ...defaultValues, group: 'testGroup', release, startTime: deploymentCreationTime };
    expect(deploymentResolver(values, { deploymentDeviceCount: 0, devices: [], group: 'testGroup' }).errors.startTime?.message).toEqual(
      deploymentErrors.emptyGroupSchedule
    );
    expect(deploymentResolver(values, context).errors.startTime).toBeFalsy();
  });
  it('requires a device count once the deployment is limited', async () => {
    const values = { ...defaultValues, group: 'testGroup', release, shouldLimit: true };
    expect(deploymentResolver(values, context).errors.maxDevices?.message).toEqual(deploymentErrors.maxDevicesRequired);
    expect(deploymentResolver({ ...values, maxDevices: undefined }, context).errors.maxDevices?.message).toEqual(deploymentErrors.maxDevicesRequired);
    expect(deploymentResolver({ ...values, maxDevices: -90 }, context).errors.maxDevices?.message).toEqual(deploymentErrors.numberRange);
    expect(deploymentResolver({ ...values, maxDevices: 2 }, context).errors.maxDevices).toBeFalsy();
  });
  it('rejects phases that would not hold any devices', async () => {
    const values = { ...defaultValues, group: 'testGroup', release, usesPattern: true };
    const percentagePhases = [{ batchSize: 10 }, { batchSize: 10 }];
    expect(deploymentResolver({ ...values, phases: percentagePhases }, { ...context, deploymentDeviceCount: 5 }).errors.phases?.message).toMatch(
      /^Phase 1: 10% rounds down to 0 devices/
    );
    expect(deploymentResolver({ ...values, phases: percentagePhases }, { ...context, deploymentDeviceCount: 100 }).errors.phases).toBeFalsy();
    const deviceValues = { ...values, phases: [{ batchSize: 2 }], rolloutMode: rolloutModes.device_count.key };
    expect(deploymentResolver(deviceValues, { ...context, deploymentDeviceCount: 2 }).errors.phases?.message).toEqual(
      'Phase 2: Phases must have at least 1 device'
    );
    expect(deploymentResolver(deviceValues, { ...context, deploymentDeviceCount: 3 }).errors.phases).toBeFalsy();
  });
  it('rejects phases exceeding the target or the device limit', async () => {
    const values = { ...defaultValues, group: 'testGroup', release, rolloutMode: rolloutModes.device_count.key, usesPattern: true };
    const phases = [{ batchSize: 10 }];
    expect(deploymentResolver({ ...values, phases }, context).errors.phases?.message).toEqual('Phase 1: Rollout size exceeds total target group size');
    // a dynamic group can still grow into an oversized phase, so it only warrants a warning
    expect(deploymentResolver({ ...values, phases }, { ...context, filter }).errors.phases).toBeFalsy();
    const limited = { ...values, maxDevices: 2, phases: [{ batchSize: 3 }], shouldLimit: true };
    expect(deploymentResolver(limited, { ...context, deploymentDeviceCount: 20 }).errors.phases?.message).toEqual(
      'Phase 1: Rollout size cannot exceed the maximum number devices'
    );
  });
  it('leaves a rollout without a pattern alone', async () => {
    const values = { ...defaultValues, group: 'testGroup', release, phases: [] };
    expect(deploymentResolver(values, { ...context, deploymentDeviceCount: 0 }).errors.phases).toBeFalsy();
  });
  it('validates a uniform rollout like a repeating phase', async () => {
    const values = {
      ...defaultValues,
      group: 'testGroup',
      release,
      rolloutMode: rolloutModes.device_count.key,
      rolloutPattern: rolloutPatterns.uniform.key,
      usesPattern: true
    };
    const phases = [{ batchSize: 10, delay: 1, delayUnit: 'hours' }];
    expect(deploymentResolver({ ...values, phases }, context).errors.phases?.message).toEqual('Rollout size exceeds total target group size');
    expect(deploymentResolver({ ...values, phases }, { ...context, deploymentDeviceCount: 20 }).errors.phases).toBeFalsy();
  });
});
