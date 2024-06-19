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

import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import * as UserActions from '../../actions/userActions';
import Login from './login';

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

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const loginSpy = jest.spyOn(UserActions, 'loginUser');
    const ui = <Login />;
    const { rerender } = render(ui, { preloadedState });
    await user.type(screen.getByLabelText(/your email/i), 'something-2fa@example.com');
    await waitFor(() => rerender(ui));
    await user.click(screen.getByRole('button', { name: /Log in/i }));
    expect(loginSpy).toHaveBeenCalled();
    await waitFor(() => rerender(ui));
    await user.type(screen.getByLabelText(/password/i), 'mysecretpassword!123');
    expect(await screen.findByLabelText(/Two Factor Authentication Code/i)).not.toBeVisible();
    await user.click(screen.getByRole('button', { name: /Log in/i }));
    expect(loginSpy).toHaveBeenCalled();
    await waitFor(() => rerender(ui));
    await act(async () => jest.runAllTicks());
    expect(await screen.findByLabelText(/Two Factor Authentication Code/i)).toBeVisible();
    const input = screen.getByDisplayValue('something-2fa@example.com');
    await user.clear(input);
    await user.type(input, 'something@example.com');
    await user.type(screen.getByLabelText(/Two Factor Authentication Code/i), '123456');
    await waitFor(() => rerender(ui));
    await user.click(screen.getByRole('button', { name: /Log in/i }));
    await act(async () => jest.runAllTicks());
    expect(loginSpy).toHaveBeenCalledWith({ email: 'something@example.com', password: 'mysecretpassword!123', token2fa: '123456' }, false);
  }, 10000);
});
