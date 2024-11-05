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

import { yes } from '@northern.tech/store/constants';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { undefineds } from '../../../tests/mockData';
import { render } from '../../../tests/setupTests';
import CopyCode from './copy-code';

describe('CopyCode Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<CopyCode code="sudo it all!" />);
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const submitCheck = jest.fn();
    document.execCommand = jest.fn(yes);
    const ui = <CopyCode code="sudo it all!" onCopy={submitCheck} withDescription={true} />;
    const { rerender } = render(ui);

    expect(screen.queryByText(/Copied to clipboard/i)).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /Copy to clipboard/i }));
    expect(submitCheck).toHaveBeenCalledTimes(1);
    expect(document.execCommand).toHaveBeenCalledTimes(1);
    expect(screen.queryByText(/Copied to clipboard/i)).toBeInTheDocument();
    act(() => jest.advanceTimersByTime(6000));
    await waitFor(() => rerender(ui));
    expect(screen.queryByText(/Copied to clipboard/i)).not.toBeInTheDocument();
  });
});
