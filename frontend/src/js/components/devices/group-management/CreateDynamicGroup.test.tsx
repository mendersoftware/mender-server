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
import { render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import CreateDynamicGroup from './CreateDynamicGroup';

describe('CreateDynamicGroup Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<CreateDynamicGroup onClose={vi.fn} onCreate={vi.fn} />);
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('only validates on submit', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onCreate = vi.fn();
    render(<CreateDynamicGroup onClose={vi.fn} onCreate={onCreate} />);
    await user.type(screen.getByPlaceholderText(/group name/i), 'testGroup');
    expect(screen.queryByText(/already exists/i)).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /create group/i }));
    await screen.findByText(/a group with the same name already exists/i);
    expect(onCreate).not.toHaveBeenCalled();
  });

  it('requires a name', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onCreate = vi.fn();
    render(<CreateDynamicGroup onClose={vi.fn} onCreate={onCreate} />);
    await user.click(screen.getByRole('button', { name: /create group/i }));
    await screen.findByText(/group name is required/i);
    expect(onCreate).not.toHaveBeenCalled();
  });

  it('creates a group with a valid name', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const onCreate = vi.fn();
    render(<CreateDynamicGroup onClose={vi.fn} onCreate={onCreate} />);
    await user.type(screen.getByPlaceholderText(/group name/i), 'brandNewGroup');
    await user.click(screen.getByRole('button', { name: /create group/i }));
    await waitFor(() => expect(onCreate).toHaveBeenCalledWith('brandNewGroup'));
  });
});
