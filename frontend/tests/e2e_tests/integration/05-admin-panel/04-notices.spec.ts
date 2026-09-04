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
import { adminPanelApiUrl, noticePreview } from '../../utils/adminPanel';
import { timeouts } from '../../utils/constants';
import type { Email, EmailClient } from '../../utils/email';
import { setupEmailClient } from '../../utils/email';

// deliberately markdown flavoured: the pipeline sends text/plain, so this has to survive as written
const noticeBody = '# Retiring the widget API\n\n- shuts down on 2026-12-01\n- migrate to the gadget API instead';
const noticeHeadline = 'Retiring the widget API';

// the framing `applyDefaults` fills in when the operator leaves the optional fields blank
const defaultTitle = 'Notice';
const defaultGreeting = 'Hello,';

/** `test.skip` aborts the test, which the type checker cannot see - hence the assertion */
const requireEmailClient = (username: string, environment: string): EmailClient => {
  const emailClient = setupEmailClient(username, environment);
  test.skip(!emailClient, 'test requires a reachable mailbox');
  return emailClient as EmailClient;
};

/**
 * The workflow bcc's its recipients, so the `To` header the mail server reports stays empty and
 * filtering by recipient finds nothing - matching a unique subject is the reliable way in.
 */
const awaitNotice = async (emailClient: EmailClient, subject: string): Promise<Email> => {
  let notice: Email | undefined;
  await expect
    .poll(
      async () => {
        const emails = await emailClient.getEmails();
        notice = emails.find(email => email.subject === subject);
        return !!notice;
      },
      {
        message: `expected a notice with subject "${subject}" to be delivered`,
        timeout: timeouts.fifteenSeconds
      }
    )
    .toBeTruthy();
  return notice!;
};

test.describe('Admin panel notices', () => {
  test.beforeEach(async ({ environment }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
  });

  test.describe('composing', () => {
    test('previews what the recipients will see, once there is something to preview', async ({ adminBaseUrl, page }) => {
      await page.goto(`${adminBaseUrl}notices`);
      await expect(page.getByRole('heading', { name: 'Send Notice' })).toBeVisible();
      await expect(page.getByText(/click .*preview rendered email.* to see how it will look/i)).toBeVisible();

      await page.getByRole('button', { name: /preview rendered email/i }).click();
      await expect(page.getByText('Message body is required.')).toBeVisible();
      await expect(noticePreview(page).getByText(defaultGreeting)).toHaveCount(0);

      await page.getByLabel(/feature details/i).fill(noticeBody);
      await page.getByRole('button', { name: /preview rendered email/i }).click();

      const preview = noticePreview(page);
      await expect(preview.getByText(noticeHeadline)).toBeVisible();
      // the framing is filled in for the operator, the body itself is left untouched
      await expect(preview.getByText(defaultGreeting)).toBeVisible();
      await expect(preview.getByText(defaultTitle)).toBeVisible();
    });

    test('lets the operator take over the framing', async ({ adminBaseUrl, page }) => {
      const customTitle = 'Scheduled maintenance';
      await page.goto(`${adminBaseUrl}notices`);
      await page.getByLabel(/feature details/i).fill(noticeBody);

      // the defaults are good enough most of the time, so the overrides start out collapsed
      await expect(page.getByLabel('Email subject')).not.toBeVisible();
      await page.getByRole('button', { name: /customize subject, title, intro/i }).click();
      await page.getByLabel('Title (H1 in the email)').fill(customTitle);

      await page.getByRole('button', { name: /preview rendered email/i }).click();
      const preview = noticePreview(page);
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

  // What is left here is delivery: the request rejections the panel and its API perform are covered
  // by the admin-panel service's own handler tests, what nothing else covers is whether a notice
  // makes it through workflows into an actual inbox.
  test.describe('sending', () => {
    test('delivers a test notice as plain text to the named address', async ({ adminBaseUrl, environment, page, username }) => {
      const emailClient = requireEmailClient(username, environment);
      const subject = `e2e admin panel test notice ${Date.now()}`;
      const testEmail = 'notice-recipient@example.com';

      await page.goto(`${adminBaseUrl}notices`);
      await page.getByLabel(/feature details/i).fill(noticeBody);
      await page.getByRole('button', { name: /customize subject, title, intro/i }).click();
      await page.getByLabel('Email subject').fill(subject);
      await page.getByLabel('Test recipient').fill(testEmail);
      await page.getByRole('button', { name: /send test message/i }).click();
      await expect(page.getByText(`Test message sent to ${testEmail}.`)).toBeVisible({ timeout: timeouts.tenSeconds });

      const notice = await awaitNotice(emailClient, subject);
      // what the preview promised is what actually lands in the inbox
      expect(notice.body).toContain(noticeHeadline);
      expect(notice.body).toContain('migrate to the gadget API instead');
      expect(notice.body).toContain(defaultTitle);
      expect(notice.body).toContain(defaultGreeting);
    });

    test('broadcasts to the tenant admins across tenant boundaries', async ({ adminBaseUrl, environment, request, username }) => {
      const emailClient = requireEmailClient(username, environment);
      const subject = `e2e admin panel broadcast ${Date.now()}`;

      // driven through the API on purpose: the UI path ends in a `window.confirm` that playwright
      // dismisses by default, and the confirmation guard in front of it is covered above
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
