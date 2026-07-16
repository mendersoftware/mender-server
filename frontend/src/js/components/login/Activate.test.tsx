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
import { Provider } from 'react-redux';
import { MemoryRouter, Route, Routes } from 'react-router';

import { getConfiguredStore } from '@northern.tech/store/store';
import * as StoreThunks from '@northern.tech/store/thunks';
import { act, render, screen, waitFor } from '@testing-library/react';
import { vi } from 'vitest';

import Activate from './Activate';

describe('Activate Component', () => {
  let store;
  beforeEach(() => {
    store = getConfiguredStore();
  });

  const renderAtPath = path =>
    render(
      <Provider store={store}>
        <MemoryRouter initialEntries={[path]}>
          <Routes>
            <Route path="activate/:code" element={<Activate />} />
            <Route path="email-change/:secretHash" element={<Activate />} />
          </Routes>
        </MemoryRouter>
      </Provider>
    );

  it('verifies the email with the code from the url', async () => {
    const { verifyEmailComplete: completeSpy } = StoreThunks;
    const code = 'superSecret';

    renderAtPath(`/activate/${code}`);

    expect(completeSpy).toHaveBeenCalledWith(code);
    await act(() => vi.runAllTimersAsync());
    await waitFor(() => expect(screen.getByText(/Your new email address has been successfully confirmed./i)).toBeVisible());
  });

  it('completes the email change with the hash from the url', async () => {
    const { completeEmailChange: completeSpy } = StoreThunks;
    const secretHash = 'superSecret';

    renderAtPath(`/email-change/${secretHash}`);

    expect(completeSpy).toHaveBeenCalledWith(secretHash);
    await act(() => vi.runAllTimersAsync());
    await waitFor(() => expect(screen.getByText(/Your new email address has been successfully confirmed./i)).toBeVisible());
  });
});
