// Copyright 2022 Northern.tech AS
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
import { accessTokens, undefineds } from '@northern.tech/testing/mockData';
import { act, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import AccessTokenManagement, { AccessTokenCreationDialog, AccessTokenRevocationDialog } from './AccessTokenManagement';

vi.mock('@northern.tech/store/thunks', { spy: true });

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      isEnterprise: true
    }
  },
  users: {
    ...defaultState.users,
    byId: {
      ...defaultState.users.byId,
      [defaultState.users.currentUser]: {
        ...defaultState.users.byId[defaultState.users.currentUser],
        tokens: accessTokens
      }
    }
  }
};

describe('AccessTokenManagement Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<AccessTokenManagement />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
  it('works as expected', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { getTokens: getSpy, generateToken: createSpy } = StoreThunks;
    getSpy.mockReset();
    const ui = <AccessTokenManagement />;
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('button', { name: /generate a token/i }));
    const generateButton = screen.getByRole('button', { name: /create token/i });
    expect(generateButton).toBeDisabled();
    await user.type(screen.getByPlaceholderText(/name/i), 'somename');
    expect(generateButton).not.toBeDisabled();
    await user.click(generateButton);
    await waitFor(() => rerender(ui));
    expect(createSpy).toHaveBeenCalledWith({ expiresIn: 31536000, name: 'somename' });
    await act(async () => {
      vi.runOnlyPendingTimers();
      vi.runAllTicks();
    });
    await waitFor(() => rerender(ui));
    await waitFor(() => expect(getSpy).toHaveBeenCalledTimes(1));
    expect(screen.queryByRole('button', { name: /create token/i })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /close/i })).toBeInTheDocument();
    expect(screen.getByText('aNewToken')).toBeInTheDocument();
  });

  [AccessTokenCreationDialog, AccessTokenRevocationDialog].forEach(async (Component, index) => {
    it(`renders ${Component.displayName || Component.name} correctly`, () => {
      const { baseElement } = render(
        <Component onCancel={vi.fn} generateToken={vi.fn} revokeToken={vi.fn} setToken={vi.fn} token={index ? accessTokens[0] : 'afreshtoken'} userRoles={[]} />
      );
      const view = baseElement.getElementsByClassName('MuiPaper-root')[0];
      expect(view).toMatchSnapshot();
      expect(view).toEqual(expect.not.stringMatching(undefineds));
    });
  });
});
