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
import { Route, Routes } from 'react-router-dom';

import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Cookies from 'universal-cookie';

import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import Signup from './signup';

describe('Signup Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Signup match={{ params: { campaign: '' } }} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('allows signing up', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const ui = (
      <>
        <Signup location={{ state: { from: '' } }} match={{ params: {} }} />
        <Routes>
          <Route path="/" element={<div>signed up</div>} />
        </Routes>
      </>
    );
    const { container, rerender } = render(ui);
    expect(screen.getByText('Sign up with:')).toBeInTheDocument();
    await user.type(screen.getByLabelText(/Email/i), 'test@example.com');
    await user.type(screen.getByLabelText('Password *'), 'mysecretpassword!123');
    expect(screen.getByRole('button', { name: /sign up/i })).toBeDisabled();
    await user.type(screen.getByLabelText(/Confirm password/i), 'mysecretpassword!123');
    expect(container.querySelector('#pass-strength > meter')).toBeVisible();
    await act(async () => jest.runOnlyPendingTimers());
    expect(screen.getByRole('button', { name: /sign up/i })).toBeEnabled();
    await user.click(screen.getByRole('button', { name: /sign up/i }));
    await waitFor(() => screen.queryByText('Company or organization name *'));
    await user.type(screen.getByRole('textbox', { name: /company or organization name \*/i }), 'test');
    expect(screen.getByRole('button', { name: /complete signup/i })).toBeDisabled();
    await user.click(screen.getByRole('checkbox', { name: /by checking this you agree to our/i }));
    await act(async () => jest.runOnlyPendingTimers());
    expect(screen.getByRole('button', { name: /complete signup/i })).toBeEnabled();

    const cookies = new Cookies();
    cookies.set.mockReturnValue();
    await user.click(screen.getByRole('button', { name: /complete signup/i }));
    await waitFor(() => expect(container.querySelector('.loaderContainer')).toBeVisible());
    await act(async () => jest.advanceTimersByTime(5000));
    await waitFor(() => rerender(ui));
    await waitFor(() =>
      expect(cookies.set).toHaveBeenLastCalledWith('firstLoginAfterSignup', true, { domain: '.mender.io', maxAge: 60, path: '/', sameSite: false })
    );
  }, 10000);
});
