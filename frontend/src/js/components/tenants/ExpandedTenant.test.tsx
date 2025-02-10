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
import { Tenant } from '@northern.tech/store/api/types/Tenant';
import { initialState as initialOrganizationState } from '@northern.tech/store/organizationSlice';
import * as OrganizationActions from '@northern.tech/store/organizationSlice/thunks';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import { ExpandedTenant } from './ExpandedTenant';

const state = {
  ...defaultState,
  organization: {
    ...defaultState.organization,
    tenantList: {
      ...initialOrganizationState.tenantList,
      tenants: [
        {
          id: '671a0f1dd58c813118fe8622',
          parent_tenant_id: '6718de64b42e08dea2a2065d',
          name: 'child2',
          tenant_token: 'mQDYRCr-tGbDuJhPp7fArbfTA5htVTWE9G204AzhDUM',
          status: 'active',
          additional_info: {
            marketing: false,
            campaign: ''
          },
          plan: 'enterprise',
          trial: false,
          trial_expiration: null,
          service_provider: false,
          created_at: '2024-10-24T09:10:53.281Z',
          cancelled_at: null,
          children_tenants: null,
          max_child_tenants: 0,
          device_count: 0,
          device_limit: 100,
          binary_delta: true
        }
      ]
    },
    organization: {
      ...defaultState.organization.organization,
      device_count: 20,
      device_limit: 200
    }
  }
};
const tenant: Tenant = state.organization.tenantList.tenants[0];

describe('ExpandedTenant', () => {
  it('renders correctly', () => {
    const { baseElement } = render(<ExpandedTenant onCloseClick={vi.fn} tenant={tenant} />, {
      preloadedState: state
    });
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const editDeviceLimit = vi.spyOn(OrganizationActions, 'editTenantDeviceLimit');

    const newLimit = '5';
    render(<ExpandedTenant onCloseClick={vi.fn} tenant={tenant} />, { preloadedState: state });
    expect(screen.queryByText(`Tenant Information for ${tenant.name}`));
    await user.click(screen.getByRole('button', { name: /edit device limit/i }));
    const limitInput = screen.getByTestId('dev-limit-input');
    await user.clear(limitInput);
    await user.type(limitInput, newLimit);
    await user.click(screen.getByRole('button', { name: /save/i }));
    expect(editDeviceLimit).toHaveBeenCalledWith({ newLimit: Number(newLimit), name: tenant.name, id: tenant.id });
  });
});
