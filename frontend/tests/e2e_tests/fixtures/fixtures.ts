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
import type { BrowserContext, Page } from '@playwright/test';
import { test as nonCoveredTest } from '@playwright/test';

import { getPeristentLoginInfo, isLoggedIn, prepareNewPage } from '../utils/commands.ts';
import { storagePath, timeouts } from '../utils/constants.ts';

type DemoArtifactVersionInfo = {
  artifactVersion: string;
  updateVersion: string;
};

export type TestEnvironment = 'enterprise' | 'staging' | 'localhost';

type TestFixtures = {
  baseUrl: string;
  config: unknown;
  demoArtifactVersion: DemoArtifactVersionInfo;
  demoDeviceName: string;
  environment: TestEnvironment;
  loggedInPage: Page;
  loggedInTenantPage: Page;
  password: string;
  spTenantUsername: string;
  username: string;
};

const urls = {
  localhost: 'https://docker.mender.io/',
  staging: 'https://staging.hosted.mender.io/',
  production: 'https://hosted.mender.io/'
};

const defaultConfig = {
  baseUrl: urls.localhost,
  spTenantUsername: 'tenant-demo@example.com',
  username: 'mender-demo@example.com',
  password: 'mysecretpassword!123',
  demoDeviceName: 'original'
};

const loginCommon = async ({
  page,
  username,
  storageLocation = storagePath,
  use,
  context
}: {
  context: BrowserContext;
  page: Page;
  storageLocation?: string;
  use: (r: Page) => Promise<void>;
  username: string;
}) => {
  await isLoggedIn(page);
  const isHeaderComplete = await page.getByText(username).isVisible();
  if (!isHeaderComplete) {
    await page.reload();
    await page.getByText(username).waitFor({ timeout: timeouts.default });
  }
  await context.storageState({ path: storageLocation });
  await use(page);
};
const test = (process.env.TEST_ENVIRONMENT === 'staging' ? nonCoveredTest : coveredTest).extend<TestFixtures>({
  loggedInPage: async ({ baseUrl, context, password, username }, use) => {
    const page = await prepareNewPage({ baseUrl, context, password, username });
    await loginCommon({ page, username, use, context });
  },
  loggedInTenantPage: async ({ baseUrl, context, password, spTenantUsername }, use) => {
    const storageLocation = `tenant-${storagePath}`;
    const page = await prepareNewPage({ baseUrl, context, password, storageLocation, username: spTenantUsername });
    await loginCommon({ page, username: spTenantUsername, use, context });
  },
  // eslint-disable-next-line no-empty-pattern
  environment: async ({}, use) => {
    const environment = process.env.TEST_ENVIRONMENT ? process.env.TEST_ENVIRONMENT : 'localhost';
    await use(environment);
  },
  spTenantUsername: async ({ environment }, use) => {
    test.skip(environment !== 'enterprise', 'not available in OS');
    let spTenantUsername = defaultConfig.spTenantUsername;
    if (environment === 'staging') {
      spTenantUsername = getPeristentLoginInfo().tenantUsername;
    }
    await use(spTenantUsername);
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
  demoDeviceName: defaultConfig.demoDeviceName,
  demoArtifactVersion: { artifactVersion: '3.8.0', updateVersion: '5.0.0' }
});

export { expect };
export default test;
