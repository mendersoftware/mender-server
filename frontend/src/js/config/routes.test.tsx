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
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';

import { ThemeProvider } from '@emotion/react';
import { createTheme } from '@mui/material';

import { getConfiguredStore } from '@northern.tech/store/store';
import { light as lightTheme } from '@northern.tech/themes/Mender';
import { act, screen, render as testingLibRender } from '@testing-library/react';
import { vi } from 'vitest';

import { defaultState } from '../../../tests/mockData';
import { PublicRoutes } from './routes';

const theme = createTheme(lightTheme);

describe('Router', () => {
  let store;

  beforeEach(() => {
    store = getConfiguredStore({
      preloadedState: {
        ...defaultState,
        app: {
          ...defaultState.app,
          features: {
            ...defaultState.features,
            isHosted: true
          }
        }
      }
    });
  });

  test('invalid path should redirect to Dashboard', async () => {
    testingLibRender(
      <ThemeProvider theme={theme}>
        <MemoryRouter initialEntries={['/random']}>
          <Provider store={store}>
            <PublicRoutes />
          </Provider>
        </MemoryRouter>
      </ThemeProvider>
    );
    expect(screen.getAllByText('Next')).toBeTruthy();
    expect(screen.queryByText('Settings')).toBeFalsy();
    await act(async () => vi.runAllTicks());
  });

  test('valid path should not redirect to 404', async () => {
    testingLibRender(
      <ThemeProvider theme={theme}>
        <MemoryRouter initialEntries={['/']}>
          <Provider store={store}>
            <PublicRoutes />
          </Provider>
        </MemoryRouter>
      </ThemeProvider>
    );
    expect(screen.getAllByText('Next')).toBeTruthy();
    expect(screen.queryByText('Settings')).toBeFalsy();
    await act(async () => vi.runAllTicks());
  });
});
