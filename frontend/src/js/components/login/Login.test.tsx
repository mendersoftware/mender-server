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
import { defaultState } from '@/testUtils';
import { render } from '@/testUtils';
import { TIMEOUTS } from '@northern.tech/store/constants';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import Login from './Login';

vi.mock('@northern.tech/store/thunks', { spy: true });

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: Object.freeze({
      ...defaultState.app.features,
      isHosted: true
    })
  }
};

describe('Login Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Login />, { preloadedState });
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', { timeout: 2 * TIMEOUTS.fiveSeconds }, async () => {
    window.localStorage.getItem.mockImplementation(() => null);
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { loginUser: loginSpy } = StoreThunks;
    const ui = <Login />;
    const { rerender } = render(ui, { preloadedState });
    await user.type(screen.getByLabelText(/your email/i), 'something-2fa@example.com');
    const loginButton = screen.getByRole('button', { name: /Next/i });
    await user.click(loginButton);
    await act(async () => vi.runAllTicks());
    await waitFor(() => expect(loginSpy).toHaveBeenCalled());
    await user.type(screen.getByLabelText(/password/i), 'mysecretpassword!123');
    expect(await screen.findByLabelText(/Two Factor Authentication Code/i)).not.toBeVisible();
    await waitFor(() => expect(loginButton).toBeEnabled());
    loginSpy.mockClear();
    await user.click(loginButton);
    await waitFor(() => expect(loginSpy).toHaveBeenCalled());
    await waitFor(() => expect(screen.getByLabelText(/Two Factor Authentication Code/i)).toBeVisible());
    const input = screen.getByDisplayValue('something-2fa@example.com');
    expect(input).toBeDisabled();
    await user.click(screen.getByTestId('EditIcon'));
    expect(input).not.toBeDisabled();
    await user.clear(input);
    await user.type(input, 'something@example.com');
    await user.type(screen.getByLabelText(/Two Factor Authentication Code/i), '123456');
    await waitFor(() => rerender(ui));
    await act(async () => vi.runAllTicks());
    await waitFor(() => expect(loginButton).toBeEnabled());
    await user.click(loginButton);
    await act(async () => vi.runAllTicks());
    expect(loginSpy).toHaveBeenCalledWith({ email: 'something@example.com', password: 'mysecretpassword!123', token2fa: '123456', stayLoggedIn: false });
    window.localStorage.getItem.mockReset();
  });
});
