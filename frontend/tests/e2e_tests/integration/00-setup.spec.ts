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

import test from '../fixtures/fixtures.ts';
import { isEnterpriseOrStaging, isLoggedIn, login, prepareNewPage, startDockerClient, stopDockerClient, tenantTokenRetrieval } from '../utils/commands.ts';
import { selectors, storagePath, timeouts } from '../utils/constants.ts';

test.describe('Test setup', () => {
  test.describe('basic window checks', () => {
    test('get the global window object', async ({ baseUrl, context, page }) => {
      page = await context.newPage();
      await page.goto(`${baseUrl}ui/`);
      const theWindow = await page.evaluate(() => window.innerWidth);
      expect(theWindow).toBeDefined();
      try {
        fs.unlinkSync('loginInfo.json');
        await stopDockerClient();
      } catch {
        // ...continue
      }
    });
    test('get the document object', async ({ page }) => {
      const documentCharset = await page.evaluate(() => document.charset);
      expect(documentCharset).toBeDefined();
      expect(documentCharset).toEqual('UTF-8');
    });
    test('get the title', async ({ baseUrl, context, page }) => {
      page = await context.newPage();
      await page.goto(`${baseUrl}ui/`);
      await expect(page).toHaveTitle(/Mender/i);
    });
  });

  test.describe('account creation', () => {
    test('allows account creation', async ({ baseUrl, context, environment, page, password, username }) => {
      test.skip(environment !== 'staging');
      try {
        const { token } = await login(username, password, baseUrl);
        test.skip(!!token, 'looks like the account was created already, continue with the remaining tests');
      } catch {
        // looks like this is the first run, let's continue
      }
      await page.goto(`${baseUrl}ui/`);
      await expect(page.getByText(/Sign up/i)).toBeVisible();
      await page.getByText(/Sign up/i).click();
      console.log(`creating user with username: ${username} and password: ${password}`);
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
      await page.waitForTimeout(timeouts.default);
      await page.getByRole('button', { name: /Complete/i }).click();
      await isLoggedIn(page, timeouts.fifteenSeconds);

      // the following sets the UI up for easier navigation by disabling onboarding
      const newPage = await prepareNewPage({ baseUrl, context, password, username });
      await isLoggedIn(newPage);
      await context.storageState({ path: storagePath });
    });
  });

  test.describe('enterprise setting features, that happens to start up a docker client', () => {
    test('supports tenant token retrieval', async ({ baseUrl, context, environment, password, username }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      console.log(`logging in user with username: ${username} and password: ${password}`);
      const page = await prepareNewPage({ baseUrl, context, password, username });
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
    });
  });
});
