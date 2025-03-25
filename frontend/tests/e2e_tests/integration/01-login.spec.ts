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
import test, { expect } from '../fixtures/fixtures.ts';
import { baseUrlToDomain, isLoggedIn, prepareCookies, processLoginForm } from '../utils/commands.ts';
import { selectors, storagePath, timeouts } from '../utils/constants.ts';

test.describe('Login', () => {
  test.describe('works as expected', () => {
    test('Logs in using UI', async ({ baseUrl, environment, context, page, password, username }) => {
      console.log(`logging in user with username: ${username} and password: ${password}`);
      // enter valid username and password
      await page.goto(`${baseUrl}ui/`);
      await processLoginForm({ username, password, page, environment });
      // confirm we have logged in successfully
      await isLoggedIn(page);
      await page.evaluate(() => localStorage.setItem(`onboardingComplete`, 'true'));
      await context.storageState({ path: storagePath });
    });

    test('does not stay logged in across sessions, after browser restart', async ({ baseUrl, page }) => {
      await page.goto(`${baseUrl}ui/`);
      await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
    });

    test('Logs out using UI', async ({ baseUrl, environment, page, password, username }) => {
      await page.goto(`${baseUrl}ui/`);
      await processLoginForm({ username, password, page, environment });
      // now we can log out
      await page.getByRole('button', { name: username }).click();
      await page.getByText(/log out/i).click();
      await page.getByRole('button', { name: /next/i }).waitFor({ timeout: timeouts.default });
      await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
    });

    test('fails to access unknown resource', async ({ baseUrl, page, request }) => {
      await page.goto(`${baseUrl}ui/`);
      const response = await request.get(`${baseUrl}/users`);
      expect(response.ok()).toBeTruthy();
      await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
    });

    test('Does not log in with invalid password', async ({ baseUrl, environment, page, username }) => {
      console.log(`logging in user with username: ${username} and password: lewrongpassword`);
      await page.goto(`${baseUrl}ui/`);
      // enter valid username and invalid password
      await processLoginForm({ username, password: 'lewrongpassword', page, environment });

      // still on /login page plus an error is displayed
      await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
      await expect(page.getByText('Incorrect email address and / or password')).toBeVisible();
    });

    test('Does not log in without password', async ({ baseUrl, environment, page, username }) => {
      test.skip(environment === 'staging');
      console.log(`logging in user with username: ${username} and without a password`);
      await page.goto(`${baseUrl}ui/`);
      // enter valid username and invalid password
      await page.waitForSelector(selectors.email);
      await page.click(selectors.email);
      await page.fill(selectors.email, username);
      await page.getByRole('button', { name: /next/i }).click();
      await page.waitForTimeout(timeouts.default);
      await page.getByRole('button', { name: /next/i }).click();
      if (environment === 'enterprise') {
        await expect(page.getByText('Incorrect email address and / or password')).toBeVisible();
      }
    });
  });

  test('stays logged in across sessions, after browser restart if selected', async ({ baseUrl, environment, browser, context, password, username }) => {
    console.log(`logging in user with username: ${username} and password: ${password}`);
    const domain = baseUrlToDomain(baseUrl);
    await context.addCookies([{ name: 'cookieconsent_status', value: 'allow', path: '/', domain }]);
    const page = await context.newPage();
    await page.goto(`${baseUrl}ui/`);
    // enter valid username and password
    await processLoginForm({ username, password, page, environment, stayLoggedIn: true });

    // confirm we have logged in successfully
    await isLoggedIn(page);
    await expect(page.getByRole('button', { name: /next/i })).not.toBeVisible();
    await page.getByText(/Releases/i).click();
    await context.storageState({ path: storagePath });
    let differentContext = await browser.newContext({ storageState: storagePath });
    differentContext = await prepareCookies(differentContext, domain, '');
    const differentPage = await differentContext.newPage();
    await differentPage.goto(`${baseUrl}ui/`);
    // page.reload();
    await expect(page.getByRole('button', { name: /next/i })).not.toBeVisible();
    await expect(differentPage.getByText('Getting started')).not.toBeVisible();
  });
});
