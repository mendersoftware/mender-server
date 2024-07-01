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
import React from 'react';

import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterMoment } from '@mui/x-date-pickers/AdapterMoment';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import CreateDeployment from './createdeployment';
import { RolloutPatternSelection } from './deployment-wizard/phasesettings';
import { ForceDeploy, Retries, RolloutOptions } from './deployment-wizard/rolloutoptions';
import { ScheduleRollout } from './deployment-wizard/schedulerollout';
import { Devices, ReleasesWarning, Software } from './deployment-wizard/softwaredevices';

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

describe('CreateDeployment Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<CreateDeployment deploymentObject={{}} setDeploymentSettings={jest.fn} />, { preloadedState });
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  describe('smaller components', () => {
    [Devices, ForceDeploy, ReleasesWarning, Software, ScheduleRollout, Retries, RolloutOptions, RolloutPatternSelection].forEach(Component => {
      const getReleasesMock = jest.fn();
      getReleasesMock.mockResolvedValue();
      const props = {
        commonClasses: { columns: 'test' },
        deploymentObject: { phases: [{ batch_size: 0 }] },
        getReleases: getReleasesMock,
        getSystemDevices: jest.fn,
        groups: defaultState.devices.groups.byId,
        hasDynamicGroups: true,
        open: true,
        previousRetries: 0,
        releases: Object.keys(defaultState.releases.byId),
        releasesById: defaultState.releases.byId,
        setDeploymentSettings: jest.fn
      };
      it(`renders ${Component.displayName || Component.name} correctly`, () => {
        const { baseElement } = render(
          <LocalizationProvider dateAdapter={AdapterMoment}>
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
          <LocalizationProvider dateAdapter={AdapterMoment}>
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
});
