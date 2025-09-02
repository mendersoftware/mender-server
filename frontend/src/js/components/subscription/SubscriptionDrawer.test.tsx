// Copyright 2025 Northern.tech AS
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
import { PLANS } from '@northern.tech/store/appSlice/constants';
import { Elements } from '@stripe/react-stripe-js';
import { loadStripe } from '@stripe/stripe-js';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import { SubscriptionDrawer } from './SubscriptionDrawer';

const createBillingProfileReq = {
  billingProfile: {
    address: {
      city: 'Oslo',
      country: 'PL',
      line1: 'Blindernveien',
      postal_code: '5678',
      state: 'Oslo'
    },
    email: 'a@b.com',
    name: 'test'
  }
};
const preloadedState = {
  organization: {
    ...defaultState.organization,
    organization: {
      ...defaultState.organization.organization,
      billing_profile: {
        email: 'ok@ok.ok',
        name: 'Test account',
        address: {
          country: 'NO',
          city: 'test City',
          line1: 'Test address',
          postal_code: '1234',
          state: 'test State'
        },
        shipping: {
          name: 'Test account',
          address: {
            country: 'NO',
            city: 'test City',
            line1: 'Test address',
            postal_code: '1234',
            state: 'test State'
          }
        }
      },
      subscription: {
        id: 'sub_1RqbZwFlFfXikjZVg0q5xIV1',
        period_start: '2025-07-30T15:06:04Z',
        period_end: '2025-08-30T15:06:04Z',
        total: 3200,
        currency: 'usd',
        lines: [
          {
            description: '50 device × Mender Basic (Tier 1 at $0.00 / month)',
            amount: 0,
            currency: 'usd',
            quantity: 50,
            price_id: 'price_1PYktlFlFfXikjZVh2aNHp7i'
          },
          {
            description: 'Mender Basic (Tier 1 at $32.00 / month)',
            amount: 3200,
            currency: 'usd',
            quantity: 0,
            price_id: 'price_1PYktlFlFfXikjZVh2aNHp7i'
          }
        ],
        status: 'active',
        plan: 'os',
        products: [
          {
            name: 'mender_standard',
            quantity: 50
          }
        ]
      }
    },
    card: {
      last4: '1234',
      expiration: { month: 8, year: 1230 },
      brand: 'Visa'
    }
  }
};

describe('Subscription Summary component', () => {
  it('renders correctly', async () => {
    const stripe = loadStripe();

    const { baseElement } = render(
      <Elements stripe={stripe}>
        <SubscriptionDrawer
          organization={defaultState.organization.organization}
          onClose={vi.fn()}
          plan={PLANS.os}
          addons={{ monitor: false, configure: false, troubleshoot: false }}
          isTrial={true}
        />
      </Elements>
    );
    const view = baseElement.lastElementChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(() => vi.runAllTimersAsync());
  });

  it('allows creating billing profile', async () => {
    const stripe = loadStripe();

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const organizationActions = await import('@northern.tech/store/organizationSlice/thunks');
    const createBillingProfile = vi.spyOn(organizationActions, 'createBillingProfile');
    const ui = (
      <Elements stripe={stripe}>
        <SubscriptionDrawer
          organization={defaultState.organization.organization}
          onClose={vi.fn()}
          plan={PLANS.os}
          addons={{ monitor: false, configure: false, troubleshoot: false }}
          isTrial={true}
        />
      </Elements>
    );
    render(ui, {
      preloadedState: {
        ...defaultState,
        organization: { ...defaultState.organization, organization: { ...defaultState.organization.organization, trial: true } }
      }
    });
    await act(() => vi.runAllTimersAsync());
    expect(screen.getByText(/Subscribe to Mender Basic/i)).toBeVisible();
    await act(async () => vi.runOnlyPendingTimers());

    const input = screen.getByLabelText<HTMLInputElement>('Country or region');

    const addressInput = screen.getByRole('textbox', { name: /address line 1/i });
    const stateInput = screen.getByRole('textbox', { name: /state/i });
    const cityInput = screen.getByRole('textbox', { name: /city/i });
    const zipInput = screen.getByRole('textbox', { name: /zip or postal code/i });

    await user.type(addressInput, 'Blindernveien');
    await user.type(stateInput, 'Oslo');
    await user.type(cityInput, 'Oslo');
    await act(async () => await user.type(zipInput, '5678'));

    const countryAutoComplete = screen.getByRole('combobox', { name: /country/i });
    await act(async () => {
      await user.type(countryAutoComplete, 'Polan');
      await user.keyboard('[ArrowUp]');
      await user.keyboard('[Enter]');
    });

    expect(input.value).toEqual('Poland');

    await act(async () => await user.click(screen.getByRole('button', { name: /Save Billing details/i })));

    await waitFor(() => expect(createBillingProfile).toHaveBeenCalledWith(createBillingProfileReq));
  }, 10000);
  const newOrder = { plan: 'professional', products: [{ name: 'mender_standard', quantity: 250, addons: [] }] };
  it('Allows upgrading subscription', async () => {
    const stripe = loadStripe();

    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const organizationActions = await import('@northern.tech/store/organizationSlice/thunks');
    const requestPlanUpgrade = vi.spyOn(organizationActions, 'requestPlanUpgrade');
    const ui = (
      <Elements stripe={stripe}>
        <SubscriptionDrawer
          organization={defaultState.organization.organization}
          onClose={vi.fn()}
          plan={PLANS.professional}
          order={newOrder}
          addons={{ monitor: false, configure: false, troubleshoot: false }}
          isTrial={false}
          currentPlanId={PLANS.os.id}
        />
      </Elements>
    );
    render(ui, { preloadedState });
    await act(() => vi.runAllTimersAsync());

    expect(screen.getByText(/Upgrade your subscription/i)).toBeVisible();
    expect(screen.getByText(/ok@ok.ok/i)).toBeVisible();

    await act(() => vi.runAllTimersAsync());
    await user.click(screen.getByRole('button', { name: /confirm subscription/i }));
    expect(requestPlanUpgrade).toHaveBeenCalledWith(newOrder);

    await act(() => vi.runAllTimersAsync());
  });
});
