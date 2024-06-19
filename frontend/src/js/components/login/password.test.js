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

import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import * as UserActions from '../../actions/userActions';
import Password from './password';

describe('Password Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Password />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const startSpy = jest.spyOn(UserActions, 'passwordResetStart');
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const ui = <Password />;
    const { rerender } = render(ui);
    await user.type(screen.queryByLabelText(/your email/i), 'something@example.com');
    await waitFor(() => rerender(ui));
    await user.click(screen.getByRole('button', { name: /Send password reset link/i }));
    await waitFor(() => expect(startSpy).toHaveBeenCalledWith('something@example.com'));
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(screen.queryByText(/sending you an email/i)).toBeInTheDocument(), { timeout: 5000 });
  });
});
