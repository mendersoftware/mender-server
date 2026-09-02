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
import { Route, Routes } from 'react-router';

import { render } from '@/testUtils';
import { TIMEOUTS } from '@northern.tech/store/constants';
import { undefineds } from '@northern.tech/testing/mockData';
import { defaultDeviceId } from '@northern.tech/testing/requestHandlers/deviceHandlers';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import SearchV2 from './SearchV2';

describe('SearchV2 Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<SearchV2 />);
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('opens on the keyboard shortcut, searches and navigates to the selected device', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const ui = (
      <Routes>
        <Route path="/" element={<SearchV2 />} />
        <Route path="/devices/accepted" element={<div>accepted devices route</div>} />
      </Routes>
    );
    render(ui);
    await user.keyboard('/');
    const input = await screen.findByPlaceholderText(/starting with/i);
    await user.type(input, defaultDeviceId);
    // the search term is debounced, so the request only goes out once the timers have moved past it
    await act(async () => {
      vi.advanceTimersByTime(TIMEOUTS.oneSecond);
      vi.runAllTicks();
    });
    const dialog = screen.getByRole('dialog');
    await waitFor(() => expect(dialog).toHaveTextContent(defaultDeviceId));
    expect(dialog.querySelector('mark')).toHaveTextContent(defaultDeviceId);
    expect(dialog).toHaveTextContent('raspberrypi4 · - · Latest activity: 12 days ago');

    await user.keyboard('{Enter}');
    await act(async () => {
      vi.advanceTimersByTime(TIMEOUTS.oneSecond);
      vi.runAllTicks();
    });
    await waitFor(() => expect(screen.getByText(/accepted devices route/i)).toBeVisible());
  });
});
