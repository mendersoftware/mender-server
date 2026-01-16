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
import { defaultState, render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import DeviceConnectionDialog from './DeviceConnectionDialog';

describe('DeviceConnectionDialog Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<DeviceConnectionDialog onCancel={vi.fn} />);
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('renders Zephyr correctly', async () => {
    const { baseElement } = render(<DeviceConnectionDialog onCancel={vi.fn} />, {
      preloadedState: {
        ...defaultState,
        app: {
          ...defaultState.app,
          features: {
            ...defaultState.app.features,
            hasMCUEnabled: true
          }
        }
      }
    });
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<DeviceConnectionDialog onCancel={vi.fn} />, {
      preloadedState: {
        ...defaultState,
        app: {
          ...defaultState.app,
          features: {
            ...defaultState.app.features,
            isHosted: true
          }
        }
      }
    });
    const getStartedButtons = screen.getAllByRole('button', { name: /get started/i });
    await user.click(getStartedButtons[0]);
    expect(screen.getByText(/Enter your device type/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /back/i }));
    await user.click(screen.getByText(/Try the virtual device/i));
    expect(screen.getByText(/run the following command to start the virtual device/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Waiting for device/i })).toBeInTheDocument();
    await act(async () => vi.runAllTicks());
  });
});
