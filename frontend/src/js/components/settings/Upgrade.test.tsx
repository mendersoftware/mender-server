// Copyright 2019 Northern.tech AS
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

import { getSessionInfo } from '@northern.tech/store/auth';
import { actions } from '@northern.tech/store/organizationSlice/index';
import { Elements } from '@stripe/react-stripe-js';
import { loadStripe } from '@stripe/stripe-js';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import Upgrade, { PricingContactNote } from './Upgrade';

const changeRequestBase = {
  content: {
    current_addons: '-',
    current_plan: 'Basic',
    requested_addons: '-',
    requested_plan: 'Basic',
    user_message: ''
  },
  tenantId: 1
};
describe('smaller components', () => {
  [PricingContactNote].forEach(Component => {
    it(`renders ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(
        <Component
          trial_expiration="2019-10-05T13:00:00.000Z"
          isTrial={true}
          handleCancelSubscription={vi.fn}
          orgName="test"
          mailBodyTexts={{ billing: 'bill this', upgrade: 'upgrade here' }}
        />
      );
      const view = baseElement.firstChild.firstChild;
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
    });
  });
});

describe('Upgrade Component', () => {
  it('renders correctly', async () => {
    window.localStorage.getItem.mockImplementation(() => null);
    const stripe = loadStripe();
    const { baseElement } = render(
      <Elements stripe={stripe}>
        <Upgrade />
      </Elements>,
      {
        preloadedState: {
          ...defaultState,
          app: { ...defaultState.app, features: { ...defaultState.app.features, hasDeviceConfig: true, hasDeviceConnect: true } },
          users: { ...defaultState.users, currentSession: getSessionInfo() }
        }
      }
    );
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    window.localStorage.getItem.mockReset();
  });
  const professionalRequestArgs = {
    ...changeRequestBase,
    content: {
      ...changeRequestBase.content,
      requested_plan: 'Professional'
    }
  };
  const trialState = {
    preloadedState: {
      ...defaultState,
      organization: {
        ...defaultState.organization,
        organization: { ...defaultState.organization.organization, trial: true, plan: 'enterprise' }
      }
    }
  };
  it('signup works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<Upgrade />, trialState);
    vi.spyOn(actions, 'setOrganization').mockImplementation(() => trialState);

    const upgradeButton = await screen.getAllByRole('button', { name: /subscribe/i })[0];
    await user.click(upgradeButton);
    await user.type(await screen.getByRole('textbox', { name: /address line 1/i }), 'Blindernveien');
    await user.type(await screen.getByRole('textbox', { name: /state/i }), 'Oslo');
    await user.type(await screen.getByRole('textbox', { name: /city/i }), 'Oslo');
    await user.type(await screen.getByRole('textbox', { name: /zip or postal code/i }), '1234');
    const countryAutoComplete = await screen.getByRole('combobox', { name: /country/i });
    const input = await screen.getByLabelText('Country');
    await user.type(countryAutoComplete, 'Norw');
    await user.keyboard('[ArrowUp]');
    await user.keyboard('[Enter]');
    expect(input.value).toEqual('Norway');
  });
  it('upgrade works as intended', async () => {
    const OrganizationActions = await import('@northern.tech/store/organizationSlice/thunks');
    const professionalRequest = vi.spyOn(OrganizationActions, 'requestPlanChange');

    const storageMock = vi.spyOn(Storage.prototype, 'setItem');
    Storage.prototype.setItem = vi.fn();

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { rerender } = render(<Upgrade />, { preloadedState: defaultState });
    const upgradeButton = await screen.findByRole('button', { name: /upgrade/i });
    await user.click(upgradeButton);

    let confirmButton = await screen.findByRole('button', { name: /confirm/i });
    await user.click(confirmButton);

    await waitFor(() => {
      expect(localStorage.setItem).toBeCalledTimes(1);
    });
    let localStorageState = localStorage.setItem.mock.calls[0][1];
    window.localStorage.getItem.mockImplementation(() => localStorageState);
    rerender(<Upgrade />);
    await screen.getByRole('button', { name: /pending/i });
    expect(professionalRequest).toHaveBeenCalledWith(professionalRequestArgs);
    localStorage.setItem.mockClear();
    storageMock.mockClear();
  });

  it('adding addon works as intended', async () => {
    const OrganizationActions = await import('@northern.tech/store/organizationSlice/thunks');
    const addonRequest = vi.spyOn(OrganizationActions, 'requestPlanChange');

    const storageMock = vi.spyOn(Storage.prototype, 'setItem');
    Storage.prototype.setItem = vi.fn();
    window.localStorage.getItem.mockImplementation(() => null);

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });

    render(<Upgrade />, { preloadedState: defaultState });
    const addToPlanButton = await screen.getAllByRole('button', { name: /add to plan/i });
    await user.click(addToPlanButton[0]);

    const confirmButton = await screen.findByRole('button', { name: /confirm/i });
    await user.click(confirmButton);

    await user.click(addToPlanButton[2]);
    await user.click(await screen.findByRole('button', { name: /confirm/i }));

    await waitFor(() => {
      expect(localStorage.setItem).toBeCalledTimes(2);
    });
    const localStorageState = localStorage.setItem.mock.calls[0][1];
    const currentState = { configure: { pending: true, isAdd: true } };
    expect(addonRequest).toHaveBeenCalledWith({ ...changeRequestBase, content: { ...changeRequestBase.content, requested_addons: 'configure' } });

    const localStorageExpectedState = JSON.stringify(currentState);
    expect(localStorageState).toEqual(localStorageExpectedState);
    localStorage.setItem.mockClear();
    storageMock.mockClear();
  });
  const enterpriseRequestArgs = {
    ...changeRequestBase,
    content: {
      ...changeRequestBase.content,
      requested_plan: 'Enterprise',
      user_message: 'Interested in updating to Enterprise version'
    }
  };
  it('enterprise request works as intended', async () => {
    const OrganizationActions = await import('@northern.tech/store/organizationSlice/thunks');
    const enterpriseRequest = vi.spyOn(OrganizationActions, 'requestPlanChange');
    window.localStorage.getItem.mockImplementation(() => null);
    vi.spyOn(Storage.prototype, 'setItem');
    Storage.prototype.setItem = vi.fn();

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<Upgrade />, { preloadedState: defaultState });

    const contactButton = await screen.findByRole('button', { name: /contact us/i });
    await user.click(contactButton);
    const messageInput = await screen.getByRole('textbox');
    await user.type(messageInput, enterpriseRequestArgs.content.user_message);
    const addOnCombobox = await screen.getByRole('combobox');
    await user.click(addOnCombobox);

    const submitButton = await screen.getByRole('button', { name: /submit request/i, hidden: true });
    await user.click(submitButton);
    await waitFor(() => {
      expect(localStorage.setItem).toBeCalledTimes(1);
    });
    expect(enterpriseRequest).toHaveBeenCalledWith(enterpriseRequestArgs);
    const localStorageState = localStorage.setItem.mock.calls[0][1];
    expect(localStorageState).toEqual(JSON.stringify({ enterprise: { pending: true, isAdd: true } }));
    localStorage.setItem.mockReset();
  });
});
