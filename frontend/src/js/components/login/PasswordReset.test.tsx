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
import { Provider } from 'react-redux';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import { render } from '@/testUtils';
import { getConfiguredStore } from '@northern.tech/store/store';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, render as testingLibRender, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import Password from './Password';
import PasswordReset from './PasswordReset';

const goodPassword = 'mysecretpassword!123';
const badPassword = 'mysecretpassword!546';

vi.mock('@northern.tech/store/thunks', { spy: true });

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
    await act(() => vi.runAllTimersAsync());
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { passwordResetComplete: completeSpy } = StoreThunks;

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
    const passwordConfirmationInput = screen.getByLabelText(/confirm password \*/i);
    await user.type(passwordConfirmationInput, goodPassword);
    await waitFor(() => rerender(ui));
    expect(screen.getByRole('button', { name: /Save password/i })).toBeDisabled();
    expect(screen.getByText('The passwords you provided do not match, please check again.')).toBeVisible();
    await user.clear(passwordInput);
    await user.type(passwordInput, goodPassword);
    await waitFor(() => rerender(ui));
    await act(async () => {
      vi.runAllTicks();
      vi.runAllTimers();
    });
    const saveButton = screen.getByRole('button', { name: /Save password/i });
    await waitFor(() => expect(saveButton).not.toBeDisabled());
    await act(() => user.click(saveButton));
    await waitFor(() => expect(completeSpy).toHaveBeenCalledWith({ secretHash, newPassword: goodPassword }));
    await waitFor(() => expect(screen.queryByText(/Your password has been updated./i)).toBeVisible());
  });
});
