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
import type { PlaywrightTestConfig } from '@playwright/test';
import { devices } from '@playwright/test';

import { contextArgs, launchOptions, projectParamsByBrowser, viewport } from './playwright.config';
import { testDirBase } from './utils/constants';

const options: PlaywrightTestConfig = {
  forbidOnly: !!process.env.CI,
  projects: [
    { name: 'setup-chromium', testMatch: /.*\.setup\.ts/, use: { ...devices['Desktop Chrome'], viewport, permissions: ['clipboard-read'] } },
    { name: 'basic-qemu', testDir: `${testDirBase}/01-basic`, use: projectParamsByBrowser.chrome, dependencies: ['setup-chromium'], workers: 4 },
    { name: 'qemu-tests', testDir: `${testDirBase}/04-qemu-dependent`, use: projectParamsByBrowser.chrome, dependencies: ['setup-chromium'], workers: 1 }
  ],
  reporter: process.env.CI ? [['line'], ['junit', { outputFile: 'junit/results.xml' }]] : 'line',
  retries: 2,
  testDir: testDirBase,
  timeout: 180000,
  use: {
    ...contextArgs,
    contextOptions: contextArgs,
    screenshot: 'only-on-failure',
    trace: 'on',
    video: 'retain-on-failure',
    launchOptions
  }
};

export default options;
