// Copyright 2021 Northern.tech AS
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
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import DeviceAdditionWidget from './DeviceAdditionWidget';

describe('DeviceAdditionWidget Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<DeviceAdditionWidget features={{}} onConnectClick={vi.fn} tenantCapabilities={{}} />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const clickMock = vi.fn();
    render(<DeviceAdditionWidget features={{}} onConnectClick={clickMock} tenantCapabilities={{}} />);
    await user.click(screen.getByRole('button', { name: /connect a new device/i }));
    expect(clickMock).toHaveBeenCalled();
  });
});
