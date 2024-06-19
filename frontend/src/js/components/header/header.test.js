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

import { act, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import Header from './header';

const preloadedState = {
  ...defaultState,
  deployments: {
    ...defaultState.deployments,
    byStatus: {
      ...defaultState.deployments.byStatus,
      inprogress: {
        ...defaultState.deployments.byStatus.inprogress,
        total: 0
      }
    }
  },
  users: {
    ...defaultState.users,
    globalSettings: {
      ...defaultState.users.globalSettings,
      [defaultState.users.currentUser]: {
        ...defaultState.users.globalSettings[defaultState.users.currentUser],
        trackingConsentGiven: true
      }
    }
  }
};

describe('Header Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Header />, { preloadedState });
    const view = baseElement.getElementsByClassName('MuiDialog-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const view = <Header />;
    const { rerender } = render(view, { preloadedState });
    expect(screen.queryByText(defaultState.users.byId[defaultState.users.currentUser].email)).toBeInTheDocument();
    const selectButton = screen.getByRole('button', { name: defaultState.users.byId[defaultState.users.currentUser].email });
    await user.click(selectButton);
    const listbox = document.body.querySelector('ul[role=menu]');
    const listItem = within(listbox).getByText(/log out/i);
    await user.click(listItem);
    await act(async () => jest.runAllTicks());
    await waitFor(() => rerender(view));
    expect(screen.queryByText(defaultState.users.byId[defaultState.users.currentUser].email)).not.toBeInTheDocument();
  });
});
