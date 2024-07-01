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
import { LaunchOptions, PlaywrightTestConfig } from '@playwright/test';
import * as path from 'path';

const contextArgs = {
  acceptDownloads: true,
  ignoreHTTPSErrors: true,
  viewport: { width: 1600, height: 900 }
};

const launchOptions: LaunchOptions = {
  ...contextArgs,
  args: process.env.TEST_ENVIRONMENT === 'staging' ? [] : ['--disable-dev-shm-usage', '--disable-web-security'],
  slowMo: process.env.TEST_ENVIRONMENT === 'staging' ? undefined : 50
  // to ease running the test locally and "headful" uncomment and modify the below option to match your preferred browser installation
  // this might also require adjusting the `runWith` call at the bottom of the file
  // executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
};

export const contextOptions = {
  ...contextArgs,
  contextOptions: contextArgs,
  screenshot: 'only-on-failure',
  video: 'retain-on-failure',
  // headless: false,
  launchOptions
};

const options: PlaywrightTestConfig = {
  forbidOnly: !!process.env.CI,
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
  testDir: 'integration',
  timeout: 60000,
  use: contextOptions
};

export default options;
