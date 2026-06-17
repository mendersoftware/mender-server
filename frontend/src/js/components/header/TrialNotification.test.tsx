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
import { render } from '@/testUtils';
import { undefineds } from '@northern.tech/testing/mockData';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import dayjs from 'dayjs';

import TrialNotification from './TrialNotification';

describe('TrialNotification Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<TrialNotification iconClassName="" sectionClassName="" />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders correctly with an expiration date', async () => {
    const { baseElement } = render(<TrialNotification expiration="2019-02-01T12:16:22.667Z" iconClassName="" sectionClassName="" />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('warns about the remaining time when the trial expires soon', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<TrialNotification expiration={dayjs().add(10, 'day').toISOString()} sectionClassName="" />);
    await user.hover(screen.getByRole('button', { name: /trial plan/i }));
    expect(await screen.findByText(/your trial ends in/i)).toBeInTheDocument();
  });

  it('shows the default message when the trial is not expiring soon', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<TrialNotification expiration={dayjs().add(1, 'year').toISOString()} sectionClassName="" />);
    await user.hover(screen.getByRole('button', { name: /trial plan/i }));
    expect(await screen.findByText(/free to try for 12 months/i)).toBeInTheDocument();
  });
});
