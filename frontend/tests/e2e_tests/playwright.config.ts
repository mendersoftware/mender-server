// Copyright 2021 Northern.tech AS
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
import type { LaunchOptions, PlaywrightTestConfig, ViewportSize } from '@playwright/test';
import { devices } from '@playwright/test';
import * as path from 'path';
import { fileURLToPath } from 'url';

import { storagePath, testDirBase } from './utils/constants';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export const viewport: ViewportSize = { width: 1600, height: 900 };

export const contextArgs = {
  acceptDownloads: true,
  ignoreHTTPSErrors: true,
  viewport
};

export const launchOptions: LaunchOptions = {
  ...contextArgs,
  args: process.env.TEST_ENVIRONMENT === 'staging' ? [] : ['--disable-dev-shm-usage', '--disable-web-security'],
  slowMo: process.env.TEST_ENVIRONMENT === 'staging' ? undefined : 50
  // to ease running the test locally and "headful" uncomment and modify the below option to match your preferred browser installation
  // this might also require adjusting the `runWith` call at the bottom of the file
  // executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
};

export const projectParamsByBrowser = {
  chrome: { ...devices['Desktop Chrome'], storageState: storagePath, permissions: ['clipboard-read'], viewport },
  firefox: { ...devices['Desktop Firefox'], storageState: storagePath, viewport },
  webkit: { ...devices['Desktop Safari'], storageState: storagePath, viewport }
};

const options: PlaywrightTestConfig = {
  forbidOnly: !!process.env.CI,
  projects: [
    { name: 'setup-chromium', testMatch: /.*\.setup\.ts/, use: { ...devices['Desktop Chrome'], viewport, permissions: ['clipboard-read'] } },
    { name: 'setup-firefox', testMatch: /.*\.setup\.ts/, use: { ...devices['Desktop Firefox'], viewport } },
    { name: 'setup-webkit', testMatch: /.*\.setup\.ts/, use: { ...devices['Desktop Safari'], viewport } },

    { name: 'basic-chromium', testDir: `${testDirBase}/01-basic`, use: projectParamsByBrowser.chrome, dependencies: ['setup-chromium'], workers: 4 },
    { name: 'basic-firefox', testDir: `${testDirBase}/01-basic`, use: projectParamsByBrowser.firefox, dependencies: ['setup-firefox'], workers: 4 },
    { name: 'basic-webkit', testDir: `${testDirBase}/01-basic`, use: projectParamsByBrowser.webkit, dependencies: ['setup-webkit'], workers: 4 },

    { name: 'baseline-chromium', testDir: `${testDirBase}/02-baseline`, use: projectParamsByBrowser.chrome, dependencies: ['basic-chromium'], workers: 4 },
    { name: 'baseline-firefox', testDir: `${testDirBase}/02-baseline`, use: projectParamsByBrowser.firefox, dependencies: ['basic-firefox'], workers: 4 },
    { name: 'baseline-webkit', testDir: `${testDirBase}/02-baseline`, use: projectParamsByBrowser.webkit, dependencies: ['basic-webkit'], workers: 4 },

    { name: 'advanced-chromium', testDir: `${testDirBase}/03-advanced`, use: projectParamsByBrowser.chrome, dependencies: ['baseline-chromium'], workers: 1 },
    { name: 'advanced-firefox', testDir: `${testDirBase}/03-advanced`, use: projectParamsByBrowser.firefox, dependencies: ['baseline-firefox'], workers: 1 },
    { name: 'advanced-webkit', testDir: `${testDirBase}/03-advanced`, use: projectParamsByBrowser.webkit, dependencies: ['baseline-webkit'], workers: 1 },

    { name: 'chromium', testDir: `${testDirBase}/09-risky`, use: projectParamsByBrowser.chrome, dependencies: ['advanced-chromium'], workers: 1 },
    { name: 'firefox', testDir: `${testDirBase}/09-risky`, use: projectParamsByBrowser.firefox, dependencies: ['advanced-firefox'], workers: 1 },
    { name: 'webkit', testDir: `${testDirBase}/09-risky`, use: projectParamsByBrowser.webkit, dependencies: ['advanced-webkit'], workers: 1 }
  ],
  reporter: process.env.CI
    ? [
        ['line'],
        [
          '@bgotink/playwright-coverage',
          {
            sourceRoot: __dirname,
            exclude: ['**/webpack/**', '**/src/less/**'],
            resultDir: path.join(__dirname, 'coverage'),
            // Configure the reports to generate.
            // The value is an array of istanbul reports, with optional configuration attached.
            reports: [['lcov'], ['text-summary', { file: null }]],
            rewritePath: ({ relativePath }) => {
              let sourcePath;
              if (process.env.GUI_REPOSITORY) {
                sourcePath = path.join(process.env.GUI_REPOSITORY, relativePath.substring(relativePath.indexOf('/')));
              } else {
                sourcePath = path.join(__dirname, '..', '..', relativePath.substring(relativePath.indexOf('/')));
              }
              return sourcePath;
            }
          }
        ],
        ['junit', { outputFile: 'junit/results.xml' }]
      ]
    : 'line',
  // Two retries for each test.
  retries: 2,
  testDir: testDirBase,
  testIgnore: '**/.*-qemu-dependent/**',
  timeout: 60000,
  use: {
    ...contextArgs,
    contextOptions: contextArgs,
    screenshot: 'only-on-failure',
    trace: 'on',
    video: 'retain-on-failure',
    // headless: false,
    launchOptions
  }
};

export default options;
