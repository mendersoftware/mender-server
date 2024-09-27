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
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';

import { ThemeProvider, createTheme } from '@mui/material/styles';
import { LocalizationProvider } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';

import { TIMEOUTS } from '@northern.tech/store/constants';
import { getConfiguredStore } from '@northern.tech/store/store';
import { prettyDOM, screen, render as testingLibRender, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { light as lightTheme } from '../../../../src/js/themes/Mender';
import { defaultState, undefineds } from '../../../../tests/mockData';
import { render, selectMaterialUiSelectOption } from '../../../../tests/setupTests';
import AuditLogs from './auditlogs';

const preloadedState = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      hasAuditlogs: true,
      isEnterprise: true
    }
  }
};

const preloadedStateNoAuditlogs = {
  ...defaultState,
  app: {
    ...defaultState.app,
    features: {
      ...defaultState.app.features,
      hasAuditlogs: true,
      isEnterprise: false
    }
  }
};

describe('Auditlogs Component', () => {
  it('renders correctly', async () => {
    const { baseElement } = render(
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <AuditLogs />
      </LocalizationProvider>,
      { preloadedState: preloadedStateNoAuditlogs }
    );
    const view = prettyDOM(baseElement.firstChild, 100000, { highlight: false })
      .replace(/id="mui-[0-9]*"/g, '')
      .replace(/aria-labelledby="(mui-[0-9]* *)*"/g, '')
      .replace(/\\/g, '');
    expect(view).toMatchSnapshot();
    expect(view).toEqual(expect.not.stringMatching(undefineds));
  });

  it('works as expected', async () => {
    let store = getConfiguredStore({ preloadedState });
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    render(
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <Provider store={store}>
          <AuditLogs />
        </Provider>
      </LocalizationProvider>,
      { preloadedState }
    );
    const input = screen.getByPlaceholderText(/type/i);
    await user.type(input, 'art');
    await selectMaterialUiSelectOption(input, /artifact/i, user);
    await user.click(screen.getByText(/clear filter/i));
    await user.click(screen.getByRole('button', { name: /Download results as csv/i }));
    await user.click(screen.getByText(/open_terminal/i));
  });

  it('allows navigating by url as expected', async () => {
    let store = getConfiguredStore({ preloadedState });
    const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime });
    const theme = createTheme(lightTheme);
    const ui = (
      <LocalizationProvider dateAdapter={AdapterDayjs}>
        <ThemeProvider theme={theme}>
          <MemoryRouter initialEntries={['/auditlog?startDate=2018-01-01']}>
            <Provider store={store}>
              <AuditLogs />
            </Provider>
          </MemoryRouter>
        </ThemeProvider>
      </LocalizationProvider>
    );
    const { rerender } = testingLibRender(ui);
    await jest.advanceTimersByTimeAsync(TIMEOUTS.oneSecond);
    await waitFor(() => rerender(ui));
    await user.click(screen.getByText(/clear filter/i));
  });
});
