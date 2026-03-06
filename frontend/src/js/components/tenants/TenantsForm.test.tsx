// Copyright 2024 Northern.tech AS
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
import { defaultState } from '@/testUtils';
import { render } from '@/testUtils';
import { TIMEOUTS, rolesByName } from '@northern.tech/store/constants';
import * as StoreThunks from '@northern.tech/store/thunks';
import { spTenantLimits, undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, vi } from 'vitest';

import { TenantCreateForm } from './TenantCreateForm';
import { TenantPage } from './TenantPage';

vi.mock('@northern.tech/store/thunks', { spy: true });

describe('TenantsForm', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<TenantCreateForm open={true} onCloseClick={vi.fn} />);
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as expected', { timeout: TIMEOUTS.refreshDefault }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { addTenant: submitTenantSpy, checkEmailExists: checkEmailExistsSpy } = StoreThunks;

    const newChildTenant = { name: 'ChildTenant', email: 'child+123@example.com', password: 'MySecurePassword2025', limit: '2' };
    const preloadedState = {
      ...defaultState,
      organization: {
        ...defaultState.organization,
        organization: {
          ...defaultState.organization.organization,
          device_limits: spTenantLimits
        }
      }
    };
    render(<TenantPage />, { preloadedState });
    expect(screen.queryByText('You are not currently managing any tenants'));
    await user.click(screen.getByRole('button', { name: /create a tenant/i }));
    await user.type(screen.getByPlaceholderText('Name'), newChildTenant.name);
    const emailInput = screen.getByLabelText(/admin user/i);
    await user.type(emailInput, `bad-${newChildTenant.email}`);
    const micro = screen.getByLabelText(/micro device/i);
    await user.click(micro);
    await user.type(screen.getAllByLabelText(/device limit/i)[0], newChildTenant.limit);
    await act(async () => vi.runOnlyPendingTimers());
    const submitButton = screen.getByRole('button', { name: /Create Tenant/i });

    await act(async () => await user.click(submitButton));
    expect(emailInput).toBeVisible();
    await user.clear(emailInput);
    await user.type(emailInput, newChildTenant.email);
    await act(async () => vi.runOnlyPendingTimers());

    await waitFor(() => expect(checkEmailExistsSpy).toHaveBeenCalledWith(newChildTenant.email));
    await act(async () => await user.click(submitButton));

    await waitFor(() =>
      expect(submitTenantSpy).toHaveBeenCalledWith({
        users: [{ email: newChildTenant.email, role: rolesByName.admin }],
        name: newChildTenant.name,
        deviceLimits: {
          micro: '2',
          standard: 0,
          system: 0
        },
        binary_delta: true,
        restrict_sso_to_parent: false,
        sso: false
      })
    );
  });
});
