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
import { isEnterpriseOrStaging, startDockerClient, startWebhookServer, stopDockerClient, tenantTokenRetrieval } from '../utils/commands.ts';
import { storagePath, timeouts } from '../utils/constants.ts';

const baseWebhookLocation = 'http://docker.mender.io:9000/webhooks';

test.describe('Webhooks Functionality', () => {
  let server: Server;
  test.use({ storageState: storagePath });
  test.beforeAll(() => {
    server = startWebhookServer();
  });
  test.afterAll(() => server.close());
  test('allows configuring basic webhooks', async ({ baseUrl, loggedInPage: page }) => {
    await page.goto(`${baseUrl}ui/settings`);
    await page.getByText(/integrations/i).click();
    await page.getByLabel(/add an integration/i).click();
    await page.getByRole('option', { name: /Webhooks/i }).click();
    await page.getByTitle(/webhook details/i).isVisible({ timeout: timeouts.default });
    await expect(page.getByRole('button', { name: /save/i })).toBeDisabled();
    await page.getByLabel(/url/i).fill(`${baseWebhookLocation}/all`);
    await page.getByLabel(/description/i).fill('some description');
    await expect(page.getByRole('button', { name: /save/i })).not.toBeDisabled();
    await page.getByLabel(/device authentication/i).click();
    await page.getByRole('button', { name: /save/i }).click();
    await expect(page.getByText(/view details/i)).toBeVisible();
    await expect(page.getByText(/one active integration at a time/i)).toBeVisible();
  });
  test('shows webhook details', async ({ baseUrl, loggedInPage: page }) => {
    await page.goto(`${baseUrl}ui/settings/integrations`);
    await page.getByText(/view details/i).click();
    await page
      .getByText(/device provisioned/i)
      .first()
      .click();
    await expect(page.getByText('pubkey')).toBeVisible();
    await page.getByText(/back to webhook/i).click();
  });
  test('allows deleting a webhook', async ({ baseUrl, loggedInPage: page }) => {
    await page.goto(`${baseUrl}ui/settings/integrations`);
    await page.getByText(/view details/i).click();
    await page.getByText(/delete webhook/i).click();
    await expect(page.getByLabel(/add an integration/i)).toBeVisible();
  });
  test('allows configuring inventory webhooks', async ({ baseUrl, environment, loggedInPage: page }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    await page.goto(`${baseUrl}ui/settings/integrations`);
    await page.getByLabel(/add an integration/i).click();
    await page.getByRole('option', { name: /Webhooks/i }).click();
    await page.getByTitle(/webhook details/i).isVisible({ timeout: timeouts.default });
    await page.getByLabel(/url/i).fill(`${baseWebhookLocation}/inventory`);
    await page.getByLabel(/device authentication/i).click();
    await page.getByLabel(/device inventory/i).click();
    await page.getByRole('button', { name: /save/i }).click();
    await page.screenshot({ path: './test-results/save-webhook.png' });
    await expect(page.getByText(/view details/i)).toBeVisible();
    await expect(page.getByText(/one active integration at a time/i)).toBeVisible();
    await page.screenshot({ path: './test-results/view-webhook.png' });
  });
  test('shows webhook details for inventory events', async ({ baseUrl, loggedInPage: page }) => {
    await stopDockerClient();
    const token = await tenantTokenRetrieval(baseUrl, page);
    await startDockerClient(baseUrl, token);
    await page.goto(`${baseUrl}ui/settings/integrations`);
    await page.getByText(/view details/i).click();
    await page
      .getByText(/inventory changed/i)
      .first()
      .click();
    await expect(page.getByText(/device_type/)).toBeVisible();
    await page.getByText(/back to webhook/i).click();
  });
});
