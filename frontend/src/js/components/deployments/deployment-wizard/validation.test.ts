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
import {
  deploymentErrors,
  deploymentResolver,
  disabledReasons,
  getDeviceLimitDisabledReason,
  getPausesDisabledReason,
  getRolloutPatternDisabledReason
} from './validation';

const deploymentCreationTime = defaultState.deployments.byId.d1.created;
const filter = { id: 'filterId', name: 'testGroupDynamic' };

describe('disabled reasons', () => {
  it('keeps the options available until a target rules them out', async () => {
    expect(getDeviceLimitDisabledReason({ deploymentDeviceCount: 0, filter: undefined, group: null })).toEqual('');
    expect(getRolloutPatternDisabledReason({ deploymentDeviceCount: 0, filter: undefined, group: null })).toEqual('');
    expect(getPausesDisabledReason({ deploymentDeviceCount: 0, filter: undefined, group: null })).toEqual('');
  });
  it('rules out limiting the device count for static groups & direct device targets', async () => {
    expect(getDeviceLimitDisabledReason({ deploymentDeviceCount: 5, filter: undefined, group: 'testGroup' })).toEqual(disabledReasons.staticGroupLimit);
    expect(getDeviceLimitDisabledReason({ deploymentDeviceCount: 5, filter, group: 'testGroupDynamic' })).toEqual('');
    expect(getDeviceLimitDisabledReason({ deploymentDeviceCount: 1, devices: [defaultState.devices.byId.a1], filter: undefined, group: null })).toEqual(
      disabledReasons.deviceTargetLimit
    );
  });
  it('rules out the rollout options for empty groups', async () => {
    const emptyGroup = { deploymentDeviceCount: 0, filter, group: 'testGroupDynamic', isDeviceCountResolved: true };
    expect(getRolloutPatternDisabledReason(emptyGroup)).toEqual(disabledReasons.emptyGroupPattern);
    expect(getPausesDisabledReason(emptyGroup)).toEqual(disabledReasons.emptyGroupPauses);
  });
  it('gives a group the benefit of the doubt while its device count is being retrieved', async () => {
    const loadingGroup = { deploymentDeviceCount: 0, filter, group: 'testGroupDynamic', isDeviceCountResolved: false };
    expect(getRolloutPatternDisabledReason(loadingGroup)).toEqual('');
    expect(getPausesDisabledReason(loadingGroup)).toEqual('');
  });
  it('lets the rollout options exclude each other', async () => {
    const target = { deploymentDeviceCount: 5, filter, group: 'testGroupDynamic' };
    expect(getRolloutPatternDisabledReason({ ...target, isPaused: true })).toEqual(disabledReasons.pausedPattern);
    expect(getPausesDisabledReason({ ...target, usesPattern: true })).toEqual(disabledReasons.patternPauses);
  });
});

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
    expect(deploymentResolver(values, context).errors.maxDevices?.message).toEqual(deploymentErrors.maxDevices);
    expect(deploymentResolver({ ...values, maxDevices: 2 }, context).errors.maxDevices).toBeFalsy();
  });
});
