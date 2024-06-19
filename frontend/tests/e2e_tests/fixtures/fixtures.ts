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
import { test as coveredTest, expect } from '@bgotink/playwright-coverage';
import { Page, test as nonCoveredTest } from '@playwright/test';

import { baseUrlToDomain, getPeristentLoginInfo, isLoggedIn, login, prepareCookies } from '../utils/commands';
import { storagePath } from '../utils/constants';

type TestFixtures = {
  baseUrl: string;
  config: unknown;
  environment: string;
  loggedInPage: Page;
  username: string;
  password: string;
  demoDeviceName: string;
};

const urls = {
  localhost: 'https://docker.mender.io/',
  staging: 'https://staging.hosted.mender.io/',
  production: 'https://hosted.mender.io/'
};

const defaultConfig = {
  baseUrl: urls.localhost,
  username: 'mender-demo@example.com',
  password: 'mysecretpassword!123',
  demoDeviceName: 'original'
};

const test = (process.env.TEST_ENVIRONMENT === 'staging' ? nonCoveredTest : coveredTest).extend<TestFixtures>({
  loggedInPage: async ({ baseUrl, browserName, context, password, username }, use) => {
    if (!['firefox', 'webkit'].includes(browserName)) {
      await context.grantPermissions(['clipboard-read'], { origin: baseUrl });
    }
    const domain = baseUrlToDomain(baseUrl);
    const { token, userId } = await login(username, password, baseUrl);
    context = await prepareCookies(context, domain, userId);
    context.addInitScript(token => {
      window.localStorage.setItem('JWT', JSON.stringify({ token }));
      window.localStorage.setItem(`onboardingComplete`, 'true');
    }, token);
    const page = await context.newPage();
    await page.goto(`${baseUrl}ui/`);
    await isLoggedIn(page);
    const isHeaderComplete = await page.getByText(username).isVisible();
    if (!isHeaderComplete) {
      await page.reload();
      await page.waitForSelector(`text=${username}`);
    }
    await use(page);
    await context.storageState({ path: storagePath });
  },
  // eslint-disable-next-line no-empty-pattern
  environment: async ({}, use) => {
    const environment = process.env.TEST_ENVIRONMENT ? process.env.TEST_ENVIRONMENT : 'localhost';
    await use(environment);
  },
  username: async ({ environment }, use) => {
    let username = defaultConfig.username;
    if (environment === 'staging') {
      username = getPeristentLoginInfo().username;
    }
    await use(username);
  },
  password: async ({ environment }, use) => {
    let password = defaultConfig.password;
    if (environment === 'staging') {
      password = getPeristentLoginInfo().password;
    }
    await use(password);
  },
  baseUrl: async ({ environment }, use) => {
    const baseUrl = urls[environment] || defaultConfig.baseUrl;
    await use(baseUrl);
  },
  demoDeviceName: defaultConfig.demoDeviceName
});

export { expect };
export default test;
