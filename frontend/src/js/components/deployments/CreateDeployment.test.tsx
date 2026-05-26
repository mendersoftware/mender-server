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
import { FormProvider, useForm } from 'react-hook-form';

import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import { defaultState, render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { vi } from 'vitest';

import CreateDeployment, { defaultValues as formDefaultValues } from './CreateDeployment';
import { DeviceLimit } from './deployment-wizard/DeviceLimit';
import { RolloutPatternSelection } from './deployment-wizard/PhaseSettings';
import { ForceDeploy, Retries, RolloutOptions } from './deployment-wizard/RolloutOptions';
import { ScheduleRollout } from './deployment-wizard/ScheduleRollout';
import { Devices, ReleasesWarning, Software } from './deployment-wizard/SoftwareDevices';

const FormWrapper = ({ children, defaultValues = {} }) => {
  const methods = useForm({
    defaultValues: {
      ...formDefaultValues,
      ...defaultValues
    }
  });
  return <FormProvider {...methods}>{children}</FormProvider>;
};

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
    const { baseElement } = render(<CreateDeployment deploymentObject={{}} onValuesChange={vi.fn()} open />, { preloadedState });
    const view = baseElement.getElementsByClassName('MuiDrawer-root')[0];
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
        defaultRetries: 0,
        releases: Object.keys(defaultState.releases.byId),
        releasesById: defaultState.releases.byId
      };
      it(`renders ${Component.displayName || Component.name} correctly`, () => {
        const { baseElement } = render(
          <LocalizationProvider dateAdapter={AdapterDayjs}>
            <FormWrapper>
              <Component {...props} />
            </FormWrapper>
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
            <FormWrapper>
              <Component {...props} isEnterprise />
            </FormWrapper>
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
