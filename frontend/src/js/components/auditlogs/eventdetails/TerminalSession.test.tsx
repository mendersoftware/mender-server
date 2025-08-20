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
import { defaultState, render } from '@/testUtils';
import * as StoreThunks from '@northern.tech/store/thunks';
import { undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import { vi } from 'vitest';

import TerminalSession from './TerminalSession';

vi.mock('@northern.tech/store/thunks', { spy: true });

describe('TerminalSession Component', () => {
  let socketSpyFactory;
  let socketSpy;

  beforeEach(() => {
    socketSpyFactory = vi.spyOn(window, 'WebSocket');
    socketSpyFactory.mockImplementation(() => {
      socketSpy = {
        close: () => {},
        send: () => {}
      };
      return socketSpy;
    });
  });

  afterEach(() => {
    socketSpyFactory.mockReset();
  });
  it('renders correctly', async () => {
    const { getSessionDetails: sessionSpy } = StoreThunks;
    const ui = <TerminalSession item={defaultState.organization.auditlog.events[2]} />;
    const { baseElement, rerender } = render(ui, {
      preloadedState: {
        ...defaultState,
        organization: {
          ...defaultState.organization,
          auditlog: {
            ...defaultState.organization.auditlog,
            selectionState: {
              ...defaultState.organization.auditlog.selectionState,
              selectedId: btoa(`${defaultState.organization.auditlog.events[2].action}|${defaultState.organization.auditlog.events[2].time}`)
            }
          }
        }
      }
    });
    await waitFor(() => rerender(ui));
    await act(async () => {
      vi.runAllTimers();
      vi.runAllTicks();
    });
    expect(sessionSpy).toHaveBeenCalled();
    await waitFor(() => expect(screen.queryByText(/Device type/i)).toBeVisible());
    await act(async () => {
      vi.runAllTimers();
      vi.runAllTicks();
    });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
