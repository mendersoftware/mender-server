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

import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { undefineds } from '../../../tests/mockData';
import { render } from '../../../tests/setupTests';
import Confirm from './confirm';

describe('Confirm Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Confirm type="abort" />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const actionCheck = jest.fn();
    const cancelCheck = jest.fn();

    const { container } = render(<Confirm type="chartRemoval" action={actionCheck} cancel={cancelCheck} />);

    expect(screen.queryByText(/remove this chart\?/i)).toBeInTheDocument();
    await user.click(container.querySelector('.green'));
    expect(actionCheck).toHaveBeenCalledTimes(1);
    expect(screen.queryByText(/removing/i)).toBeInTheDocument();
    await user.click(container.querySelector('.red'));
    expect(cancelCheck).toHaveBeenCalledTimes(1);
  });
});
