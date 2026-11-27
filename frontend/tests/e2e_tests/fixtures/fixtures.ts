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
import type { Page } from '@playwright/test';
import { test as nonCoveredTest } from '@playwright/test';

import { getPersistentLoginInfo, isEnterpriseOrStaging, login, resetServerErrors, watchServerErrors } from '../utils/commands.ts';
import { timeouts } from '../utils/constants.ts';

export type TestEnvironment = 'enterprise' | 'staging' | 'os';

type TestFixtures = {
  adminBaseUrl: string;
  config: unknown;
  demoDeviceName: string;
  demoDeviceSoftware: string;
  page: Page;
  serverErrorReport: void;
  spTenantUsername: string;
  staticGroupName: string;
  uniqueId: string;
};

type WorkerFixtures = {
  baseUrl: string;
  environment: TestEnvironment;
  isRemote: boolean;
  isScheduledRun: boolean;
  password: string;
  tenantToken: string;
  username: string;
};

export type CombinedFixtures = TestFixtures & WorkerFixtures;

const urls = {
  localhost: 'https://docker.mender.io/',
  staging: 'https://staging.hosted.mender.io/',
  production: 'https://hosted.mender.io/'
};

const adminUrls = {
  localhost: 'https://admin.docker.mender.io/',
  staging: 'https://admin.staging.hosted.mender.io/',
  production: 'https://admin.hosted.mender.io/'
};

const defaultConfig = {
  adminBaseUrl: adminUrls.localhost,
  baseUrl: urls.localhost,
  spTenantUsername: 'tenant-demo@example.com',
  username: 'mender-demo@example.com',
  password: 'mysecretpassword!123',
  demoDeviceName: 'theDeviceName',
  demoDeviceSoftware: 'original',
  staticGroupName: 'test.group'
};

const test = (process.env.BASE_URL ? nonCoveredTest : coveredTest).extend<TestFixtures, WorkerFixtures>({
  isRemote: [!!process.env.BASE_URL, { scope: 'worker' }],
  isScheduledRun: [!!process.env.IS_SCHEDULED, { scope: 'worker' }],
  // auto fixtures are set up first and torn down last, so this also covers pages & contexts created through prepareNewPage & prepareNewContext
  serverErrorReport: [
    // eslint-disable-next-line no-empty-pattern
    async ({}, use, testInfo) => {
      const serverErrors = resetServerErrors();
      await use();
      if (serverErrors.length) {
        await testInfo.attach('server-errors', { body: serverErrors.join('\n'), contentType: 'text/plain' });
      }
    },
    { auto: true }
  ],
  context: async ({ baseUrl, context, isRemote }, use) => {
    const stopWatching = isRemote ? await watchServerErrors(context, baseUrl) : undefined;
    await use(context);
    await stopWatching?.();
  },
  page: async ({ baseUrl, page }, use, { project }) => {
    if (project.name !== 'admin-panel') {
      await page.goto(baseUrl);
      await page.waitForTimeout(timeouts.oneSecond);
    }
    await use(page);
  },
  environment: [
    // eslint-disable-next-line no-empty-pattern
    async ({}, use) => {
      const environment = (process.env.TEST_ENVIRONMENT ? process.env.TEST_ENVIRONMENT : 'os') as TestEnvironment;
      await use(environment);
    },
    { scope: 'worker' }
  ],
  spTenantUsername: async ({ isRemote }, use) => {
    const spTenantUsername = isRemote ? getPersistentLoginInfo().tenantUsername : defaultConfig.spTenantUsername;
    await use(spTenantUsername);
  },
  uniqueId: async ({ isRemote }, use) => {
    const uniqueId = isRemote ? getPersistentLoginInfo().uniqueId : '';
    await use(uniqueId);
  },
  username: [
    async ({ isRemote }, use) => {
      const username = isRemote ? getPersistentLoginInfo().username : defaultConfig.username;
      await use(username);
    },
    { scope: 'worker' }
  ],
  password: [
    async ({ isRemote }, use) => {
      const password = isRemote ? getPersistentLoginInfo().password : defaultConfig.password;
      await use(password);
    },
    { scope: 'worker' }
  ],
  adminBaseUrl: async ({ environment, isRemote }, use) => {
    const adminBaseUrl = process.env.ADMIN_BASE_URL ?? adminUrls[environment] ?? (isRemote ? '' : defaultConfig.adminBaseUrl);
    await use(adminBaseUrl);
  },
  baseUrl: [
    async ({ environment }, use) => {
      const baseUrl = process.env.BASE_URL ?? urls[environment] ?? defaultConfig.baseUrl;
      await use(baseUrl);
    },
    { scope: 'worker' }
  ],
  demoDeviceName: defaultConfig.demoDeviceName,
  demoDeviceSoftware: defaultConfig.demoDeviceSoftware,
  staticGroupName: defaultConfig.staticGroupName,
  tenantToken: [
    async ({ baseUrl, environment, password, playwright, username }, use) => {
      if (!isEnterpriseOrStaging(environment)) {
        await use('');
        return;
      }
      // the built-in request fixture is test scoped, so a request context of its own is used to retrieve the token
      const request = await playwright.request.newContext({ ignoreHTTPSErrors: true });
      const { token } = await login(username, password, baseUrl, request);
      const response = await request.get(`${baseUrl}api/management/v2/tenantadm/tenants/me/token`, { headers: { Authorization: `Bearer ${token}` } });
      expect(response.ok(), await response.text()).toBeTruthy();
      const { tenant_token: tenantToken } = await response.json();
      await request.dispose();
      await use(tenantToken);
    },
    { scope: 'worker' }
  ]
});

export { expect };
export default test;
