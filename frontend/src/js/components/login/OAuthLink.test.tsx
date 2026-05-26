// Copyright 2026 Northern.tech AS
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
import { render } from '@/testUtils';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Cookies from 'universal-cookie';
import { vi } from 'vitest';

import { OAuthLink } from './OAuthLink';

const correctPassword = 'mysecretpassword!123';
const defaultEmail = 'someone@example.com';
const twoFaEmail = 'something-2fa@example.com';

describe('OAuthLink Component', () => {
  let cookies: Cookies;
  beforeEach(() => {
    cookies = new Cookies();
    cookies.set('oauth', 'google', { path: '/' });
    cookies.set('email', defaultEmail, { path: '/' });
  });
  afterEach(() => {
    cookies.remove('oauth', { path: '/' });
    cookies.remove('email', { path: '/' });
    vi.mocked(StoreThunks.confirmOAuthLink).mockClear();
  });

  it('renders correctly', async () => {
    const { baseElement } = render(<OAuthLink />);
    expect(await screen.findByText(/Link Google account\?/i)).toBeInTheDocument();
    expect(baseElement.firstChild).toEqual(expect.not.stringMatching(undefineds));
  });

  it('submits email + password and omits token2fa on the first verify call', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<OAuthLink />);
    await user.click(await screen.findByRole('button', { name: /Link my accounts/i }));
    await user.type(screen.getByLabelText(/Password/i), correctPassword);
    await user.click(screen.getByRole('button', { name: /Confirm/i }));
    await act(async () => vi.runAllTicks());
    await waitFor(() =>
      expect(StoreThunks.confirmOAuthLink).toHaveBeenCalledWith({
        email: defaultEmail,
        password: correctPassword,
        stayLoggedIn: true
      })
    );
  });

  it('sends email, password and token2fa together on the 2fa retry', async () => {
    cookies.set('email', twoFaEmail, { path: '/' });
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<OAuthLink />);
    await user.click(await screen.findByRole('button', { name: /Link my accounts/i }));
    await user.type(screen.getByLabelText(/Password/i), correctPassword);
    await user.click(screen.getByRole('button', { name: /Confirm/i }));
    const codeInput = await screen.findByLabelText(/Verification code/i);
    await user.type(codeInput, '123456');
    await user.click(screen.getByRole('button', { name: /Confirm/i }));
    await act(async () => vi.runAllTicks());
    await waitFor(() =>
      expect(StoreThunks.confirmOAuthLink).toHaveBeenLastCalledWith({
        email: twoFaEmail,
        password: correctPassword,
        stayLoggedIn: true,
        token2fa: '123456'
      })
    );
  });

  it('shows the error message in an alert when the failure is not a 2fa challenge', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<OAuthLink />);
    await user.click(await screen.findByRole('button', { name: /Link my accounts/i }));
    await user.type(screen.getByLabelText(/Password/i), 'wrong');
    await user.click(screen.getByRole('button', { name: /Confirm/i }));
    expect(await screen.findByText(/There was a problem logging in/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Password/i)).toBeInTheDocument();
    expect(screen.queryByLabelText(/Verification code/i)).not.toBeInTheDocument();
  });
});
