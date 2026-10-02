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
import { delayUnits, rolloutModes, rolloutPatterns } from './phases/constants';
import { buildPhasePayload } from './utils';

const startTime = '2019-01-01T12:00:00.000Z';

describe('buildPhasePayload', () => {
  it('submits a uniform rollout as a single repeating phase definition', () => {
    expect(
      buildPhasePayload({
        phases: [{ batchSize: 10, delay: 3, delayUnit: delayUnits.hours }],
        rolloutMode: rolloutModes.device_count.key,
        rolloutPattern: rolloutPatterns.uniform.key,
        startTime
      })
    ).toEqual({
      phases: undefined,
      uniform_phases: { batch_size_devices: 10, delay: 3, delayUnit: delayUnits.hours, start_ts: startTime, time_interval: '10800s' }
    });
    const { uniform_phases } = buildPhasePayload({
      phases: [{ batchSize: 25, delay: 1, delayUnit: delayUnits.days }],
      rolloutMode: rolloutModes.percentage.key,
      rolloutPattern: rolloutPatterns.uniform.key
    });
    expect(uniform_phases).toEqual({ batch_size: 25, delay: 1, delayUnit: delayUnits.days, time_interval: '86400s' });
  });
  it('submits a custom rollout with a closing remainder phase', () => {
    const { phases, uniform_phases } = buildPhasePayload({
      phases: [{ batchSize: 10, delay: 1, delayUnit: delayUnits.hours }],
      rolloutMode: rolloutModes.percentage.key,
      rolloutPattern: rolloutPatterns.custom.key,
      startTime
    });
    expect(uniform_phases).toBeUndefined();
    expect(phases).toHaveLength(2);
    expect(phases[0]).toMatchObject({ batch_size: 10, start_ts: startTime });
    expect(phases[1]).not.toHaveProperty('batch_size');
  });
});
