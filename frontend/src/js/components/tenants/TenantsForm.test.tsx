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
import Api from '@northern.tech/store/api/general-api';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect } from 'vitest';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import { TenantCreateForm } from './TenantCreateForm';
import { TenantPage } from './TenantPage';

describe('TenantsForm', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<TenantCreateForm open={true} onCloseClick={vi.fn} />);
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as expected', async () => {
    const OrganizationActions = await import('@northern.tech/store/organizationSlice/thunks');
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const submitTenantSpy = vi.spyOn(OrganizationActions, 'addTenant');
    const tenantExists = vi.spyOn(Api, 'get');
    tenantExists.mockResolvedValue({ exists: false });

    const newChildTenant = { name: 'ChildTenant', email: 'child@example.com', password: 'MySecurePassword2025', dev: '2' };
    const preloadedState = {
      ...defaultState,
      organization: {
        ...defaultState.organization,
        organization: {
          ...defaultState.organization.organization,
          device_limit: 200
        }
      }
    };

    render(<TenantPage />, { preloadedState });

    expect(screen.queryByText('You are not currently managing any tenants'));
    await user.click(screen.getByRole('button', { name: /Add tenant/i }));
    await user.type(screen.getByPlaceholderText('Name'), newChildTenant.name);
    await user.type(screen.getByRole('textbox', { name: /admin user/i }), newChildTenant.email);
    await user.type(screen.getByPlaceholderText('Password'), newChildTenant.password);
    await user.type(screen.getByLabelText('Set device limit'), newChildTenant.dev);
    await user.click(screen.getByText(/enable delta artifact generation/i));
    await user.click(screen.getByText(/reset the password/i));
    const submitButton = screen.getByRole('button', { name: /Create Tenant/i });
    await waitFor(() => expect(submitButton).toBeEnabled());
    await user.click(submitButton);

    await waitFor(() =>
      expect(submitTenantSpy).toHaveBeenCalledWith({
        admin: { email: newChildTenant.email, password: newChildTenant.password, send_reset_password: true },
        name: newChildTenant.name,
        device_limit: Number(newChildTenant.dev),
        binary_delta: true,
        sso: false
      })
    );
  });
});
