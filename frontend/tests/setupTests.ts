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
import { cleanUp } from '@northern.tech/store/auth';
import handlers from '@northern.tech/testing/requestHandlers/requestHandlers';
import { afterAll as ntAfterAll, afterEach as ntAfterEach, beforeAll as ntBeforeAll, beforeEach as ntBeforeEach } from '@northern.tech/testing/setupTests';
import '@testing-library/jest-dom/vitest';
import { setupServer } from 'msw/node';
import { afterAll, afterEach, beforeAll, beforeEach, expect, vi } from 'vitest';

import Tracking from '../src/js/tracking';

// monaco editor is not available in jsdom, stub it to avoid dom modifications
vi.mock('@monaco-editor/react', () => ({
  default: () => null,
  DiffEditor: () => null,
  loader: { config: () => {} }
}));

// avoid appending the recaptcha input
vi.mock('react-google-recaptcha', () => ({
  default: () => null
}));

vi.mock('@northern.tech/store/thunks', { spy: true });

// Avoid inserting links in DOM
Tracking.cookieconsent = () => Promise.resolve({ trackingConsentGiven: false });

// Unified id's assigned by react to stabilize snapshots
expect.addSnapshotSerializer({
  test: val => typeof val === 'string' && /_r_[a-z0-9]+_/.test(val),
  serialize: (val, config, indent, depth, refs, printer) => printer(val.replace(/_r_[a-z0-9]+_/g, '_r_X_'), config, indent, depth, refs)
});

process.on('unhandledRejection', err => {
  throw err;
});

// Setup requests interception
export const server = setupServer(...handlers);

// ensure consistent snapshots across dev machines and CI
// - module loading order prevents this from fitting into the regular hooks
Object.defineProperty(process, 'platform', { value: 'linux', writable: true });

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(function (query) {
    return {
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn()
    };
  })
});

beforeAll(async () => {
  await server.listen({ onUnhandledRequest: 'error' });
  await ntBeforeAll({ expect, vi });
});

beforeEach(async () => {
  // reset module-level tokenCache
  cleanUp();
  await ntBeforeEach({ vi });
});

afterEach(async () => {
  await ntAfterEach({ vi });
  // Reset any runtime handlers tests may use.
  await server.resetHandlers();
  // clean-up after MUI modals append styles to body
  document.body.removeAttribute('style');
});

afterAll(async () => {
  // Clean up once the tests are done.
  await server.close();
  await ntAfterAll({ vi });
});
