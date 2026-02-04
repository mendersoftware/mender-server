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
import { act, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { DeviceTypes, SubscriptionPage } from './SubscriptionPage';

vi.mock('@northern.tech/store/thunks', { spy: true });
export const standardDeviceTier: Partial<DeviceTypes> = {
  standard: { id: 'standard', stripeProductName: 'mender_standard', label: 'Standard devices', summaryLabel: 'Standard', tooltipId: 'standardDevice' }
};
export const microDeviceTier: Partial<DeviceTypes> = {
  micro: {
    id: 'micro',
    stripeProductName: 'mender_micro',
    label: 'Micro devices',
    summaryLabel: 'Micro',
    tooltipId: 'microDevice',
    limitConstrains: { os: { div: 100, min: 100, max: 250 }, professional: { div: 100, min: 100, max: 250 } },
    addonsByPlan: {
      os: [],
      professional: []
    }
  }
};
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
      name: 'mender_micro',
      quantity: 500
    },
    {
      addons: [],
      name: 'mender_standard',
      quantity: 250
    }
  ]
};
const { addons: ADDONS, plans: PLANS } = defaultState.organization.products;
const professionalReq300 = { ...professionalReq, products: [{ ...professionalReq.products[0] }, { ...professionalReq.products[1], quantity: 300 }] };
const professionalReq300WithMonitor = {
  ...professionalReq,
  products: [{ ...professionalReq.products[0] }, { ...professionalReq.products[1], quantity: 300, addons: [{ name: ADDONS.monitor.id }] }]
};
describe('Subscription Page component', () => {
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
        organization: { ...defaultState.organization, organization: { ...defaultState.organization.organization, trial: true } },
        app: {
          ...defaultState.app,
          appInitDone: true,
          features: {
            ...defaultState.app.features,
            hasMCUEnabled: true
          }
        }
      }
    });
    expect(screen.getByText('Upgrade your subscription')).toBeInTheDocument();
    const professionalRadioButton = screen.getByRole('radio', { name: new RegExp(PLANS.professional.id, 'i') });
    const monitorAddonCheckbox = screen.getByRole('checkbox', { name: new RegExp(ADDONS.monitor.title, 'i') });
    // Monitor should be disabled for Basic Plan
    expect(monitorAddonCheckbox).toBeDisabled();
    const microCheckbox = screen.getByRole('checkbox', { name: /micro devices/i });
    await user.click(microCheckbox);

    const deviceLimits = screen.getAllByLabelText('Device limit');
    const [microLimit, standardLimit] = deviceLimits;
    expect(standardLimit).toHaveValue(PLANS.os.tierLimitsConstrains.standard.min);
    expect(microLimit).toHaveValue(PLANS.os.tierLimitsConstrains.micro.min);

    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    await waitFor(() => expect(professionalRadioButton).toBeEnabled());

    await user.click(professionalRadioButton);
    expect(professionalRadioButton).toBeChecked();
    await waitFor(() => expect(microLimit).toHaveValue(PLANS.professional.tierLimitsConstrains.micro.min));
    await waitFor(() => expect(standardLimit).toHaveValue(PLANS.professional.tierLimitsConstrains.standard.min));
    // Monitor addon should not be disabled for Professional Plan
    expect(monitorAddonCheckbox).not.toBeDisabled();
    await waitFor(() => expect(getBillingPreview).toHaveBeenCalled());
    expect(getBillingPreview).toHaveBeenCalledWith(professionalReq);
    await act(async () => {
      await user.clear(standardLimit);
      await user.type(standardLimit, '255');
      await user.tab();
    });
    await act(async () => vi.runOnlyPendingTimers());
    await waitFor(() => expect(standardLimit).toHaveValue(300));
    await act(async () => vi.runOnlyPendingTimers());
    expect(getBillingPreview).toHaveBeenCalledWith(professionalReq300);
    await user.click(monitorAddonCheckbox);
    expect(getBillingPreview).toHaveBeenCalledWith(professionalReq300WithMonitor);

    await user.click(screen.getByRole('radio', { name: new RegExp(PLANS.enterprise.id, 'i') }));

    await act(async () => vi.runOnlyPendingTimers());

    const entRequestMessage = screen.getByLabelText('Your message');
    await user.type(entRequestMessage, 'I want mender so bad');
    const addonSelect = screen.getByRole('combobox', { name: /select add-ons \(optional\)/i });
    await user.click(addonSelect);

    const options = within(screen.getByRole('listbox')).getAllByRole('option');

    await user.click(options[1]);
    expect(addonSelect).toHaveTextContent('Mender Monitor');
    await user.keyboard('{Escape}');
    await user.click(screen.getByRole('button', { name: /Submit request/i }));

    expect(requestEnterprise).toHaveBeenCalledWith(enterpriseReq);
  });
});
