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
import { expect } from '@playwright/test';
import * as fs from 'fs';

import test from '../fixtures/fixtures';
import { isEnterpriseOrStaging, isLoggedIn, login, prepareNewPage, startDockerClient, stopDockerClient, tenantTokenRetrieval } from '../utils/commands';
import { emptyStorageState, selectors, spStoragePath, storageFolder, storagePath, switchTenantStoragePath, timeouts } from '../utils/constants';

const pollDeployment = async ({
  location,
  attempt = 1,
  maxAttempts = 5,
  delay = timeouts.tenSeconds
}: {
  attempt: number;
  delay: number;
  location: string;
  maxAttempts: number;
}) => {
  const response = await fetch(location);
  const versionInfo = response.headers.get('x-mender-version')?.split('-') || [];
  const pipelineId = versionInfo.length ? Number(versionInfo[1]) : 0;
  const currentPipeline = Number(process.env.CI_PIPELINE_ID);
  if (pipelineId === currentPipeline) {
    return Promise.resolve();
  } else if (pipelineId > currentPipeline) {
    process.env.ALLOWED_TO_FAIL = '1';
    return Promise.resolve();
  }
  if (attempt >= maxAttempts) {
    throw new Error(`Couldn't get ${location} after ${maxAttempts} attempts`);
  }
  await new Promise(resolve => setTimeout(resolve, delay));
  return pollDeployment({ location, attempt: attempt + 1, maxAttempts, delay });
};

test.describe('Test setup', () => {
  test.beforeAll(async () => {
    if (!fs.existsSync(storageFolder)) {
      fs.mkdirSync(storageFolder, { recursive: true });
    }
    try {
      fs.unlinkSync('loginInfo.json');
      await stopDockerClient();
    } catch {
      // ...continue
    }
  });
  test('allows account creation', async ({ baseUrl, context, environment, page, password, request, username }) => {
    test.skip(environment !== 'staging');
    test.setTimeout(6 * timeouts.sixtySeconds);
    await pollDeployment({ location: baseUrl, attempt: 1, maxAttempts: 6, delay: 3 * timeouts.tenSeconds }); // give max 6 * 30s to see a version update, same as current CI job delay
    // wait a little extra to allow rollout to complete if we happen to poll the first replica in a deployment
    await new Promise(resolve => setTimeout(resolve, 3 * timeouts.tenSeconds));
    try {
      const { token } = await login(username, password, baseUrl, request);
      test.skip(!!token, 'looks like the account was created already, continue with the remaining tests');
    } catch {
      // looks like this is the first run, let's continue
    }
    await page.goto(`${baseUrl}ui/`);
    await expect(page.getByText(/Sign up/i)).toBeVisible();
    await page.getByText(/Sign up/i).click();
    console.log(`creating user with username: ${username} and password: ${password}`);
    await expect(page.getByText(/Sign up with/i)).toBeVisible();
    await page.fill(selectors.email, username);
    await page.fill(selectors.password, password);
    await page.fill(selectors.password, '');
    await page.fill(selectors.password, password);
    await page.fill(selectors.passwordConfirmation, password);

    await page.getByRole('button', { name: /Sign up/i }).click();
    await page.getByRole('button', { name: /Complete/i }).waitFor();
    await page.getByLabel(/organization name/i).fill('CI test corp');
    await page.getByLabel(/terms of service/i).check();
    const frameHandle = await page.waitForSelector('iframe[title="reCAPTCHA"]');
    await page.waitForTimeout(300);
    const recaptchaFrame = await frameHandle.contentFrame();
    await recaptchaFrame.waitForSelector('#recaptcha-anchor');
    const recaptcha = await recaptchaFrame.$('#recaptcha-anchor');
    await recaptcha.click();
    await page.getByRole('button', { name: /Complete/i }).click();
    await isLoggedIn(page, timeouts.sixtySeconds); // in some cases the backend takes _a lot_ longer to respond here...
    // the following sets the UI up for easier navigation by disabling onboarding & feedback collection
    const newPage = await prepareNewPage({ baseUrl, context, password, request, username });
    await isLoggedIn(newPage);
    await context.storageState({ path: storagePath });
    await context.close();
  });

  test('OS login', async ({ baseUrl, context, environment, password, request, username }) => {
    test.skip(isEnterpriseOrStaging(environment));
    const newPage = await prepareNewPage({ baseUrl, context, password, request, username });
    await isLoggedIn(newPage);
    await context.storageState({ path: storagePath });
    await context.close();
  });

  test.describe('enterprise setting features', () => {
    test('supports tenant token retrieval, that happens to start up a docker client', async ({
      baseUrl,
      context,
      environment,
      password,
      request,
      username
    }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      console.log(`logging in user with username: ${username} and password: ${password}`);
      const page = await prepareNewPage({ baseUrl, context, password, request, username });
      await page.goto(`${baseUrl}ui/settings`);
      const isVisible = await page.getByRole('button', { name: /change email/i }).isVisible();
      if (!isVisible) {
        console.log('settings may not be loaded - move around');
        await page.goto(`${baseUrl}ui/help`);
        await page.goto(`${baseUrl}ui/settings`);
      }
      const token = await tenantTokenRetrieval(baseUrl, page);
      if (environment === 'staging') {
        await startDockerClient(baseUrl, token);
      }
      await context.storageState({ path: storagePath });
      expect(token).toBeTruthy();
      await context.close();
    });
    test('SP tenant login', async ({ baseUrl, browser, environment, password, request, spTenantUsername }) => {
      if (environment !== 'enterprise') {
        fs.writeFileSync(spStoragePath, JSON.stringify(emptyStorageState));
        test.skip(true, 'only relevant on enterprise setups for now');
      }
      fs.writeFileSync(switchTenantStoragePath, JSON.stringify(emptyStorageState));
      const page = await prepareNewPage({ baseUrl, browser, password, request, username: spTenantUsername });
      await isLoggedIn(page);
      await page.context().storageState({ path: spStoragePath });
      await page.context().close();
    });
  });
});
