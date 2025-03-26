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

import { actions } from '@northern.tech/store/appSlice';
import { yes } from '@northern.tech/store/constants';
import { useAppDispatch } from '@northern.tech/store/store';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';

import { defaultState, undefineds } from '../../../tests/mockData';
import { render } from '../../../tests/setupTests';
import SharedSnackbar from './SharedSnackbar';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    snackbar: {
      ...defaultState.snackbar,
      message: 'test',
      open: true
    }
  }
};
const preloadedStateNoCopy = {
  ...defaultState,
  app: {
    ...defaultState.app,
    snackbar: {
      ...defaultState.snackbar,
      message: 'test',
      preventClickToCopy: true,
      open: true
    }
  }
};

describe('SharedSnackbar Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(<SharedSnackbar />, { preloadedState });
    const view = baseElement.firstChild.firstChild;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as intended', async () => {
    const mockDispatch = vi.fn();
    useAppDispatch.mockReturnValue(mockDispatch);
    vi.mock('@northern.tech/store/store', async importOriginal => {
      const actual = await importOriginal();
      return {
        ...actual,
        useAppDispatch: vi.fn()
      };
    });
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const copyCheck = vi.fn(yes);
    document.execCommand = copyCheck;

    render(<SharedSnackbar />, { preloadedState });
    expect(screen.queryByText(/test/i)).toBeInTheDocument();
    await user.click(screen.getByText(/test/i));
    expect(mockDispatch).toHaveBeenCalledWith(actions.setSnackbar('Copied to clipboard'));
    expect(copyCheck).toHaveBeenCalled();
  });

  it('works as intended with a click listener', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    const actionCheck = vi.fn();
    const copyCheck = vi.fn(yes);
    document.execCommand = copyCheck;

    render(<SharedSnackbar />, { preloadedState: preloadedStateNoCopy });
    await user.click(screen.getByText(/test/i));
    expect(actionCheck).not.toHaveBeenCalled();
    expect(copyCheck).not.toHaveBeenCalled();
  });
});
