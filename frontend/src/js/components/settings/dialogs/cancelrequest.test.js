// Copyright 2020 Northern.tech AS
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

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { undefineds } from '../../../../../tests/mockData';
import CancelRequestDialog from './cancelrequest';

describe('CancelRequestDialog Component', () => {
  beforeEach(() => {
    Math.random = jest.fn(() => 0);
  });
  it('renders correctly', async () => {
    const { baseElement } = render(<CancelRequestDialog />);
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const submitMock = jest.fn();
    render(<CancelRequestDialog onCancel={jest.fn} onSubmit={submitMock} />);
    expect(screen.getByRole('button', { name: /Continue/i })).toBeDisabled();
    await user.click(screen.getByRole('radio', { name: /My project is delayed/i }));
    expect(screen.getByRole('button', { name: /Continue/i })).not.toBeDisabled();
    await user.click(screen.getByRole('radio', { name: /other/i }));
    expect(screen.getByRole('button', { name: /Continue/i })).toBeDisabled();
    await user.type(screen.getByPlaceholderText(/reason/i), 'test reason');
    expect(screen.getByRole('button', { name: /Continue/i })).not.toBeDisabled();
    await user.type(screen.getByPlaceholderText(/suggestions/i), 'test suggestion');
    await user.click(screen.getByRole('button', { name: /Continue/i }));

    expect(screen.queryByText(/thank you/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /Confirm/i }));

    expect(submitMock).toHaveBeenCalledWith(`test reason\ntest suggestion`);
  });
});
