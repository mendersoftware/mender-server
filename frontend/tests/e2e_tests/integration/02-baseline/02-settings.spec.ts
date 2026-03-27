// Copyright 2025 Northern.tech AS
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
import test, { expect } from '../../fixtures/fixtures';
import { timeouts } from '../../utils/constants';
import { setupEmailClient } from '../../utils/email';
import { poll } from '../../utils/utils';

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
      const copyButton = page.getByRole('button', { name: /copy to clipboard/i });
      await copyButton.click();
      await page.getByText(/copied to clipboard/i).waitFor();
      let token = '';
      if (browserName === 'chromium') {
        token = await page.evaluate(() => navigator.clipboard.readText());
      } else {
        token = await copyButton.locator('..').locator('.copyable-content').innerText();
      }
      expect(token).toBeTruthy();
      await page.getByRole('button', { name: /Close/i }).click();
      await page.getByText(/in a year/i).waitFor();
    });
  });

  test.describe('Basic setting features', () => {
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

    test('verify email address', async ({ baseUrl, environment, page, username }) => {
      test.skip(environment !== 'staging' && environment !== 'enterprise');
      const emailClient = setupEmailClient(username, environment);
      test.skip(!emailClient, 'test requires configuring a mailbox');
      await page.goto(`${baseUrl}ui/settings/my-profile`);

      test.skip(!(await page.getByText('Not verified').isVisible()), 'email is already verified');

      await page.getByRole('button', { name: /Verify/i }).click();

      // Wait five seconds for the email to be sent
      await new Promise(resolve => setTimeout(resolve, timeouts.oneSecond * 3));
      const emails = await poll({
        callback: () => emailClient.getEmails({ to: username, unread: true }),
        condition: emails => emails.length > 0,
        delay: timeouts.oneSecond / 2,
        message: 'timeout waiting for verification email'
      });
      const match = emails[0].body.match(new RegExp(`${baseUrl.replace(/\/$/, '')}/ui/activate/[0-9a-f-]+`));
      test.expect(match, 'found verification URL in email').not.toBeNull();
      await page.goto(match[0]);

      await page.getByText('Email verified').waitFor({ timeout: timeouts.default }); // Activate landing page (success)
      await page.goto(`${baseUrl}ui/settings/my-profile`);
      await page.getByText(/^Verified$/).waitFor({ timeout: timeouts.default }); // "Verified" chip next to email
    });

    test('allows email changes', async ({ baseUrl, page }) => {
      await page.goto(`${baseUrl}ui/settings/my-profile`);
      await page.getByRole('button', { name: /change email/i }).click();
      await expect(page.getByLabel(/current password/i)).toBeVisible();
    });
  });
});
