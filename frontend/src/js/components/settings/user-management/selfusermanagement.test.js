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

import { act, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import { getSessionInfo } from '../../../auth';
import { yes } from '../../../constants/appConstants';
import SelfUserManagement from './selfusermanagement';

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
    await act(async () => jest.runAllTimers());
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const preloadedState = { ...defaultState, app: { ...defaultState.app, features: { ...defaultState.app.features, isEnterprise: true } } };

    const copyCheck = jest.fn(yes);
    document.execCommand = copyCheck;
    const ui = <SelfUserManagement />;
    const { rerender } = render(ui, { preloadedState });

    await user.click(screen.getByRole('button', { name: /email/i }));
    const input = screen.getByDisplayValue(defaultState.users.byId.a1.email);
    await user.clear(input);
    await user.type(input, 'test@test');
    expect(screen.getByText(/enter a valid email address/i)).toBeInTheDocument();
    await user.type(input, '.com');
    expect(screen.queryByText(/enter a valid email address/i)).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /cancel/i }));

    await user.click(screen.getByRole('button', { name: /change password/i }));
    const form = screen.getByLabelText('Password *').parentElement.parentElement.parentElement.parentElement;
    const passwordGeneration = within(form).getByRole('button', { name: /generate/i });
    await user.click(passwordGeneration);
    expect(copyCheck).toHaveBeenCalled();
    await user.click(screen.getByRole('button', { name: /cancel/i }));
    await user.click(screen.getByText(/Enable Two Factor authentication/i));
    await act(async () => jest.runAllTicks());
    await waitFor(() => rerender(ui));
    expect(screen.getByText(/Scan the QR code on the right/i)).toBeInTheDocument();
    await user.type(screen.getByPlaceholderText(/Verification code/i), '1234');
    expect(screen.getByText(/Must be at least 6 characters long/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Verify/i })).toBeDisabled();
    await user.type(screen.getByPlaceholderText(/Verification code/i), '56');
    await waitFor(() => rerender(ui));
    expect(screen.getByRole('button', { name: /Verify/i })).not.toBeDisabled();
    expect(screen.queryByText(/Must be at least 6 characters long/i)).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /Verify/i }));
    await act(async () => jest.runAllTicks());
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.queryByText(/Verifying/)).not.toBeInTheDocument(), { timeout: 5000 });
    await user.click(screen.getByRole('button', { name: /Save/i }));
    await waitFor(() => rerender(ui));
    await act(async () => jest.runAllTicks());
  }, 15000);
});
