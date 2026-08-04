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
import GeneralApi from '@northern.tech/store/api/general-api';
import { ALL_DEVICES } from '@northern.tech/store/constants';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import CreateDeployment, { defaultValues as formDefaultValues } from './CreateDeployment';
import { DeviceLimit } from './deployment-wizard/DeviceLimit';
import { RolloutPatternSelection } from './deployment-wizard/PhaseSettings';
import { ForceDeploy, Retries, RolloutOptions } from './deployment-wizard/RolloutOptions';
import { ScheduleRollout } from './deployment-wizard/ScheduleRollout';
import { Devices, ReleasesWarning, Software } from './deployment-wizard/SoftwareDevices';
import { deploymentErrors } from './deployment-wizard/validation';

const FormWrapper = ({ children, defaultValues = {} }) => {
  const methods = useForm({
    defaultValues: {
      ...formDefaultValues,
      ...defaultValues
    }
  });
  return <FormProvider {...methods}>{children}</FormProvider>;
};

// the rollout options exclude each other, so they have to be enabled one at a time to render their expanded state
const expandedDefaultValues = { RolloutOptions: { isPaused: true }, RolloutPatternSelection: { usesPattern: true } };

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

const renderWrapper = ({ deploymentObject = {}, onScheduleSubmit = vi.fn(), preloadedState: preloadedStateProp = preloadedState }) =>
  render(
    <LocalizationProvider dateAdapter={AdapterDayjs}>
      <CreateDeployment deploymentObject={deploymentObject} onScheduleSubmit={onScheduleSubmit} onValuesChange={vi.fn()} open />
    </LocalizationProvider>,
    { preloadedState: preloadedStateProp }
  );

describe('CreateDeployment Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = renderWrapper({});
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
            <FormWrapper defaultValues={expandedDefaultValues[Component.name] ?? {}}>
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
            <FormWrapper defaultValues={expandedDefaultValues[Component.name] ?? {}}>
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

  describe('validation', () => {
    it('accepts a click on an incomplete deployment & points out what is missing', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      const onScheduleSubmit = vi.fn();
      renderWrapper({ onScheduleSubmit });
      const submitButton = screen.getByRole('button', { name: /create deployment/i });
      expect(submitButton).toBeEnabled();
      await user.click(submitButton);
      await waitFor(() => expect(screen.getByText(deploymentErrors.release)).toBeVisible());
      expect(screen.getByText(deploymentErrors.group)).toBeVisible();
      expect(onScheduleSubmit).not.toHaveBeenCalled();
    });

    it('drops an error once its field is taken care of', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWrapper({});
      await user.click(screen.getByRole('button', { name: /create deployment/i }));
      await waitFor(() => expect(screen.getByText(deploymentErrors.group)).toBeVisible());
      const groupSelect = screen.getByPlaceholderText(/select a device group/i);
      await user.type(groupSelect, 'testGroupDyn');
      await user.keyboard('{ArrowDown}{Enter}');
      await waitFor(() => expect(screen.queryByText(deploymentErrors.group)).toBeFalsy());
      expect(screen.getByText(deploymentErrors.release)).toBeVisible();
    });

    it('drops an error once the option that caused it is switched off again', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWrapper({});
      await user.click(screen.getByRole('button', { name: /advanced options/i }));
      const limitCheckbox = screen.getByRole('checkbox', { name: /maximum number of devices/i });
      await user.click(limitCheckbox);
      await user.click(screen.getByRole('button', { name: /create deployment/i }));
      await waitFor(() => expect(screen.getByText(deploymentErrors.maxDevices)).toBeVisible());
      await user.click(limitCheckbox);
      await waitFor(() => expect(screen.queryByText(deploymentErrors.maxDevices)).toBeFalsy());
    });

    it('applies a schedule even without a rollout pattern', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      // scheduling is plan gated, so the plain preloadedState would leave the schedule selection disabled
      const enterpriseState = {
        ...defaultState,
        app: { ...defaultState.app, features: { ...defaultState.app.features, isEnterprise: true } }
      };
      renderWrapper({ deploymentObject: { group: ALL_DEVICES, release: defaultState.releases.byId.r1 }, preloadedState: enterpriseState });
      await user.click(screen.getByText(/start immediately/i));
      await user.click(await screen.findByRole('option', { name: /schedule the start date/i }));
      await user.click(await screen.findByRole('gridcell', { name: '28' }));
      await user.click(screen.getByRole('button', { name: 'Next' }));
      await user.click(await screen.findByRole('button', { name: /ok/i }));
      // let the picker dialog finish its exit transition to give the drawer back its visibility
      await act(async () => vi.runOnlyPendingTimers());
      const post = vi.spyOn(GeneralApi, 'post');
      await user.click(screen.getByRole('button', { name: /create deployment/i }));
      await waitFor(() =>
        expect(post).toHaveBeenCalledWith(
          '/api/management/v1/deployments/deployments',
          expect.objectContaining({ phases: [{ batch_size: 100, start_ts: expect.stringMatching(/^2019-01-28/) }] })
        )
      );
    });

    it('expands the advanced options to show an error hidden in them', async () => {
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      renderWrapper({ deploymentObject: { group: 'testGroupDynamic', maxDevices: -90 } });
      const accordionToggle = screen.getByRole('button', { name: /advanced options/i });
      expect(accordionToggle).toHaveAttribute('aria-expanded', 'false');
      await user.click(screen.getByRole('button', { name: /create deployment/i }));
      await waitFor(() => expect(accordionToggle).toHaveAttribute('aria-expanded', 'true'));
      expect(screen.getByText(deploymentErrors.maxDevices)).toBeVisible();
    });
  });
});
