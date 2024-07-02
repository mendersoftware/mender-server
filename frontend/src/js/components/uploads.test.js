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

import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { defaultState, undefineds } from '../../../tests/mockData';
import { render } from '../../../tests/setupTests';
import Uploads from './uploads';

describe('Uploads Component', () => {
  it('renders correctly', async () => {
    const preloadedState = {
      ...defaultState,
      app: {
        ...defaultState.app,
        uploadsById: {
          ...defaultState.app.uploadsById,
          test: {
            name: 'testfile',
            size: 2048,
            uploadProgress: 20
          }
        }
      }
    };
    const { baseElement } = render(<Uploads />, { preloadedState });
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    user.hover(screen.getByRole('progressbar'));
    await waitFor(() => screen.queryByText(/in progress/i));
    const view = baseElement;
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });
});
