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
import { Provider } from 'react-redux';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import { act, screen, render as testingLibRender, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import * as UserActions from '../../actions/userActions';
import { getConfiguredStore } from '../../reducers';
import Password from './password';
import PasswordReset from './passwordreset';

const goodPassword = 'mysecretpassword!123';
const badPassword = 'mysecretpassword!546';

describe('PasswordReset Component', () => {
  let store;
  beforeEach(() => {
    store = getConfiguredStore();
  });

  it('renders correctly', async () => {
    const { baseElement } = render(<PasswordReset match={{ params: { secretHash: '' } }} />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const completeSpy = jest.spyOn(UserActions, 'passwordResetComplete');

    const secretHash = 'leHash';

    const ui = (
      <Provider store={store}>
        <MemoryRouter initialEntries={[`/password/${secretHash}`]}>
          <Routes>
            <Route path="password" element={<Password />} />
            <Route path="password/:secretHash" element={<PasswordReset />} />
          </Routes>
        </MemoryRouter>
      </Provider>
    );
    const { rerender } = testingLibRender(ui);

    const passwordInput = screen.getByLabelText('Password *');
    await user.type(passwordInput, badPassword);
    await waitFor(() => rerender(ui));
    await user.type(passwordInput, badPassword);
    await user.type(screen.getByLabelText(/confirm password \*/i), goodPassword);
    await waitFor(() => rerender(ui));
    expect(screen.getByRole('button', { name: /Save password/i })).toBeDisabled();
    expect(screen.getByText('The passwords you provided do not match, please check again.')).toBeVisible();
    await user.clear(passwordInput);
    await user.type(passwordInput, goodPassword);
    await act(async () => jest.runOnlyPendingTimers());
    await user.click(screen.getByRole('button', { name: /Save password/i }));
    await waitFor(() => expect(completeSpy).toHaveBeenCalledWith(secretHash, goodPassword));
    await act(async () => {
      jest.runAllTimers();
      jest.runAllTicks();
      return Promise.resolve();
    });
    await waitFor(() => rerender(ui));
    expect(screen.queryByText(/Your password has been updated./i)).toBeInTheDocument();
  });
});
