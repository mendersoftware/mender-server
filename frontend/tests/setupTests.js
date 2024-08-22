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
import { createMocks } from 'react-idle-timer';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';

import { ThemeProvider, createTheme } from '@mui/material/styles';

import '@testing-library/jest-dom';
import { act, cleanup, queryByRole, render, waitFor, within } from '@testing-library/react';
import { setupServer } from 'msw/node';
import { MessageChannel } from 'worker_threads';

import { getSessionInfo } from '../src/js/auth';
import { yes } from '../src/js/constants/appConstants';
import { getConfiguredStore } from '../src/js/reducers';
import { light as lightTheme } from '../src/js/themes/Mender';
import handlers from './__mocks__/requestHandlers';
import { defaultState, menderEnvironment, mockDate, token as mockToken } from './mockData';

export const RETRY_TIMES = 3;
export const TEST_LOCATION = 'localhost';

export const mockAbortController = { signal: { addEventListener: () => {}, removeEventListener: () => {} } };

// Setup requests interception
let server;

const oldWindowLocalStorage = window.localStorage;
const oldWindowLocation = window.location;
const oldWindowSessionStorage = window.sessionStorage;

jest.retryTimes(RETRY_TIMES);
jest.mock('universal-cookie', () => {
  const mockCookie = {
    get: jest.fn(),
    set: jest.fn(),
    remove: jest.fn()
  };
  return jest.fn(() => mockCookie);
});

jest.mock('uuid', () => ({ v4: () => 'mock-uuid' }));

jest.setSystemTime(mockDate);
jest.useFakeTimers({ now: mockDate, doNotFake: ['queueMicrotask'] });

const storage = {};
global.HTMLCanvasElement.prototype.getContext = jest.fn();

beforeAll(async () => {
  // Enable the mocking in tests.
  delete window.location;
  window.location = {
    ...oldWindowLocation,
    hostname: TEST_LOCATION,
    assign: jest.fn(),
    replace: jest.fn(),
    reload: jest.fn()
  };
  delete window.sessionStorage;
  window.sessionStorage = {
    ...oldWindowSessionStorage,
    getItem: jest.fn(yes),
    setItem: jest.fn(),
    removeItem: jest.fn()
  };
  delete window.localStorage;
  window.localStorage = {
    ...oldWindowLocalStorage,
    getItem: jest.fn(name => {
      if (name === 'JWT') {
        return JSON.stringify({ token: mockToken });
      }
      return storage[name];
    }),
    setItem: jest.fn(name => storage[name]),
    removeItem: jest.fn()
  };
  window.mender_environment = menderEnvironment;
  window.ENV = 'test';
  global.AbortController = jest.fn().mockImplementation(() => mockAbortController);
  global.MessageChannel = MessageChannel;
  global.ResizeObserver = jest.fn().mockImplementation(() => ({
    observe: jest.fn(),
    unobserve: jest.fn(),
    disconnect: jest.fn()
  }));
  window.RTCPeerConnection = () => {
    return {
      createOffer: () => {},
      setLocalDescription: () => {},
      createDataChannel: () => {}
    };
  };
  createMocks();
  server = setupServer(...handlers);
  await server.listen({ onUnhandledRequest: 'error' });
  Object.defineProperty(navigator, 'appVersion', { value: 'Test', writable: true });
  const intersectionObserverMock = () => ({
    observe: jest.fn,
    disconnect: jest.fn
  });
  window.IntersectionObserver = jest.fn().mockImplementation(intersectionObserverMock);
  jest.spyOn(React, 'useEffect').mockImplementation(React.useLayoutEffect);
});

afterEach(async () => {
  // Reset any runtime handlers tests may use.
  await server.resetHandlers();
});

afterAll(async () => {
  // Clean up once the tests are done.
  await server.close();
  // restore `window.location` etc. to the original `jsdom` `Location` object
  window.localStorage = oldWindowLocalStorage;
  window.location = oldWindowLocation;
  window.sessionStorage = oldWindowSessionStorage;
  React.useEffect.mockRestore();
  cleanup();
});

export const selectMaterialUiSelectOption = async (element, optionText, user) => {
  // The button that opens the dropdown, which is a sibling of the input
  const selectButton = element.parentNode.querySelector('[role=combobox]');
  // Open the select dropdown
  await act(async () => await user.click(selectButton));
  // Get the dropdown element. We don't use getByRole() because it includes <select>s too.
  const listbox = document.body.querySelector('ul[role=listbox]');
  // Click the list item
  const listItem = within(listbox).getByText(optionText);
  await user.click(listItem);
  // Wait for the listbox to be removed, so it isn't visible in subsequent calls
  await waitFor(() => expect(queryByRole(document.documentElement, 'listbox')).not.toBeInTheDocument());
  return Promise.resolve();
};

const theme = createTheme(lightTheme);

const customRender = (ui, options = {}) => {
  const {
    preloadedState = { ...defaultState, users: { ...defaultState.users, currentSession: getSessionInfo() } },
    store = getConfiguredStore({ preloadedState }),
    ...remainder
  } = options;
  const AllTheProviders = ({ children }) => (
    <ThemeProvider theme={theme}>
      <MemoryRouter>
        <Provider store={store}>{children}</Provider>
      </MemoryRouter>
    </ThemeProvider>
  );
  return { store, ...render(ui, { wrapper: AllTheProviders, ...remainder }) };
};

// re-export everything
// eslint-disable-next-line import/export
export * from '@testing-library/react';
// override render method
// eslint-disable-next-line import/export
export { customRender as render };
