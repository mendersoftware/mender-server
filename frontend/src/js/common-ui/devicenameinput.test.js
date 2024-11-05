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

import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../tests/mockData';
import { render } from '../../../tests/setupTests';
import DeviceNameInput from './devicenameinput';

describe('DeviceNameInput Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<DeviceNameInput device={defaultState.devices.byId.a1} isHovered />);
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const ui = <DeviceNameInput device={{ ...defaultState.devices.byId.a1, tags: { name: 'testname' } }} isHovered />;
    const { rerender, store } = render(ui);
    expect(screen.queryByDisplayValue(/testname/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button'));
    await waitFor(() => rerender(ui));
    await user.type(screen.getByDisplayValue(/testname/i), 'something');
    await user.click(screen.getAllByRole('button')[0]);
    await act(async () => jest.runAllTicks());
    await waitFor(() => expect(store.getState().app.snackbar.message).toBe('Device name changed'));
    await waitFor(() => expect(store.getState().devices.byId.a1.tags).toStrictEqual({ name: 'testnamesomething' }));
  });
});
