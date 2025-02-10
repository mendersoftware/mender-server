// Copyright 2025 Northern.tech AS
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

import { createSerializer } from '@emotion/jest';
import { ThemeProvider, createTheme } from '@mui/material/styles';

import { getSessionInfo } from '@northern.tech/store/auth';
import { yes } from '@northern.tech/store/constants';
import { getConfiguredStore } from '@northern.tech/store/store';
import '@testing-library/jest-dom';
import { act, cleanup, queryByRole, render, waitFor, within } from '@testing-library/react';
import { setupServer } from 'msw/node';
import { afterAll, afterEach, beforeAll, vi } from 'vitest';
import { MessageChannel } from 'worker_threads';

import { light as lightTheme } from '../src/js/themes/Mender';
import handlers from './__mocks__/requestHandlers';
import { defaultState, menderEnvironment, mockDate, token as mockToken } from './mockData';

process.on('unhandledRejection', err => {
  throw err;
});

expect.addSnapshotSerializer(createSerializer({ includeStyles: true }));

afterEach(() => {
  cleanup();
});
export const TEST_LOCATION = 'localhost';

export const mockAbortController = { signal: { addEventListener: () => {}, removeEventListener: () => {} } };

// Setup requests interception
const server = setupServer(...handlers);

const oldWindowLocalStorage = window.localStorage;
const oldWindowLocation = window.location;
const oldWindowSessionStorage = window.sessionStorage;

vi.mock('universal-cookie', () => {
  const mockCookie = {
    get: vi.fn(),
    set: vi.fn(),
    remove: vi.fn()
  };
  return { default: vi.fn(() => mockCookie) };
});

vi.mock('uuid', () => ({ v4: () => 'mock-uuid' }));

vi.useFakeTimers({ now: mockDate });
vi.setSystemTime(mockDate);

const storage = {};
global.HTMLCanvasElement.prototype.getContext = vi.fn();

beforeAll(async () => {
  // Temporarily workaround for bug in @testing-library/react when use user-event with `vi.useFakeTimers()`

  // Enable the mocking in tests.
  delete window.location;
  window.location = {
    ...oldWindowLocation,
    hostname: TEST_LOCATION,
    origin: 'http://localhost',
    assign: vi.fn(),
    replace: vi.fn(),
    reload: vi.fn()
  };
  delete window.sessionStorage;
  window.sessionStorage = {
    ...oldWindowSessionStorage,
    getItem: vi.fn(yes),
    setItem: vi.fn(),
    removeItem: vi.fn()
  };
  delete window.localStorage;
  window.localStorage = {
    ...oldWindowLocalStorage,
    getItem: vi.fn(name => {
      if (name === 'JWT') {
        return JSON.stringify({ token: mockToken });
      }
      return storage[name];
    }),
    setItem: vi.fn(name => storage[name]),
    removeItem: vi.fn()
  };
  window.mender_environment = menderEnvironment;
  window.ENV = 'test';
  global.AbortController = vi.fn().mockImplementation(() => mockAbortController);
  global.MessageChannel = MessageChannel;
  global.ResizeObserver = vi.fn().mockImplementation(() => ({
    observe: vi.fn(),
    unobserve: vi.fn(),
    disconnect: vi.fn()
  }));
  window.RTCPeerConnection = () => {
    return {
      createOffer: () => {},
      setLocalDescription: () => {},
      createDataChannel: () => {}
    };
  };

  createMocks();
  await server.listen({ onUnhandledRequest: 'error' });
  Object.defineProperty(navigator, 'appVersion', { value: 'Test', writable: true });
  const intersectionObserverMock = () => ({
    observe: vi.fn,
    disconnect: vi.fn
  });
  window.IntersectionObserver = vi.fn().mockImplementation(intersectionObserverMock);
  vi.spyOn(React, 'useEffect').mockImplementation(React.useLayoutEffect);

  //TODO: remove, once https://github.com/testing-library/react-testing-library/issues/1197 resolved
  const _jest = globalThis.jest;

  globalThis.jest = {
    ...globalThis.jest,
    advanceTimersByTime: vi.advanceTimersByTime.bind(vi)
  };
  return () => void (globalThis.jest = _jest);
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
const theme = createTheme(lightTheme);

export const selectMaterialUiSelectOption = async (element, optionText, user) => {
  // The button that opens the dropdown, which is a sibling of the input
  const selectButton = element.parentNode.querySelector('[role=combobox]');
  // Open the select dropdown
  await act(async () => await user.click(selectButton));
  // Get the dropdown element. We don't use getByRole() because it includes <select>s too.
  const listbox = queryByRole(document.documentElement, 'listbox');
  // Click the list item
  const listItem = within(listbox).getByText(optionText);
  await user.click(listItem);
  // Wait for the listbox to be removed, so it isn't visible in subsequent calls
  await waitFor(() => expect(queryByRole(document.documentElement, 'listbox')).not.toBeInTheDocument());
  return Promise.resolve();
};

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

export * from '@testing-library/react';
// override render method
// eslint-disable-next-line import/export
export { customRender as render };
