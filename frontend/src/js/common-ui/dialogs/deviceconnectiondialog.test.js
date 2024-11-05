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

import { act, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import DeviceConnectionDialog from './deviceconnectiondialog';

describe('DeviceConnectionDialog Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<DeviceConnectionDialog onCancel={jest.fn} />);
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    render(<DeviceConnectionDialog onCancel={jest.fn} />);
    await user.click(screen.getByText(/get started/i));
    expect(screen.getByText(/Enter your device type/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /back/i }));
    await user.click(screen.getByText(/Try a virtual device/i));
    expect(screen.getByText(/run the following command to start the virtual device/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Waiting for device/i })).toBeInTheDocument();
    await act(async () => jest.runAllTicks());
  });
});
