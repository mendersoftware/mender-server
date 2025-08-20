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
import { Route, Routes } from 'react-router-dom';

import { render } from '@/testUtils';
import { TIMEOUTS } from '@northern.tech/store/constants';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Cookies from 'universal-cookie';
import { vi } from 'vitest';

import Signup from './Signup';

const cookies = new Cookies();

describe('Signup Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Signup match={{ params: { campaign: '' } }} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(() => vi.runAllTimersAsync());
  });

  it('allows signing up', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const ui = (
      <>
        <Signup />
        <Routes>
          <Route path="/" element={<div>signed up</div>} />
        </Routes>
      </>
    );
    const { container } = render(ui);
    expect(screen.getByText('Sign up with:')).toBeInTheDocument();
    await user.type(screen.getByLabelText(/Email/i), 'test@example.com');
    const passwordInput = screen.getByLabelText('Password *');
    const passwordConfirmationInput = screen.getByLabelText(/confirm password/i);
    await user.type(passwordInput, 'mysecretpassword!123');
    expect(screen.getByRole('button', { name: /sign up/i })).toBeDisabled();
    await user.type(passwordConfirmationInput, 'mysecretpassword!123');
    expect(container.querySelector('#pass-strength > meter')).toBeVisible();
    await act(async () => {
      vi.runAllTicks();
      vi.runAllTimers();
    });
    await waitFor(() => expect(screen.getByRole('button', { name: /sign up/i })).toBeEnabled());
    await user.click(screen.getByRole('button', { name: /sign up/i }));
    await act(async () => vi.runAllTicks());
    await waitFor(() => screen.getByLabelText(/company or organization name \*/i));
    await user.type(screen.getByLabelText(/company or organization name \*/i), 'test');
    expect(screen.getByRole('button', { name: /complete signup/i })).toBeDisabled();
    await user.click(screen.getByRole('checkbox', { name: /by checking this you agree to our/i }));
    await waitFor(() => expect(screen.getByRole('button', { name: /complete signup/i })).toBeEnabled());
    const cookiesSet = vi.spyOn(cookies, 'set');
    await user.click(screen.getByRole('button', { name: /complete signup/i }));
    await waitFor(() => expect(container.querySelector('.loaderContainer')).toBeVisible());
    await act(async () => vi.advanceTimersByTime(TIMEOUTS.refreshDefault));
    await waitFor(() =>
      expect(cookiesSet).toHaveBeenCalledWith('firstLoginAfterSignup', true, { domain: '.mender.io', maxAge: 60, path: '/', sameSite: false })
    );
  }, 10000);
});
