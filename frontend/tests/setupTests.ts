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
import { createMocks } from 'react-idle-timer';

import { mockDate } from '@northern.tech/testing/mockData';
import handlers from '@northern.tech/testing/requestHandlers/requestHandlers';
import { afterAll as ntAfterAll, afterEach as ntAfterEach, beforeAll as ntBeforeAll } from '@northern.tech/testing/setupTests';
import '@testing-library/jest-dom/vitest';
import { setupServer } from 'msw/node';
import { afterAll, afterEach, beforeAll, vi } from 'vitest';

process.on('unhandledRejection', err => {
  throw err;
});

// Setup requests interception
const server = setupServer(...handlers);

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn()
  }))
});

vi.useFakeTimers({ now: mockDate });
vi.setSystemTime(mockDate);

beforeAll(async () => {
  createMocks();
  await server.listen({ onUnhandledRequest: 'error' });
  await ntBeforeAll();
});

afterEach(async () => {
  await ntAfterEach();
  // Reset any runtime handlers tests may use.
  await server.resetHandlers();
});

afterAll(async () => {
  // Clean up once the tests are done.
  await server.close();
  await ntAfterAll();
});
