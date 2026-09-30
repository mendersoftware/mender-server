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

import { getDisplayablePhases } from './utils';

const phasedDeployment = {
  ...defaultState.deployments.byId.d3,
  phases: [
    { id: 'a1b2c3d4-0000-4000-8000-000000000001', batch_size: 50, device_count: 50, status: 'finished' },
    { id: 'a1b2c3d4-0000-4000-8000-000000000002', batch_size: 50, device_count: 10, status: 'inprogress' }
  ]
};

describe('getDisplayablePhases', () => {
  it('uses the phase position as id instead of the backend provided id', () => {
    expect(getDisplayablePhases(phasedDeployment).map(({ id }) => id)).toEqual(['1', '2']);
  });
  it('marks only the last phase of a phased rollout as final', () => {
    expect(getDisplayablePhases(phasedDeployment).map(({ isFinal }) => isFinal)).toEqual([false, true]);
    const [singlePhase] = getDisplayablePhases({ ...phasedDeployment, phases: [phasedDeployment.phases[0]] });
    expect(singlePhase.isFinal).toBeFalsy();
  });
  it('marks the last phase as growing only for uncapped deployments to dynamic groups', () => {
    const filter = { id: 'filter1', name: 'dynamic group', filters: [] };
    expect(getDisplayablePhases(phasedDeployment).some(({ isGrowing }) => isGrowing)).toBeFalsy();
    expect(getDisplayablePhases({ ...phasedDeployment, filter, max_devices: 0 }).map(({ isGrowing }) => isGrowing)).toEqual([false, true]);
    expect(getDisplayablePhases({ ...phasedDeployment, filter, max_devices: 80 }).some(({ isGrowing }) => isGrowing)).toBeFalsy();
  });
  it('falls back to a single phase covering the whole deployment', () => {
    const { created, device_count, status } = defaultState.deployments.byId.d2;
    expect(getDisplayablePhases({ ...defaultState.deployments.byId.d2, phases: undefined })).toEqual([
      { id: '1', device_count, initial_batch_device_count: device_count, isFinal: false, isGrowing: false, start_ts: created, status }
    ]);
  });
});
