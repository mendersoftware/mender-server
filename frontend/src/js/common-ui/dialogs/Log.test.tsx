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
import { render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import LogDialog from './Log';

describe('LogDialog Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<LogDialog onClose={vi.fn} logData="things" />);
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders the config update log variant correctly', async () => {
    render(<LogDialog onClose={vi.fn} logData="things" type="configUpdateLog" />);
    expect(screen.getByText(/config update log for device/i)).toBeVisible();
    expect(screen.getByText('things')).toBeVisible();
  });

  it('renders the passed additional content', async () => {
    render(
      <LogDialog onClose={vi.fn} logData="things">
        <div>some log analysis</div>
      </LogDialog>
    );
    expect(screen.getByText(/some log analysis/i)).toBeVisible();
  });

  it('relies on the shared snackbar for copy confirmation', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { store } = render(<LogDialog onClose={vi.fn} logData="things" />);
    await user.click(screen.getByRole('button', { name: /copy to clipboard/i }));
    await waitFor(() => expect(store.getState().app.snackbar.message).toEqual('Copied to clipboard'));
    expect(screen.queryByText('Copied to clipboard.')).not.toBeInTheDocument();
  });
});
