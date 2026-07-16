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
import { getSessionInfo } from '@northern.tech/store/auth';
import { TIMEOUTS, yes } from '@northern.tech/store/constants';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import SelfUserManagement from './SelfUserManagement';

describe('SelfUserManagement Component', () => {
  it('renders correctly', async () => {
    const preloadedState = {
      ...defaultState,
      users: {
        ...defaultState.users,
        byId: {
          ...defaultState.users.byId,
          [defaultState.users.currentUser]: {
            ...defaultState.users.byId[defaultState.users.currentUser],
            sso: [{ kind: 'oauth2/google' }]
          }
        },
        currentSession: getSessionInfo()
      }
    };
    const { baseElement } = render(<SelfUserManagement />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(async () => vi.runAllTimers());
  });

  it('works as intended', { timeout: TIMEOUTS.refreshDefault + TIMEOUTS.fiveSeconds }, async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const preloadedState = { ...defaultState, app: { ...defaultState.app, features: { ...defaultState.app.features, isEnterprise: true } } };

    const copyCheck = vi.fn(yes);
    document.execCommand = copyCheck;
    const ui = <SelfUserManagement />;
    const { rerender } = render(ui, { preloadedState });

    await user.click(screen.getByRole('button', { name: /email/i }));
    const input = screen.getByLabelText(/new email address/i);
    await user.clear(input);
    await user.type(input, 'test@test');
    expect(screen.getByText(/enter a valid email address/i)).toBeInTheDocument();
    await user.type(input, '.com');
    await act(async () => vi.runAllTicks());
    expect(screen.queryByText(/enter a valid email address/i)).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /cancel/i }));

    await user.click(screen.getByRole('button', { name: /change password/i }));
    await user.click(screen.getByRole('button', { name: /cancel/i }));
    await user.click(screen.getByRole('heading', { name: /Two Factor authentication/i }));
    await waitFor(() => expect(screen.getByText(/“Scan QR code” to scan the QR code below/i)).toBeInTheDocument());
    await user.type(screen.getByPlaceholderText(/Verification code/i), '1234');
    expect(screen.getByText(/Must be at least 6 characters long/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Verify/i })).toBeDisabled();
    await user.type(screen.getByPlaceholderText(/Verification code/i), '56');
    await waitFor(() => rerender(ui));
    expect(screen.getByRole('button', { name: /Verify/i })).not.toBeDisabled();
    expect(screen.queryByText(/Must be at least 6 characters long/i)).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /Verify/i }));
    await act(async () => vi.runAllTicks());
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.queryByText(/Verifying/)).not.toBeInTheDocument(), { timeout: 5000 });
    await user.click(screen.getByRole('button', { name: /Save/i }));
    await waitFor(() => rerender(ui));
    await act(async () => vi.runAllTicks());
  });

  it('changes the email through editUser on OS installations', async () => {
    const { editUser: editUserSpy } = StoreThunks;
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const preloadedState = {
      ...defaultState,
      app: { ...defaultState.app, features: { ...defaultState.app.features, isEnterprise: false, hasMultitenancy: false } }
    };
    render(<SelfUserManagement />, { preloadedState });

    await user.click(screen.getByRole('button', { name: /change email address/i }));
    const input = screen.getByLabelText(/new email address/i);
    await user.clear(input);
    await user.type(input, 'test@test.com');
    await user.click(screen.getByRole('button', { name: /save changes/i }));
    const dialog = await screen.findByRole('dialog');
    await user.type(within(dialog).getByLabelText(/password/i), 'mysecretpassword');
    await user.click(within(dialog).getByRole('button', { name: /confirm/i }));
    await act(async () => vi.runAllTicks());
    await waitFor(() =>
      expect(editUserSpy).toHaveBeenCalledWith(expect.objectContaining({ email: 'test@test.com', current_password: 'mysecretpassword' }))
    );
  });

  it('initiates a verified email change on enterprise/hosted installations', async () => {
    const { initiateEmailChange: initiateEmailChangeSpy } = StoreThunks;
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const preloadedState = {
      ...defaultState,
      app: { ...defaultState.app, features: { ...defaultState.app.features, isEnterprise: true, hasMultitenancy: true } }
    };
    render(<SelfUserManagement />, { preloadedState });

    await user.click(screen.getByRole('button', { name: /change email address/i }));
    const input = screen.getByLabelText(/new email address/i);
    await user.clear(input);
    await user.type(input, 'updated@example.com');
    await user.click(screen.getByRole('button', { name: /save changes/i }));
    const dialog = await screen.findByRole('dialog');
    await user.type(within(dialog).getByLabelText(/password/i), 'mysecretpassword');
    await user.click(within(dialog).getByRole('button', { name: /confirm/i }));
    await act(async () => vi.runAllTicks());
    await waitFor(() =>
      expect(initiateEmailChangeSpy).toHaveBeenCalledWith({ email: 'updated@example.com', current_password: 'mysecretpassword' })
    );
  });
});
