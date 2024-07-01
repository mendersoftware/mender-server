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

import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../../tests/mockData';
import { render } from '../../../../tests/setupTests';
import Releases from './releases';

describe('Releases Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<Releases />);
    await act(async () => jest.advanceTimersByTime(1000));
    const view = baseElement.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
  });

  it('works as expected', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const preloadedState = {
      ...defaultState,
      releases: {
        ...defaultState.releases,
        selectedArtifact: defaultState.releases.byId.r1.artifacts[0],
        selectedRelease: defaultState.releases.byId.r1.name
      }
    };
    const ui = <Releases />;
    const { rerender } = render(ui, { preloadedState });
    await waitFor(() => expect(screen.queryAllByText(defaultState.releases.byId.r1.name)[0]).toBeInTheDocument());
    await user.click(screen.getAllByText(defaultState.releases.byId.r1.name)[0]);
    await user.click(screen.getByText(/qemux/i));
    expect(screen.queryByText(defaultState.releases.byId.r1.artifacts[0].description)).toBeVisible();
    await user.click(screen.getByRole('button', { name: /Remove this/i }));
    await waitFor(() => expect(screen.queryByRole('button', { name: /Cancel/i })).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /Cancel/i }));
    await waitFor(() => expect(screen.queryByRole('button', { name: /Cancel/i })).not.toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /Close/i }));
    await waitFor(() => rerender(ui));
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
    expect(screen.queryByText(/release information/i)).toBeFalsy();
  }, 20000);
  it('has working search handling as expected', async () => {
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    render(<Releases />);
    expect(screen.queryByText(/Filtered from/i)).not.toBeInTheDocument();
    await user.type(screen.getByPlaceholderText(/starts with/i), 'b1');
    await waitFor(() => expect(screen.queryByText(/Filtered from/i)).toBeInTheDocument(), { timeout: 2000 });
    expect(screen.queryByText(/Filtered from/i)).toBeInTheDocument();
    await act(async () => {
      jest.runOnlyPendingTimers();
      jest.runAllTicks();
    });
  });
});
