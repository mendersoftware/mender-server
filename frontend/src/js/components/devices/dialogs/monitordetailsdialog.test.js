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
import React from 'react';

import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { undefineds } from '../../../../../tests/mockData';
import { render } from '../../../../../tests/setupTests';
import LogDialog from './monitordetailsdialog';

const alert = {
  name: 'SSH Daemon is not running',
  subject: {
    name: 'sshd',
    type: 'systemd',
    status: 'not-running',
    details: {
      description: 'session closed for user root',
      line_matching: { 'line_number': 1234, 'data': 'Jul 22 10:40:56 raspberrypi sshd[32031]: pam_unix(sshd:session): session closed for user root' },
      lines_before: [
        { line_number: 0, data: 'Jul 22 10:40:57 raspberrypi sshd[32031]: pam_unix(sshd:session): session closed for user root' },
        { line_number: 137, data: 'Jul 22 10:40:58 raspberrypi sshd[32031]: pam_unix(sshd:session): session closed for user root' }
      ]
    }
  }
};

describe('MonitorLogDialog Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<LogDialog alert={alert} onClose={jest.fn} />);
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('renders correctly for description only', async () => {
    const logLessAlert = {
      ...alert,
      subject: {
        ...alert.subject,
        details: { description: alert.description }
      }
    };
    const { baseElement } = render(<LogDialog alert={logLessAlert} onClose={jest.fn} />);
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const ui = <LogDialog alert={alert} onClose={jest.fn} />;
    const { rerender } = render(ui);
    expect(screen.getByText(/show previous/i)).toBeInTheDocument();
    expect(screen.getByText(/137/i)).not.toBeVisible();
    await user.click(screen.getByText(/show previous/i));
    await waitFor(() => rerender(ui));
    expect(screen.getByText(/137/i)).toBeVisible();
    window.open = jest.fn();
    await user.click(screen.getByRole('button', { name: /export/i }));
    expect(window.open).toHaveBeenCalledWith(
      'data:application/octet-stream,0000%20%20%20Jul%2022%2010%3A40%3A57%20raspberrypi%20sshd%5B32031%5D%3A%20pam_unix(sshd%3Asession)%3A%20session%20closed%20for%20user%20root%0A0137%20%20%20Jul%2022%2010%3A40%3A58%20raspberrypi%20sshd%5B32031%5D%3A%20pam_unix(sshd%3Asession)%3A%20session%20closed%20for%20user%20root%0A1234%20%20%20Jul%2022%2010%3A40%3A56%20raspberrypi%20sshd%5B32031%5D%3A%20pam_unix(sshd%3Asession)%3A%20session%20closed%20for%20user%20root',
      'Mender-Monitor-SSH_Daemon_is_not_running.log'
    );
  });
});
