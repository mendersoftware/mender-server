// Copyright 2024 Northern.tech AS
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
import { Server } from 'net';

import test, { expect } from '../fixtures/fixtures.ts';
import { startWebhookServer } from '../utils/commands.ts';
import { selectors, timeouts } from '../utils/constants.ts';

const baseWebhookLocation = 'http://docker.mender.io:9000/webhooks';

test.describe('Webhooks Functionality', () => {
  let server: Server;
  test.beforeAll(({ environment }) => {
    test.skip(environment === 'staging');
    server = startWebhookServer();
  });
  test.afterAll(() => server.close());
  test('allows configuring basic webhooks', async ({ baseUrl, environment, page }) => {
    test.skip(environment === 'staging');
    await page.goto(`${baseUrl}ui/settings`);
    await page.getByText(/integrations/i).click();
    await page.getByLabel(/add an integration/i).click();
    await page.getByRole('option', { name: /Webhooks/i }).click();
    await page.getByTitle(/webhook details/i).isVisible({ timeout: timeouts.default });
    await expect(page.getByRole('button', { name: /save/i })).toBeDisabled();
    await page.getByLabel(/url/i).fill(`${baseWebhookLocation}/all`);
    await page.getByLabel(/description/i).fill('some description');
    await expect(page.getByRole('button', { name: /save/i })).not.toBeDisabled();
    await page.getByRole('button', { name: /save/i }).click();
    await expect(page.getByText(/view details/i)).toBeVisible();
    await expect(page.getByText(/one active integration at a time/i)).toBeVisible();
  });
  test('shows webhook details', async ({ baseUrl, environment, page }) => {
    test.skip(environment === 'staging');
    await page.goto(`${baseUrl}ui/devices`);
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await expect(page.getByText(/Device information for/i)).toBeVisible();
    await page.getByText('Dismiss', { exact: true }).click();
    await page.getByRole('button', { name: 'Confirm' }).click();
    await page.reload();
    await page.getByText('Accept', { exact: true }).click();
    await page.goto(`${baseUrl}ui/settings/integrations`);
    await page.getByText(/view details/i).click();
    await page
      .getByText(/device provisioned/i)
      .first()
      .click();
    await expect(page.getByText('pubkey')).toBeVisible();
    await page.getByText(/back to webhook/i).click();
  });
  test('allows deleting a webhook', async ({ baseUrl, environment, page }) => {
    test.skip(environment === 'staging');
    await page.goto(`${baseUrl}ui/settings/integrations`);
    await page.getByText(/view details/i).click();
    await page.getByText(/delete webhook/i).click();
    await expect(page.getByLabel(/add an integration/i)).toBeVisible();
  });
  test('allows configuring inventory webhooks', async ({ baseUrl, environment, page }) => {
    test.skip(environment !== 'enterprise');
    await page.goto(`${baseUrl}ui/settings/integrations`);
    await page.getByLabel(/add an integration/i).click();
    await page.getByRole('option', { name: /Webhooks/i }).click();
    await page.getByTitle(/webhook details/i).isVisible({ timeout: timeouts.default });
    await page.getByLabel(/url/i).fill(`${baseWebhookLocation}/inventory`);
    await page.getByLabel(/device authentication/i).click();
    await page.getByLabel(/device inventory/i).click();
    await page.screenshot({ path: './test-results/save-webhook.png' });
    await page.getByRole('button', { name: /save/i }).click();
    await expect(page.getByText(/view details/i)).toBeVisible();
    await expect(page.getByText(/one active integration at a time/i)).toBeVisible();
  });
  test('shows webhook details for inventory events', async ({ baseUrl, environment, page }, { retry }) => {
    test.skip(environment !== 'enterprise');
    await page.goto(`${baseUrl}ui/settings/integrations`);
    await page.getByText(/view details/i).click();
    const inventoryChangeCount = (await page.getByText(/inventory changed/).all()).length;
    await page.getByLabel(/close/i).click();

    await page.getByRole('link', { name: /Devices/i }).click();
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await expect(page.getByText('Device information for')).toBeVisible();
    await page.locator('button:near(:text("Tags"))').first().click();
    await page.getByPlaceholder(/key/i).fill('foo');
    await page.getByPlaceholder(/value/i).fill(`bar ${retry}`);
    await page.getByRole('button', { name: /save/i }).click();
    await expect(page.getByText(/device tags changed/i)).toBeVisible();
    await expect(page.getByPlaceholder(/key/i)).not.toBeVisible();

    await page.goto(`${baseUrl}ui/settings/integrations`);
    await page.getByText(/view details/i).click();
    const newInventoryChangeCount = (await page.getByText(/inventory changed/).all()).length;
    expect(newInventoryChangeCount).toBeGreaterThan(inventoryChangeCount);
    await page
      .getByText(/inventory changed/)
      .first()
      .click();
    await expect(page.getByText('tags-foo')).toBeVisible();
  });
});
