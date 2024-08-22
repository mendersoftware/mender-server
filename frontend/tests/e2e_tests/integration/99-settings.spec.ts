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
import * as fs from 'fs';
import jsQR from 'jsqr';
import { PNG } from 'pngjs';

import test, { expect } from '../fixtures/fixtures.ts';
import {
  baseUrlToDomain,
  generateOtp,
  isLoggedIn,
  login,
  prepareCookies,
  prepareNewPage,
  processLoginForm,
  startClient,
  tenantTokenRetrieval
} from '../utils/commands.ts';
import { selectors, storagePath, timeouts } from '../utils/constants.ts';

test.describe('Settings', () => {
  test.describe('access token feature', () => {
    test.use({ storageState: storagePath });
    test('allows access to access tokens', async ({ baseUrl, loggedInPage: page }) => {
      await page.goto(`${baseUrl}ui/settings`);
      const tokenGenerationButton = await page.getByRole('button', { name: /Generate a token/i });
      if (!(await tokenGenerationButton.isVisible())) {
        console.log('settings may not be loaded - move around');
        await page.goto(`${baseUrl}ui/help`);
        await page.goto(`${baseUrl}ui/settings`);
      }
      await tokenGenerationButton.waitFor();
    });
    test('allows generating & revoking tokens', async ({ baseUrl, browserName, loggedInPage: page }) => {
      await page.goto(`${baseUrl}ui/settings`);
      const tokenGenerationButton = await page.getByText(/generate a token/i);
      await tokenGenerationButton.waitFor();
      const revokeButton = await page.getByText(/revoke/i);
      const revokeTokenButton = await page.getByRole('button', { name: /Revoke token/i });
      if (await revokeButton.isVisible()) {
        await revokeButton.click();
        await revokeTokenButton.waitFor();
        await revokeTokenButton.click();
      }
      await tokenGenerationButton.click();
      await page.getByText(/Create new token/i).waitFor();
      await page.getByPlaceholder('Name').fill('aNewToken');
      await page.getByText(/a year/i).click({ force: true });
      await page.getByRole('option', { name: '7 days' }).click();
      await page.getByRole('button', { name: /Create token/i }).click();
      await page.getByRole('button', { name: /Close/i }).click();
      await page.getByText(/in 7 days/i).waitFor();
      await page.getByRole('button', { name: /Revoke/i }).click();
      await revokeTokenButton.waitFor();
      await revokeTokenButton.click();
      await tokenGenerationButton.click();
      await page.getByPlaceholder(/Name/i).fill('aNewToken');
      await page.getByRole('button', { name: /Create token/i }).click();
      await page.click('.code .MuiSvgIcon-root');
      await page.getByText(/copied to clipboard/i).waitFor();
      let token = '';
      if (browserName === 'chromium') {
        token = await page.evaluate(() => navigator.clipboard.readText());
      } else {
        token = await page.innerText('.code');
      }
      expect(token).toBeTruthy();
      await page.getByRole('button', { name: /Close/i }).click();
      await page.getByText(/in a year/i).waitFor();
    });
  });
  test.describe('account upgrades', () => {
    test.use({ storageState: storagePath });
    test('allows upgrading to Professional', async ({ environment, loggedInPage: page }) => {
      test.skip(environment !== 'staging');
      await page.waitForTimeout(timeouts.default);
      const wasUpgraded = await page.isVisible(`css=#limit >> text=250`);
      if (wasUpgraded) {
        test.skip('looks like the account was upgraded already, continue with the remaining tests');
      }
      await page.getByText('Upgrade now').click();
      await page.click(`css=.planPanel >> text=Professional`);
      await page.waitForSelector('.StripeElement iframe');
      const frameHandle = await page.$('.StripeElement iframe');
      const stripeFrame = await frameHandle.contentFrame();
      await stripeFrame.fill('[name="cardnumber"]', '4242424242424242');
      await stripeFrame.fill('[name="exp-date"]', '1232');
      await stripeFrame.fill('[name="cvc"]', '123');
      await stripeFrame.fill('[name="postal"]', '12345');
      await page.click(`button:has-text('Sign up')`);
      await page.getByText(/Card confirmed./i).waitFor({ timeout: timeouts.tenSeconds });
      await page.getByText(/Your upgrade was successful/i).waitFor({ timeout: timeouts.fifteenSeconds });
      await page.getByText(/Organization name/i).waitFor({ timeout: timeouts.tenSeconds });
    });
    test('allows higher device limits once upgraded', async ({ baseUrl, environment, loggedInPage: page }) => {
      test.skip(environment !== 'staging');
      await page.waitForSelector(`css=#limit >> text=250`, { timeout: timeouts.default });
      await expect(page.locator(`css=#limit >> text=250`)).toBeVisible();
      const token = await tenantTokenRetrieval(baseUrl, page);
      await startClient(baseUrl, token, 50);
      await page.goto(`${baseUrl}ui/devices`);
      await page.getByRole('link', { name: /pending/i }).waitFor({ timeout: timeouts.fifteenSeconds });
      const pendingNotification = await page.$eval('.header-section [href="/ui/devices/pending"]', el => el.textContent);
      expect(Number(pendingNotification.split(' ')[0])).toBeGreaterThan(10);
    });
  });

  test.describe('2FA setup', () => {
    test('supports regular 2fa setup', async ({ baseUrl, environment, loggedInPage: page }) => {
      test.skip(environment !== 'staging');
      let tfaSecret;
      try {
        tfaSecret = fs.readFileSync('secret.txt', 'utf8');
      } catch {
        // moving on
      }
      if (tfaSecret) {
        test.skip('looks like the account is already 2fa enabled, continue with the remaining tests');
      }
      await page.goto(`${baseUrl}ui/settings/my-account`);
      await page.getByText(/Enable Two Factor/).click();
      await page.waitForSelector('.margin-top img');
      const qrCode = await page.$eval('.margin-top img', (el: HTMLImageElement) => el.src);
      const png = PNG.sync.read(Buffer.from(qrCode.slice('data:image/png;base64,'.length), 'base64'));
      const decodedQr = jsQR(png.data, png.width, png.height);
      const qrData = new URLSearchParams(decodedQr.data);
      console.log(qrData.get('secret'));
      const qrToken = await generateOtp(qrData.get('secret'));
      console.log('Generated otp:', qrToken);
      await page.fill('#token2fa', qrToken);
      await page.getByRole('button', { name: /Verify/i }).click();
      await page.waitForSelector(`css=ol >> text=Verified`);
      await page.getByRole('button', { name: /save/i }).click();
      await page.waitForTimeout(timeouts.default);
    });
    test(`prevents from logging in without 2fa code`, async ({ baseUrl, environment, page, password, username }) => {
      test.skip(environment !== 'staging');
      await page.goto(`${baseUrl}ui/`);
      await expect(page.getByRole('button', { name: /Log in/i })).toBeVisible();
      // enter valid username and password
      await processLoginForm({ username, password, page, environment });
      await page.waitForTimeout(timeouts.default);
      await page.fill('#token2fa', '123456');
      await page.getByRole('button', { name: /log in/i }).click();
      // still on /login page plus an error is displayed
      await expect(page.getByRole('button', { name: /Log in/i })).toBeVisible();
      await page.getByText(/There was a problem logging in/).waitFor({ timeout: timeouts.default });
    });
    test('allows turning 2fa off again', async ({ baseUrl, environment, page, password, username }) => {
      test.skip(environment !== 'staging');
      await page.goto(`${baseUrl}ui/`);
      await processLoginForm({ username, password, page, environment });
      const newToken = await generateOtp();
      await page.fill('#token2fa', newToken);
      await page.getByRole('button', { name: /log in/i }).click();
      await isLoggedIn(page);
      await page.goto(`${baseUrl}ui/settings/my-account`);
      await page.getByText(/Enable Two Factor/).click();
      await page.waitForTimeout(timeouts.default);
    });
    test('allows logging in without 2fa after deactivation', async ({ baseUrl, environment, page, password, username }) => {
      test.skip(environment !== 'staging');
      await page.goto(`${baseUrl}ui/`);
      await processLoginForm({ username, password, page, environment });
      await isLoggedIn(page);
      await page.goto(`${baseUrl}ui/settings`);
    });
  });

  test.describe('Basic setting features', () => {
    const replacementPassword = 'mysecretpassword!456';

    test('allows access to user management', async ({ baseUrl, loggedInPage: page }) => {
      await page.goto(`${baseUrl}ui/settings/user-management`);
      const userCreationButton = await page.getByRole('button', { name: /Add new user/i });
      if (!(await userCreationButton.isVisible())) {
        console.log('settings may not be loaded - move around');
        await page.goto(`${baseUrl}ui/help`);
        await page.goto(`${baseUrl}ui/settings/user-management`);
      }
      await userCreationButton.waitFor();
    });
    test('allows email changes', async ({ baseUrl, loggedInPage: page }) => {
      await page.goto(`${baseUrl}ui/settings/my-account`);
      await page.getByRole('button', { name: /change email/i }).click();
      await expect(page.getByLabel(/current password/i)).toBeVisible();
    });
    test('allows changing the password', async ({ baseUrl, browserName, context, environment, username, password }) => {
      test.skip(browserName === 'webkit');
      const domain = baseUrlToDomain(baseUrl);
      context = await prepareCookies(context, domain, '');
      const page = await context.newPage();
      await page.goto(`${baseUrl}ui/`);
      await processLoginForm({ username, password, page, environment });
      await page.getByRole('button', { name: username }).click();
      await page.getByText(/my profile/i).click();
      await page.getByRole('button', { name: /change password/i }).click();

      expect(await page.$eval(selectors.password, (el: HTMLInputElement) => el.value)).toBeFalsy();
      await page.getByRole('button', { exact: true, name: 'Generate' }).click();
      await page.click(selectors.passwordCurrent, { clickCount: 3 });
      await page.fill(selectors.passwordCurrent, password);
      const typedCurrentPassword = await page.$eval(selectors.passwordCurrent, (el: HTMLInputElement) => el.value);
      expect(typedCurrentPassword === password);
      expect(await page.$eval(selectors.password, (el: HTMLInputElement) => el.value)).toBeTruthy();
      await page.click(selectors.password, { clickCount: 3 });
      await page.fill(selectors.password, replacementPassword);
      const typedPassword = await page.$eval(selectors.password, (el: HTMLInputElement) => el.value);
      expect(typedPassword === replacementPassword);
      await page.fill(selectors.passwordConfirmation, replacementPassword);
      await page.getByRole('button', { name: /save/i }).click();
      await page.getByText(/user has been updated/i).waitFor({ timeout: timeouts.tenSeconds });
      await page.getByRole('button', { name: username }).click();
      await page.getByText(/log out/i).click();
      await page.waitForTimeout(timeouts.default);
      await page.screenshot({ path: './test-results/logout.png' });
      await page.getByRole('button', { name: /log in/i }).waitFor({ timeout: timeouts.fiveSeconds });
      await expect(page.getByRole('button', { name: /log in/i })).toBeVisible();
    });

    test('allows changing the password back', async ({ baseUrl, browserName, browser, password, username }) => {
      test.skip(browserName === 'webkit');
      const page = await prepareNewPage({ baseUrl, browser, password: replacementPassword, username });
      await page.getByRole('button', { name: username }).click();
      await page.getByText(/my profile/i).click();
      await page.getByRole('button', { name: /change password/i }).click();
      await page.fill(selectors.password, password);
      const typedPassword = await page.$eval(selectors.password, (el: HTMLInputElement) => el.value);
      if (typedPassword !== password) {
        await page.click(selectors.password, { clickCount: 3 });
        await page.fill(selectors.password, password);
      }
      await page.click(selectors.passwordConfirmation);
      await page.fill(selectors.passwordConfirmation, password);
      await page.click(selectors.passwordCurrent);
      await page.fill(selectors.passwordCurrent, replacementPassword);
      await page.getByRole('button', { name: /save/i }).click();
      await page.getByText(/user has been updated/i).waitFor({ timeout: timeouts.tenSeconds });
      await page.waitForTimeout(timeouts.default);

      const { token: newToken } = await login(username, password, baseUrl);
      expect(newToken).toBeTruthy();
    });
  });

  test.describe('Multi tenant access', () => {
    const secondaryUser = 'demo-secondary@example.com';
    test('allows adding users to tenants', async ({ baseUrl, browser, browserName, environment, loggedInPage, password }) => {
      test.skip('enterprise' !== environment || browserName !== 'chromium');
      await loggedInPage.goto(`${baseUrl}ui/settings`);
      await loggedInPage.click('text=/user management/i');
      const hasUserAlready = await loggedInPage.getByText(secondaryUser).isVisible();
      test.skip(hasUserAlready, `${secondaryUser} was added in a previous run, but success notification wasn't caught`);
      const page = await prepareNewPage({ baseUrl, browser, username: secondaryUser, password });
      await page.goto(`${baseUrl}ui/settings/my-account`);
      await page
        .getByText(/User ID/i)
        .locator('..')
        .locator('..')
        .getByRole('button', { name: /copy to clipboard/i })
        .click({ force: true });
      const content = await page.evaluateHandle(() => navigator.clipboard.readText());
      const uuid = await content.jsonValue();
      await page.getByText(/help/i).click();
      await page.getByRole('button', { name: secondaryUser }).click();
      await expect(page.getByText(/switch organization/i)).not.toBeVisible();

      await loggedInPage.getByRole('button', { name: /new user/i }).click();
      const passwordInput = await loggedInPage.getByPlaceholder(/password/i);
      const emailUuidInput = await loggedInPage.getByPlaceholder(/email/i);
      await emailUuidInput.click();
      await emailUuidInput.fill(uuid);
      await expect(passwordInput).not.toBeVisible();
      await loggedInPage.getByRole('button', { name: /add user/i }).click();
      await page.waitForTimeout(timeouts.oneSecond);

      await page.reload();
      await page.getByRole('button', { name: secondaryUser }).click();
      await expect(page.getByText(/switch organization/i)).toBeVisible();
    });
    test('allows switching tenants', async ({ baseUrl, browser, browserName, environment, loggedInPage, password }) => {
      test.skip('enterprise' !== environment || browserName !== 'chromium');
      // here we can't use prepareNewPage as it sets the initial JWT to be used on every page init
      const domain = baseUrlToDomain(baseUrl);
      let newContext = await browser.newContext();
      newContext = await prepareCookies(newContext, domain, '');
      const page = await newContext.newPage();
      await page.goto(`${baseUrl}ui/`);
      await processLoginForm({ username: secondaryUser, password, page, environment });
      await page.getByRole('button', { name: secondaryUser }).click();
      await expect(page.getByRole('menuitem', { name: /secondary/i })).toBeVisible();
      await page.getByText(/switch organization/i).click({ force: true });
      const tenantSwitch = await page.getByRole('menuitem', { name: /test/i });
      await tenantSwitch.waitFor({ timeout: timeouts.default });
      await tenantSwitch.click();
      await page.waitForTimeout(timeouts.default);
      await page.getByRole('button', { name: secondaryUser }).click();
      await expect(page.getByRole('menuitem', { name: /secondary/i })).not.toBeVisible();
      await expect(page.getByRole('menuitem', { name: /test/i })).toBeVisible();

      await loggedInPage.goto(`${baseUrl}ui/settings`);
      await loggedInPage.click('text=/user management/i');
      await loggedInPage.getByText(secondaryUser).click();
      await loggedInPage.getByRole('button', { name: /delete user/i }).click();
      await expect(loggedInPage.getByText(/delete user\?/i)).toBeVisible();
      await loggedInPage
        .getByRole('button', { name: /delete user/i })
        .last()
        .click();

      await page.reload();
      await expect(page.getByText(/log in/i)).toBeVisible();
      await processLoginForm({ username: secondaryUser, password, page, environment });
      await page.getByRole('button', { name: secondaryUser }).click();
      await expect(page.getByText(/switch organization/i)).not.toBeVisible();
    });
  });
});
