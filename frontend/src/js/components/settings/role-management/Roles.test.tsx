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

import { ALL_DEVICES, ALL_RELEASES, TIMEOUTS } from '@northern.tech/store/constants';
import { act, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render, selectMaterialUiSelectOption } from '../../../../../tests/setupTests';
import Roles from './RoleManagement';

describe('Roles Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Roles />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it(
    'works as intended',
    async () => {
      const UserActions = await import('@northern.tech/store/usersSlice/thunks');
      const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
      const editRoleSpy = vi.spyOn(UserActions, 'editRole');
      const removeRoleSpy = vi.spyOn(UserActions, 'removeRole');
      const preloadedState = {
        ...defaultState,
        releases: {
          ...defaultState.releases,
          tags: ['foo', 'bar']
        }
      };
      render(<Roles />, { preloadedState });

      const role = screen.getByText(/test description/i).parentElement;
      await user.click(within(role).getByText(/view details/i));
      await waitFor(() => expect(screen.getByText(/edit role/i)).toBeVisible());
      let collapse = screen.getByText(/edit role/i).parentElement.parentElement.parentElement;
      await user.click(screen.getByRole('button', { name: /delete/i }));
      expect(screen.queryByText(/delete the role/i)).toBeInTheDocument();
      const dialog = screen.getByText(/delete role\?/i).parentElement.parentElement;
      const confirmationButton = within(dialog).getByRole('button', { name: /confirm/i });
      expect(confirmationButton).toBeDisabled();
      await user.type(screen.getByLabelText(/test/i, { selector: 'input' }), preloadedState.users.rolesById.test.name);
      expect(confirmationButton).not.toBeDisabled();
      await user.click(within(dialog).getByRole('button', { name: /confirm/i }));
      expect(removeRoleSpy).toHaveBeenCalled();
      await user.click(within(role).getByText(/view details/i));
      await user.type(within(collapse).getByLabelText(/Description/i), 'something');
      const groupSelect = within(collapse).getByText(Object.keys(defaultState.devices.groups.byId)[0]).parentNode;
      await selectMaterialUiSelectOption(groupSelect, ALL_DEVICES, user);
      expect(screen.getByText(/For 'All devices',/)).toBeVisible();

      const releaseSelect = within(collapse).getByText('bar').parentNode;
      await selectMaterialUiSelectOption(releaseSelect, ALL_RELEASES, user);

      const permissionSelect = within(collapse).getByDisplayValue(ALL_DEVICES).parentElement?.parentElement?.parentElement;
      const selectButton = within(within(permissionSelect).getByText(/read/i).parentElement).getByRole('combobox');
      expect(selectButton).not.toBeDisabled();
      // Open the select dropdown
      // Get the dropdown element. We don't use getByRole() because it includes <select>s too.
      await user.click(selectButton);
      const listbox = await within(document.body).findByRole('listbox');
      expect(listbox).toBeTruthy();

      // Click the list item
      let listItem = within(listbox).getByText(/read/i);
      await user.click(listItem);

      const submitButton = screen.getByRole('button', { name: /submit/i, hidden: true });
      expect(submitButton).toBeDisabled();
      listItem = within(listbox).getByText(/deploy/i);
      await user.click(listItem);
      expect(submitButton).not.toBeDisabled();
      await user.click(submitButton);
      await act(async () => vi.runAllTicks());
      expect(editRoleSpy).toHaveBeenCalledWith({
        allowUserManagement: false,
        description: `${defaultState.users.rolesById.test.description}something`,
        name: 'test',
        uiPermissions: {
          auditlog: [],
          groups: [
            { disableEdit: false, item: ALL_DEVICES, notFound: false, uiPermissions: ['deploy'] },
            { disableEdit: false, item: '', notFound: false, uiPermissions: [] }
          ],
          releases: [
            { disableEdit: false, item: ALL_RELEASES, notFound: false, uiPermissions: ['read'] },
            { disableEdit: false, item: '', notFound: false, uiPermissions: [] }
          ],
          tenantManagement: [],
          userManagement: []
        },
        source: { ...defaultState.users.rolesById.test, value: defaultState.users.rolesById.test.name }
      });
      await act(async () => {
        vi.runOnlyPendingTimers();
        vi.runAllTicks();
      });
    },
    TIMEOUTS.refreshDefault
  );
});
