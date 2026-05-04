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
import { defaultState, render } from '@/testUtils';
import { getSessionInfo } from '@northern.tech/store/auth';
import { initialState as initialOrganizationState } from '@northern.tech/store/organizationSlice';
import { getTenantListWithLimits } from '@northern.tech/store/selectors';
import * as StoreThunks from '@northern.tech/store/thunks';
import { spTenantLimits, tenants, undefineds } from '@northern.tech/testing/mockData';
import { act, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { ExpandedTenant } from './ExpandedTenant';

vi.mock('@northern.tech/store/thunks', { spy: true });

const state = {
  ...defaultState,
  organization: {
    ...defaultState.organization,
    tenantList: {
      ...initialOrganizationState.tenantList,
      tenants
    },
    organization: {
      ...defaultState.organization.organization,
      device_count: 20,
      device_limit: 200,
      device_limits: spTenantLimits
    }
  }
};
const tenant = getTenantListWithLimits(state).tenants[0];

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
    const { editTenant: editDeviceLimit } = StoreThunks;

    const newLimits = ['50', '20', '0'];
    render(<ExpandedTenant onCloseClick={vi.fn} tenant={tenant} />, {
      preloadedState: { ...state, users: { ...defaultState.users, currentSession: getSessionInfo() } }
    });
    expect(screen.queryByText(`Tenant Information for ${tenant.name}`));
    await user.click(screen.getByRole('button', { name: /Manage device limits/i }));
    const limitInputs = screen.getAllByLabelText(/device limit/i);
    for (const input of limitInputs) {
      const i = limitInputs.indexOf(input);
      await user.clear(input);
      await user.type(input, newLimits[i]);
    }
    await user.click(screen.getByRole('button', { name: /save/i }));
    expect(editDeviceLimit).toHaveBeenCalledWith({
      deviceLimits: {
        micro: 50,
        standard: 20,
        system: 0
      },
      name: tenant.name,
      id: tenant.id
    });
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
  });
});
