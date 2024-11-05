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

import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { render } from '../../../../tests/setupTests';
import Feedback from './feedback';

describe('Feedback Component', () => {
  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const ui = <Feedback />;
    const { rerender } = render(ui);
    await jest.runOnlyPendingTimersAsync();
    await user.click(screen.getByTitle('Satisfied'));
    await waitFor(() => rerender(ui));
    expect(screen.getByText(/the most important thing/i)).toBeVisible();
    await user.type(screen.getByPlaceholderText(/your feedback/i), 'some feedback');
    await user.click(screen.getByRole('button', { name: /submit/i }));
    expect(screen.getByText(/Thank you/i)).toBeVisible();
  });
});
