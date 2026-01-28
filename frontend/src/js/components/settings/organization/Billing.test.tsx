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
import { defaultState, render } from '@/testUtils';
import { TIMEOUTS } from '@northern.tech/store/constants';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { Billing } from './Billing';

vi.mock('@northern.tech/store/thunks', { spy: true });

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
        },
        shipping: {
          name: 'Test account',
          address: {
            country: 'NO',
            city: 'test City',
            line1: 'Test address',
            postal_code: '1234',
          }
        }
      }
    },
    card: {
      last4: '1234',
      expiration: { month: 8, year: 1230 },
      brand: 'Visa'
    }
  }
};

const editProfileActionParams = {
  billingProfile: {
    address: {
      city: 'Oslo',
      country: 'PL',
      line1: 'Blindernveien',
      postal_code: '5678',
    },
    email: 'ok@ok.ok',
    name: 'Test account'
  }
};

describe('Billing Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Billing />, {});
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('supports modifying billing profile', { timeout: TIMEOUTS.refreshDefault }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { editBillingProfile: editProfileAction } = StoreThunks;
    const ui = <Billing />;
    render(ui, { preloadedState });
    expect(screen.getByText(/1234, test city/i)).toBeVisible();
    await act(async () => await user.click(screen.getByRole('button', { name: /edit/i })));
    const input = screen.getByLabelText<HTMLInputElement>('Country or region');

    const addressInput = screen.getByRole('textbox', { name: /address line 1/i });
    const cityInput = screen.getByRole('textbox', { name: /city/i });
    const zipInput = screen.getByRole('textbox', { name: /zip or postal code/i });
    await user.clear(addressInput);
    await user.clear(cityInput);
    await user.clear(zipInput);

    await user.type(addressInput, 'Blindernveien');
    await user.type(cityInput, 'Oslo');
    await act(async () => await user.type(zipInput, '5678'));
    const countryAutoComplete = screen.getByRole('combobox', { name: /country/i });
    await act(async () => {
      await user.type(countryAutoComplete, 'Polan');
      await user.keyboard('[ArrowUp]');
      await user.keyboard('[Enter]');
      await user.click(screen.getByRole('button', { name: /save/i }));
    });
    expect(input.value).toEqual('Poland');

    await waitFor(() => expect(editProfileAction).toHaveBeenCalledWith(editProfileActionParams));
  });
});
