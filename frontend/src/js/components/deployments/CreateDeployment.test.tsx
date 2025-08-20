// Copyright 2020 Northern.tech AS
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
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import { defaultState, render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { vi } from 'vitest';

import CreateDeployment from './CreateDeployment';
import { DeviceLimit } from './deployment-wizard/DeviceLimit';
import { RolloutPatternSelection, getPhaseDeviceCount, getRemainderPercent, validatePhases } from './deployment-wizard/PhaseSettings';
import { ForceDeploy, Retries, RolloutOptions } from './deployment-wizard/RolloutOptions';
import { ScheduleRollout } from './deployment-wizard/ScheduleRollout';
import { Devices, ReleasesWarning, Software } from './deployment-wizard/SoftwareDevices';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.features,
      isEnterprise: false,
      isHosted: false
    }
  }
};
const deploymentCreationTime = defaultState.deployments.byId.d1.created;

describe('CreateDeployment Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<CreateDeployment deploymentObject={{}} setDeploymentSettings={vi.fn()} />, { preloadedState });
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  describe('smaller components', () => {
    [DeviceLimit, Devices, ForceDeploy, ReleasesWarning, Software, ScheduleRollout, Retries, RolloutOptions, RolloutPatternSelection].forEach(Component => {
      const getReleasesMock = vi.fn();
      getReleasesMock.mockResolvedValue();
      const props = {
        commonClasses: { columns: 'test' },
        deploymentObject: { phases: [{ batch_size: 0 }] },
        getReleases: getReleasesMock,
        getSystemDevices: vi.fn(),
        groups: defaultState.devices.groups.byId,
        groupNames: ['testGroup', 'testGroupDynamic'],
        hasDynamicGroups: true,
        open: true,
        previousRetries: 0,
        releases: Object.keys(defaultState.releases.byId),
        releasesById: defaultState.releases.byId,
        setDeploymentSettings: vi.fn()
      };
      it(`renders ${Component.displayName || Component.name} correctly`, () => {
        const { baseElement } = render(
          <LocalizationProvider dateAdapter={AdapterDayjs}>
            <Component {...props} />
          </LocalizationProvider>,
          { preloadedState }
        );
        const view = baseElement.lastChild;
        expect(view).toMatchSnapshot();
        expect(view).toEqual(expect.not.stringMatching(undefineds));
        expect(view).toBeTruthy();
      });
      it(`renders ${Component.displayName || Component.name} correctly as enterprise`, () => {
        const { baseElement } = render(
          <LocalizationProvider dateAdapter={AdapterDayjs}>
            <Component {...props} isEnterprise />
          </LocalizationProvider>,
          { preloadedState }
        );
        const view = baseElement.lastChild;
        expect(view).toMatchSnapshot();
        expect(view).toEqual(expect.not.stringMatching(undefineds));
        expect(view).toBeTruthy();
      });
    });
  });

  describe('utility functions', () => {
    describe('getPhaseDeviceCount function', () => {
      it('works with empty attributes', async () => {
        expect(getPhaseDeviceCount(120, 10, 20, false)).toEqual(12);
        expect(getPhaseDeviceCount(120, 10, 20, true)).toEqual(12);
        expect(getPhaseDeviceCount(120, null, 20, true)).toEqual(24);
        expect(getPhaseDeviceCount(120, null, 20, false)).toEqual(24);
        expect(getPhaseDeviceCount(undefined, null, 20, false)).toEqual(0);
      });
    });
    describe('getRemainderPercent function', () => {
      it('remainder Percent calculated correctly', async () => {
        const phases = [
          { batch_size: 10, not: 'interested' },
          { batch_size: 10, not: 'interested' },
          { batch_size: 10, not: 'interested' }
        ];
        expect(getRemainderPercent(phases)).toEqual(80);
        expect(
          getRemainderPercent([
            { batch_size: 10, not: 'interested' },
            { batch_size: 90, not: 'interested' }
          ])
        ).toEqual(90);
        expect(
          getRemainderPercent([
            { batch_size: 10, not: 'interested' },
            { batch_size: 95, not: 'interested' }
          ])
        ).toEqual(90);
        // this will be caught in the phase validation - should still be good to be fixed in the future
        expect(
          getRemainderPercent([
            { batch_size: 50, not: 'interested' },
            { batch_size: 55, not: 'interested' },
            { batch_size: 95, not: 'interested' }
          ])
        ).toEqual(-5);
      });
    });

    describe('validatePhases function', () => {
      it('works as expected', async () => {
        const phases = [
          {
            batch_size: 10,
            delay: 2,
            delayUnit: 'hours',
            start_ts: deploymentCreationTime
          },
          { batch_size: 10, delay: 2, start_ts: deploymentCreationTime },
          { batch_size: 10, start_ts: deploymentCreationTime }
        ];
        expect(validatePhases(undefined, 10000)).toEqual(true);
        expect(validatePhases(undefined, 10000)).toEqual(true);
        expect(validatePhases(phases, 10)).toEqual(true);
        expect(validatePhases(phases, 10)).toEqual(true);
        expect(validatePhases([], 10)).toEqual(true);
        expect(
          validatePhases(
            [
              { batch_size: 50, not: 'interested' },
              { batch_size: 55, not: 'interested' },
              { batch_size: 95, not: 'interested' }
            ],
            100
          )
        ).toEqual(false);
      });
    });
  });
});
