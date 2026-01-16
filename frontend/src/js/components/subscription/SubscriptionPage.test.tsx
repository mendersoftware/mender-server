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
import { ADDONS, PLANS, TIMEOUTS } from '@northern.tech/store/constants';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { SubscriptionPage } from './SubscriptionPage';

vi.mock('@northern.tech/store/thunks', { spy: true });

const enterpriseReq = {
  content: {
    current_addons: '-',
    current_plan: 'Basic',
    requested_addons: 'monitor',
    requested_plan: 'Enterprise',
    user_message: 'I want mender so bad'
  },
  tenantId: '1'
};

const professionalReq = {
  plan: 'professional',
  preview_mode: 'recurring',
  products: [
    {
      addons: [],
      name: 'mender_standard',
      quantity: 250
    }
  ]
};
const professionalReq300 = { ...professionalReq, products: [{ ...professionalReq.products[0], quantity: 300 }] };
const professionalReq300WithMonitor = {
  ...professionalReq,
  products: [{ ...professionalReq.products[0], quantity: 300, addons: [{ name: ADDONS.monitor.id }] }]
};
describe('Subscription Summary component', () => {
  it('renders correctly', async () => {
    //
    const { baseElement } = render(<SubscriptionPage />, {
      preloadedState: {
        ...defaultState,
        organization: { ...defaultState.organization, organization: { ...defaultState.organization.organization, id: '6863115e67294908fbbd6dd0' } }
      }
    });
    const view = baseElement.lastElementChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('allows signing up', { timeout: 3 * TIMEOUTS.fiveSeconds }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { getBillingPreview, requestPlanChange: requestEnterprise } = StoreThunks;
    const ui = <SubscriptionPage />;
    render(ui, {
      preloadedState: {
        ...defaultState,
        devices: { ...defaultState.devices, limits: { ...defaultState.devices.limits, standard: 10 } },
        organization: { ...defaultState.organization, organization: { ...defaultState.organization.organization, trial: true } }
      }
    });
    expect(screen.getByText('Upgrade your subscription')).toBeInTheDocument();
    const professionalRadioButton = screen.getByRole('radio', { name: new RegExp(PLANS.professional.name, 'i') });
    const monitorAddonCheckbox = screen.getByRole('checkbox', { name: new RegExp(ADDONS.monitor.title, 'i') });

    // Monitor should be disabled for Basic Plan
    expect(monitorAddonCheckbox).toBeDisabled();
    const deviceLimit = screen.getByLabelText('Number of devices');
    expect(deviceLimit).toHaveValue(PLANS.os.minimalDeviceCount);
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    await waitFor(() => expect(professionalRadioButton).toBeEnabled());

    await user.click(professionalRadioButton);
    expect(professionalRadioButton).toBeChecked();

    await waitFor(() => expect(deviceLimit).toHaveValue(PLANS.professional.minimalDeviceCount));
    // Monitor addon should not be disabled for Professional Plan
    expect(monitorAddonCheckbox).not.toBeDisabled();
    await waitFor(() => expect(getBillingPreview).toHaveBeenCalled());
    expect(getBillingPreview).toHaveBeenCalledWith(professionalReq);
    await act(async () => {
      await user.clear(deviceLimit);
      await user.type(deviceLimit, '255');
      await user.tab();
    });
    await act(async () => vi.runOnlyPendingTimers());
    await waitFor(() => expect(deviceLimit).toHaveValue(300));
    await act(async () => vi.runOnlyPendingTimers());
    expect(getBillingPreview).toHaveBeenCalledWith(professionalReq300);
    await user.click(monitorAddonCheckbox);
    expect(getBillingPreview).toHaveBeenCalledWith(professionalReq300WithMonitor);

    await user.click(screen.getByRole('radio', { name: new RegExp(PLANS.enterprise.id, 'i') }));

    await act(async () => vi.runOnlyPendingTimers());

    const entRequestMessage = screen.getByLabelText('Your message');
    await user.type(entRequestMessage, 'I want mender so bad');
    await user.click(screen.getByRole('button', { name: /Submit request/i }));

    expect(requestEnterprise).toHaveBeenCalledWith(enterpriseReq);
  });
});
