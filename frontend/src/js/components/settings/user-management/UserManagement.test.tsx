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
import { defaultState, render } from '@/testUtils';
import { TIMEOUTS, yes } from '@northern.tech/store/constants';
import { undefineds, userId } from '@northern.tech/testing/mockData';
import { act, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { UserManagement } from './UserManagement';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      isEnterprise: true
    }
  }
};

const dropDownSelector = 'ul[role=listbox]';

describe('UserManagement Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<UserManagement />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', { timeout: 2 * TIMEOUTS.fiveSeconds }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const copyCheck = vi.fn(yes);
    document.execCommand = copyCheck;
    render(<UserManagement />, { preloadedState });

    expect(screen.queryByText(/delete the user with email/i)).not.toBeInTheDocument();
    const list = screen.getAllByText(/view details/i);
    await user.click(list[list.length - 1]);
    await user.click(screen.getByRole('button', { name: /delete user/i }));
    expect(screen.queryByText(/delete the user with email/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /cancel/i }));
    await user.click(list[list.length - 1]);
    const input = screen.getByDisplayValue(defaultState.users.byId[userId].email);
    await user.clear(input);
    await user.type(input, 'test@test');
    expect(screen.getByText(/enter a valid email address/i)).toBeInTheDocument();
    await user.type(input, '.com');
    expect(screen.queryByText(/enter a valid email address/i)).not.toBeInTheDocument();
    await user.click(screen.getByRole('checkbox', { name: /reset the password/i }));
    await user.click(screen.getByRole('checkbox', { name: /reset the password/i }));
    const selectButton = screen.getByLabelText(/roles/i).parentNode.querySelector('[role=combobox]');
    await user.click(selectButton);
    let listbox = document.body.querySelector(dropDownSelector);
    const adminItem = within(listbox).getByText(/admin/i);
    await user.click(adminItem);
    await user.type(listbox, '{Escape}');
    expect(screen.getByRole('button', { name: /Save/i })).toBeDisabled();
    await user.click(selectButton);
    listbox = document.body.querySelector(dropDownSelector);
    const listItem = within(listbox).getByText(/read access/i);
    await user.click(listItem);
    await user.type(listbox, '{Escape}');
    await user.click(screen.getByRole('button', { name: /Save/i }));
  });
  it('supports user creation', async () => {
    const UserActions = await import('@northern.tech/store/usersSlice/thunks');
    const createUserSpy = vi.spyOn(UserActions, 'createUser');
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const copyCheck = vi.fn(yes);
    document.execCommand = copyCheck;
    const ui = <UserManagement />;
    const { rerender } = render(ui, { preloadedState });
    expect(screen.queryByText(/send an email/i)).not.toBeInTheDocument();
    const userCreationButton = screen.getByRole('button', { name: /add new user/i });
    await user.click(userCreationButton);
    expect(screen.queryByText(/send an email/i)).toBeInTheDocument();
    const submitButton = screen.getByRole('button', { name: /create user/i });
    expect(submitButton).toBeDisabled();
    const input = screen.getByPlaceholderText(/email/i);
    await user.type(input, 'test@test');
    expect(screen.getByText(/enter a valid email address/i)).toBeInTheDocument();
    await user.type(input, '.com');
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(submitButton).toBeEnabled());
    await user.click(screen.getByRole('button', { name: /generate/i }));
    expect(copyCheck).toHaveBeenCalled();
    expect(submitButton).toBeEnabled();
    const passwordInput = screen.getByPlaceholderText(/password/i);
    await user.clear(passwordInput);
    expect(submitButton).toBeEnabled();
    await user.click(submitButton);
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    await waitFor(() => expect(createUserSpy).toHaveBeenCalled(), { timeout: 3000 });
  });

  it('allows role adjustments', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<UserManagement />, { preloadedState });
    const list = screen.getAllByText(/view details/i);
    await user.click(list[list.length - 1]);
    const selectButton = screen.getByLabelText(/roles/i).parentNode.querySelector('[role=combobox]');
    await user.click(selectButton);
    let listbox = document.body.querySelector(dropDownSelector);
    const adminItem = within(listbox).getByText(/admin/i);
    await user.click(adminItem);
    await user.type(listbox, '{Escape}');
    expect(screen.getByRole('button', { name: /Save/i })).toBeDisabled();
    await user.click(selectButton);
    listbox = document.body.querySelector(dropDownSelector);
    const listItem = within(listbox).getByText(/releases/i);
    await user.click(listItem);
    await user.click(screen.getByDisplayValue(defaultState.users.byId[userId].email));
    expect(screen.getByText(/the selected role may prevent/i)).toBeInTheDocument();
    await user.type(listbox, '{Escape}');
    expect(screen.getByRole('button', { name: /Save/i })).not.toBeDisabled();
  });
});
