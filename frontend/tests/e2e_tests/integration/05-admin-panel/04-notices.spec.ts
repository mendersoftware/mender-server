// Copyright 2026 Northern.tech AS
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
import { adminPanelApiUrl } from '../../utils/adminPanel';
import { timeouts } from '../../utils/constants';
import type { Email, EmailClient } from '../../utils/email';
import { setupEmailClient } from '../../utils/email';

// deliberately markdown flavoured: the pipeline sends text/plain, so this has to survive as written
const noticeBody = '# Retiring the widget API\n\n- shuts down on 2026-12-01\n- migrate to the gadget API instead';
const noticeHeadline = 'Retiring the widget API';

// the framing `applyDefaults` fills in when the operator leaves the optional fields blank
const defaultTitle = 'Notice';
const defaultGreeting = 'Hello,';

/**
 * The workflow bcc's its recipients, so the `To` header the mail server reports stays empty and
 * filtering by recipient finds nothing - matching a unique subject is the reliable way in.
 */
const findBySubject = async (emailClient: EmailClient, subject: string): Promise<Email | undefined> => {
  const emails = await emailClient.getEmails();
  return emails.find(email => email.subject === subject);
};

const awaitNotice = async (emailClient: EmailClient, subject: string): Promise<Email> => {
  await expect
    .poll(async () => !!(await findBySubject(emailClient, subject)), {
      message: `expected a notice with subject "${subject}" to be delivered`,
      timeout: timeouts.fifteenSeconds
    })
    .toBeTruthy();
  const notice = await findBySubject(emailClient, subject);
  if (!notice) {
    throw new Error(`the notice "${subject}" vanished between being polled for and being read`);
  }
  return notice;
};

test.describe('Admin panel notices', () => {
  test.beforeEach(async ({ environment }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
  });

  test.describe('composing', () => {
    test('previews what the recipients will see', async ({ adminBaseUrl, page }) => {
      await page.goto(`${adminBaseUrl}notices`);
      await expect(page.getByRole('heading', { name: 'Send Notice' })).toBeVisible();
      await expect(page.getByText(/click .*preview rendered email.* to see how it will look/i)).toBeVisible();

      await page.getByLabel(/feature details/i).fill(noticeBody);
      await page.getByRole('button', { name: /preview rendered email/i }).click();

      const preview = page.getByText('Rendered preview').locator('..');
      await expect(preview.getByText(noticeHeadline)).toBeVisible();
      // the framing is filled in for the operator, the body itself is left untouched
      await expect(preview.getByText(defaultGreeting)).toBeVisible();
      await expect(preview.getByText(defaultTitle)).toBeVisible();
    });

    test('asks for a body before previewing anything', async ({ adminBaseUrl, page }) => {
      await page.goto(`${adminBaseUrl}notices`);
      await page.getByRole('button', { name: /preview rendered email/i }).click();
      await expect(page.getByText('Message body is required.')).toBeVisible();
      await expect(page.getByText('Rendered preview').locator('..').getByText(defaultGreeting)).toHaveCount(0);
    });

    test('lets the operator take over the framing', async ({ adminBaseUrl, page }) => {
      const customTitle = 'Scheduled maintenance';
      await page.goto(`${adminBaseUrl}notices`);
      await page.getByLabel(/feature details/i).fill(noticeBody);

      // the defaults are good enough most of the time, so the overrides start out collapsed
      await expect(page.getByLabel('Email subject')).not.toBeVisible();
      await page.getByText(/customize subject, title, intro/i).click();
      await page.getByLabel('Title (H1 in the email)').fill(customTitle);

      await page.getByRole('button', { name: /preview rendered email/i }).click();
      const preview = page.getByText('Rendered preview').locator('..');
      await expect(preview.getByText(customTitle)).toBeVisible();
      await expect(preview.getByText(defaultTitle)).not.toBeVisible();
    });

    test('will not broadcast until the confirmation is typed out', async ({ adminBaseUrl, page }) => {
      await page.goto(`${adminBaseUrl}notices`);
      await page.getByLabel(/feature details/i).fill(noticeBody);
      await page.getByRole('button', { name: /send to all recipients/i }).click();
      // the confirmation is all that stands between an operator and every tenant's inbox
      await expect(page.getByText('Type SEND in the confirmation box to proceed.')).toBeVisible();
    });
  });
  test.describe('sending', () => {
    test('refuses a notice without a body', async ({ adminBaseUrl, request }) => {
      const response = await request.post(adminPanelApiUrl(adminBaseUrl, 'notices/preview'), { data: { body: '' }, failOnStatusCode: false });
      expect(response.status()).toEqual(400);
    });

    test('refuses a test send without a recipient', async ({ adminBaseUrl, request }) => {
      const response = await request.post(adminPanelApiUrl(adminBaseUrl, 'notices/send'), {
        data: { action: 'test', body: noticeBody },
        failOnStatusCode: false
      });
      expect(response.status()).toEqual(400);
    });

    test('refuses a broadcast without the confirmation', async ({ adminBaseUrl, request }) => {
      const response = await request.post(adminPanelApiUrl(adminBaseUrl, 'notices/send'), {
        data: { action: 'send_all', audience: 'admins', body: noticeBody },
        failOnStatusCode: false
      });
      expect(response.status()).toEqual(400);
    });
    test('delivers a test notice as plain text to the named address', async ({ adminBaseUrl, environment, page, username }) => {
      const emailClient = setupEmailClient(username, environment);
      test.skip(!emailClient, 'test requires a reachable mailbox');
      const subject = `e2e admin panel test notice ${Date.now()}`;
      const testEmail = 'notice-recipient@example.com';

      await page.goto(`${adminBaseUrl}notices`);
      await page.getByLabel(/feature details/i).fill(noticeBody);
      await page.getByText(/customize subject, title, intro/i).click();
      await page.getByLabel('Email subject').fill(subject);
      await page.getByLabel('Test recipient').fill(testEmail);
      await page.getByRole('button', { name: /send test message/i }).click();
      await expect(page.getByText(`Test message sent to ${testEmail}.`)).toBeVisible({ timeout: timeouts.tenSeconds });

      const notice = await awaitNotice(emailClient, subject);
      // what the preview promised is what actually lands in the inbox
      expect(notice.body).toContain(noticeHeadline);
      expect(notice.body).toContain('migrate to the gadget API instead');
      expect(notice.body).not.toContain('<h1>');
    });

    test('broadcasts to the tenant admins across tenant boundaries', async ({ adminBaseUrl, environment, request, username }) => {
      const emailClient = setupEmailClient(username, environment);
      test.skip(!emailClient, 'test requires a reachable mailbox');
      const subject = `e2e admin panel broadcast ${Date.now()}`;

      const response = await request.post(adminPanelApiUrl(adminBaseUrl, 'notices/send'), {
        data: { action: 'send_all', audience: 'admins', body: noticeBody, confirm: 'SEND', subject }
      });
      expect(response.ok()).toBeTruthy();
      const result = await response.json();
      expect(result.failures ?? []).toEqual([]);
      // the admins of every tenant the runner created get collected into the batch
      expect(result.recipients_count).toBeGreaterThanOrEqual(1);

      const notice = await awaitNotice(emailClient, subject);
      expect(notice.body).toContain(noticeHeadline);
    });
  });
});
