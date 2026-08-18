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

import AccessTokenManagement, { AccessTokenCreationDialog } from './AccessTokenManagement';

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
    createSpy.mockClear();
    const ui = <AccessTokenManagement />;
    const { rerender } = render(ui, { preloadedState });
    await user.click(screen.getByRole('button', { name: /generate a token/i }));
    const generateButton = screen.getByRole('button', { name: /create token/i });
    await user.click(generateButton);
    expect(createSpy).not.toHaveBeenCalled();
    await user.type(screen.getByLabelText(/name/i), 'somename');
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

  it('renders AccessTokenCreationDialog correctly', () => {
    const { baseElement } = render(<AccessTokenCreationDialog onCancel={vi.fn} onGenerate={vi.fn} token="afreshtoken" userRoles={[]} />);
    const view = baseElement.getElementsByClassName('MuiPaper-root')[0];
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('allows revoking a token', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const { revokeToken: revokeSpy } = StoreThunks;
    render(<AccessTokenManagement />, { preloadedState });
    await user.click(screen.getAllByRole('button', { name: /revoke/i })[0]);
    expect(screen.getByText(/are you sure you want to revoke the token/i)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /revoke token/i }));
    await waitFor(() => expect(revokeSpy).toHaveBeenCalled());
  });
});
