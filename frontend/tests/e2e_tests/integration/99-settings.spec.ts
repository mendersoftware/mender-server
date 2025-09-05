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
import { emptyStorageState, selectors, storagePath, timeouts } from '../utils/constants.ts';

test.describe('Settings', () => {
  test.describe('access token feature', () => {
    test('allows access to access tokens', async ({ baseUrl, page }) => {
      await page.goto(`${baseUrl}ui/settings`);
      const tokenGenerationButton = await page.getByRole('button', { name: /Generate a token/i });
      if (!(await tokenGenerationButton.isVisible())) {
        console.log('settings may not be loaded - move around');
        await page.goto(`${baseUrl}ui/help`);
        await page.goto(`${baseUrl}ui/settings`);
      }
      await tokenGenerationButton.waitFor();
    });
    test('allows generating & revoking tokens', async ({ baseUrl, browserName, page }, { retry }) => {
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
      const tokenName = `aNewToken-${retry}`;
      await page.getByText(/Create new token/i).waitFor();
      await page.getByPlaceholder('Name').fill(tokenName);
      await page.getByText(/a year/i).click({ force: true });
      await page.getByRole('option', { name: '7 days' }).click();
      await page.getByRole('button', { name: /Create token/i }).click();
      await page.getByRole('button', { name: /Close/i }).click();
      await page.mouse.wheel(0, 200);
      await page
        .getByText(/in 7 days/i)
        .first()
        .waitFor();
      await page.getByRole('button', { name: /Revoke/i }).click();
      await revokeTokenButton.waitFor();
      await revokeTokenButton.click();
      await tokenGenerationButton.click();
      await page.getByPlaceholder(/Name/i).fill(tokenName);
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

  test.describe('2FA setup', () => {
    test('supports regular 2fa setup', async ({ baseUrl, environment, page }) => {
      test.skip(environment !== 'staging');
      let tfaSecret;
      try {
        tfaSecret = fs.readFileSync('secret.txt', 'utf8');
      } catch {
        // moving on
      }
      test.skip(tfaSecret, 'looks like the account is already 2fa enabled, continue with the remaining tests');
      await page.goto(`${baseUrl}ui/settings/my-profile`);
      await page.getByText(/Enable Two Factor/).click();
      await page.waitForSelector('.margin-top img');
      const qrCode = await page.$eval('.margin-top img', (el: HTMLImageElement) => el.src);
      const png = PNG.sync.read(Buffer.from(qrCode.slice('data:image/png;base64,'.length), 'base64'));
      const decodedQr = jsQR(png.data, png.width, png.height);
      const qrData = new URLSearchParams(decodedQr.data);
      console.log(qrData.get('secret'));
      const qrToken = await generateOtp(qrData.get('secret'));
      console.log('Generated otp:', qrToken);
      await page.getByLabel(/Verification code/i).fill(qrToken);
      await page.getByRole('button', { name: /Verify/i }).click();
      await page.waitForSelector(`css=ol >> text=Verified`);
      await page.getByRole('button', { name: /save/i }).click();
      await page.waitForTimeout(timeouts.default);
    });
    test(`prevents from logging in without 2fa code`, async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(environment !== 'staging');
      let context = await browser.newContext({ storageState: { ...emptyStorageState } });
      const domain = baseUrlToDomain(baseUrl);
      context = await prepareCookies(context, domain, '');
      const page = await context.newPage();
      await page.goto(`${baseUrl}ui/`);
      await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
      // enter valid username and password
      await processLoginForm({ username, password, page, environment });
      await page.waitForTimeout(timeouts.default);
      await page.getByLabel(/Two Factor Authentication Code/i).fill('123456');
      await page.getByRole('button', { name: /next/i }).click();
      // still on /login page plus an error is displayed
      await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
      await page.getByText(/Incorrect email address/).waitFor({ timeout: timeouts.default });
      await context.close();
    });
    test('allows turning 2fa off again', async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(environment !== 'staging');
      let context = await browser.newContext({ storageState: { ...emptyStorageState } });
      const domain = baseUrlToDomain(baseUrl);
      context = await prepareCookies(context, domain, '');
      const page = await context.newPage();
      await page.goto(`${baseUrl}ui/`);
      await processLoginForm({ username, password, page, environment });
      const newToken = await generateOtp();
      await page.getByLabel(/Two Factor Authentication Code/i).fill(newToken);
      await page.getByRole('button', { name: /next/i }).click();
      await isLoggedIn(page);
      await page.goto(`${baseUrl}ui/settings/my-profile`);
      await page.getByText(/Enable Two Factor/).click();
      const failureNotification = await page.getByText(/There was an error disabling/i);
      await expect(failureNotification).not.toBeVisible();
      await page.waitForTimeout(timeouts.default);
      await context.close();
    });
    test('allows logging in without 2fa after deactivation', async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(environment !== 'staging');
      let context = await browser.newContext({ storageState: { ...emptyStorageState } });
      const domain = baseUrlToDomain(baseUrl);
      context = await prepareCookies(context, domain, '');
      const page = await context.newPage();
      await page.goto(`${baseUrl}ui/`);
      await processLoginForm({ username, password, page, environment });
      await isLoggedIn(page);
      await page.goto(`${baseUrl}ui/settings`);
      await context.close();
    });
  });

  test.describe('Basic setting features', () => {
    const replacementPassword = 'mysecretpassword!456';

    test('allows access to user management', async ({ baseUrl, page }) => {
      await page.goto(`${baseUrl}ui/settings/user-management`);
      const userCreationButton = await page.getByRole('button', { name: /Add new user/i });
      if (!(await userCreationButton.isVisible())) {
        console.log('settings may not be loaded - move around');
        await page.goto(`${baseUrl}ui/help`);
        await page.goto(`${baseUrl}ui/settings/user-management`);
      }
      await userCreationButton.waitFor();
    });
    test('allows email changes', async ({ baseUrl, page }) => {
      await page.goto(`${baseUrl}ui/settings/my-profile`);
      await page.getByRole('button', { name: /change email/i }).click();
      await expect(page.getByLabel(/current password/i)).toBeVisible();
    });
    test('allows changing the password', async ({ browserName, page, username, password }) => {
      test.skip(browserName === 'webkit');
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
      await page.getByRole('button', { name: /next/i }).waitFor({ timeout: timeouts.fiveSeconds });
      await expect(page.getByRole('button', { name: /next/i })).toBeVisible();
    });

    test('allows changing the password back', async ({ baseUrl, browserName, browser, password, request, username }) => {
      test.skip(browserName === 'webkit');
      const page = await prepareNewPage({ baseUrl, browser, password: replacementPassword, request, username });
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
      await page.context().storageState({ path: storagePath });
      const { token: newToken } = await login(username, password, baseUrl, request);
      expect(newToken).toBeTruthy();
      await page.context().close();
    });
  });

  test.describe('account upgrades', () => {
    test.beforeEach(({ environment }) => {
      test.skip(environment !== 'staging');
    });
    test('allows subscribing to Basic', async ({ page }) => {
      const wasUpgraded = await page.isVisible(`css=#limit >> text=250`);
      test.skip(wasUpgraded, 'looks like the account was upgraded already, continue with the remaining tests');
      await page.getByText('Upgrade now').click();

      const deviceInput = page.getByRole('spinbutton', { name: 'Number of devices' });
      await deviceInput.focus();
      // Increase by 2 steps (50 => 150)
      await page.keyboard.press('ArrowUp');
      await page.keyboard.press('ArrowUp');

      await page.getByRole('button', { name: 'Upgrade now' }).click();

      await page.getByRole('textbox', { name: /address line 1/i }).fill('Blindernveien');
      await page.getByRole('textbox', { name: /state/i }).fill('Oslo');
      await page.getByRole('textbox', { name: /city/i }).fill('Oslo');
      await page.getByRole('textbox', { name: /zip or postal code/i }).fill('12345');
      await page.getByLabel('Country').fill('Norw');
      await page.getByRole('option', { name: 'Norway' }).click();
      await page.getByRole('button', { name: 'Save Billing details' }).click();

      await page.waitForSelector('.StripeElement iframe');
      const frameHandle = await page.$('.StripeElement iframe');
      const stripeFrame = await frameHandle.contentFrame();
      await stripeFrame.fill('[name="cardnumber"]', '4242424242424242');
      await stripeFrame.fill('[name="exp-date"]', '1232');
      await stripeFrame.fill('[name="cvc"]', '123');
      await stripeFrame.fill('[name="postal"]', '12345');
      await page.getByRole('button', { name: /Confirm subscription/i }).click();
      await page.getByText(/Card confirmed./i).waitFor({ timeout: timeouts.tenSeconds });
      await page.getByText(/ You have successfully subscribed to the basic/i).waitFor({ timeout: timeouts.fifteenSeconds });
      await page.waitForTimeout(timeouts.default); // the tenant state seems to not be populated right away, so the explicit wait to increase chances of the follow up test succeeding
    });

    test('allows upgrading to Professional', async ({ baseUrl, browser, password, request, username }) => {
      const page = await prepareNewPage({ baseUrl, browser, password, request, username });
      const wasUpgraded = await page.isVisible(`css=#limit >> text=350`);
      test.skip(wasUpgraded, 'looks like the account was upgraded already, continue with the remaining tests');
      await page.goto(`${baseUrl}ui/subscription`);

      const deviceNumberInput = page.getByRole('spinbutton', { name: 'Number of devices' });
      await deviceNumberInput.fill('310');
      await page.waitForTimeout(timeouts.default);
      await expect(deviceNumberInput).toHaveValue('350');

      await page.getByRole('radio', { name: 'Professional' }).click();

      await page.waitForTimeout(timeouts.default);

      await page.getByRole('checkbox', { name: 'Troubleshoot' }).click();

      await page.getByRole('button', { name: 'Upgrade now' }).click();
      await page.waitForTimeout(timeouts.default);

      await expect(page.getByRole('heading', { name: '$777' })).toBeVisible();
      await page.getByRole('button', { name: /Confirm subscription/i }).click();

      await page.getByText(/ You have successfully subscribed to the professional/i).waitFor({ timeout: timeouts.fifteenSeconds });
      await page.context().close();
    });
    test('allows higher device limits once upgraded', async ({ baseUrl, browser, password, request, username }) => {
      const page = await prepareNewPage({ baseUrl, browser, password, request, username });
      await page.waitForSelector(`css=#limit >> text=350`, { timeout: timeouts.default });
      await expect(page.locator(`css=#limit >> text=350`)).toBeVisible();
      const token = await tenantTokenRetrieval(baseUrl, page);
      await startClient(baseUrl, token, 50);
      await page.goto(`${baseUrl}ui/devices`);
      await page.getByRole('link', { name: /pending/i }).waitFor({ timeout: timeouts.sixtySeconds });
      await expect(async () => {
        const pendingNotification = await page.getByRole('link', { name: /pending/i }).innerText();
        expect(Number(pendingNotification.split(' ')[0])).toBeGreaterThan(10);
      }).toPass({ timeout: timeouts.sixtySeconds });
      await page.context().close();
    });
    test('allows billing profile editing', async ({ baseUrl, browser, password, request, username }) => {
      const page = await prepareNewPage({ baseUrl, browser, password, request, username });
      await page.goto(`${baseUrl}ui/settings/billing`);
      await page.getByRole('button', { name: /edit/i }).click();
      await page.getByRole('textbox', { name: /address line 1/i }).fill('Gaustadalleen 12');
      await page.getByRole('textbox', { name: /state/i }).fill('Moss');
      await page.getByRole('textbox', { name: /city/i }).fill('Moss');
      await page.getByRole('textbox', { name: /zip or postal code/i }).fill('54321');
      await page.getByLabel('Country').fill('Pol');
      await page.getByRole('option', { name: 'Poland' }).click();
      await page.getByRole('button', { name: /edit/i }).click();

      await page.waitForSelector('.StripeElement iframe');
      const frameHandle = await page.$('.StripeElement iframe');
      const stripeFrame = await frameHandle.contentFrame();
      await stripeFrame.fill('[name="cardnumber"]', '4242424242424242');
      await stripeFrame.fill('[name="exp-date"]', '0134');
      await stripeFrame.fill('[name="cvc"]', '333');
      await stripeFrame.fill('[name="postal"]', '02040');
      await page.getByRole('button', { name: /save/i }).click();
      await expect(page.getByText('Gaustadalleen 12')).toBeVisible();
      await page.context().close();
    });
  });
});
