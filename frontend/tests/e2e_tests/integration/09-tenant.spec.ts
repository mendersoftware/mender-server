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
import * as fs from 'fs';

import test, { expect } from '../fixtures/fixtures.ts';
import { prepareNewPage } from '../utils/commands.ts';
import { storagePath, timeouts } from '../utils/constants.ts';

const tenant = {
  name: 'Child Tenant',
  adminUser: 'child@example.com',
  limit: '10'
};

const tenantRole = {
  name: 'SP_tenant_role',
  description: 'Test role for SP tenant'
};
test.describe('Tenant Functionality', () => {
  test.beforeAll(async ({ baseUrl, browser, password, spTenantUsername }) => {
    const storageLocation = `tenant-${storagePath}`;
    fs.writeFileSync(storageLocation, JSON.stringify({ cookies: [], origins: [] }));
    const context = await browser.newContext();
    await prepareNewPage({ baseUrl, context, password, storageLocation, username: spTenantUsername });
    await context.storageState({ path: storageLocation });
  });
  test.beforeEach(async ({ loggedInTenantPage: page, environment }) => {
    test.skip(environment !== 'enterprise', 'not available in OS');
    await page
      .locator('.leftFixed.leftNav')
      .getByRole('link', { name: /Tenants/i })
      .click();
  });
  test('tenant creation', async ({ loggedInTenantPage: page }) => {
    await expect(page.locator('h2:has-text("Tenants")')).toBeVisible();
    await page.getByRole('button', { name: /add tenant/i }).click();
    const nameInput = page.getByPlaceholder(/Name/i);
    await nameInput.fill(tenant.name);
    const adminInput = page.getByRole('textbox', { name: /admin user/i });
    await adminInput.fill(tenant.adminUser);
    await page.getByRole('button', { name: /generate/i }).click();
    const deviceLimitInput = page.getByLabel(/Set device limit/i);
    await deviceLimitInput.fill(tenant.limit);
    await page.getByText(/enable delta artifact generation/i).click();
    const submitButton = page.getByRole('button', { name: /Create Tenant/i });
    await submitButton.scrollIntoViewIfNeeded();
    await submitButton.click();
    await expect(page.getByText('Child Tenant')).toBeVisible();
  });
  test('tenant edit', async ({ loggedInTenantPage: page }) => {
    await page.getByText('View details').click();
    await page.getByRole('button', { name: /edit device limit/i }).click();
    await page.getByTestId('dev-limit-input').fill('12');
    await page.getByRole('button', { name: /save/i }).click();
    await page.getByTestId('CloseIcon').click();
    await expect(page.getByText('0/12')).toBeVisible();
  });
  test('tenant removal', async ({ loggedInTenantPage: page }) => {
    await page.getByText('View details').click();
    await page.getByRole('button', { name: /delete tenant/i }).click();
    const confirmInput = page.getByRole('textbox', { name: /delete/i });
    await confirmInput.fill('delete');
    await page.getByRole('button', { name: /Confirm/i }).click();
  });
  test('create a tenant Role', async ({ baseUrl, loggedInTenantPage: page }) => {
    await page.goto(`${baseUrl}ui/settings/role-management`);
    await page.getByRole('button', { name: /add a role/i }).click();
    await page.getByLabel('Name').fill(tenantRole.name);
    await page.getByLabel('Description').fill(tenantRole.description);
    await page.locator('#mui-component-select-tenantManagement').click();
    await page.getByRole('option', { name: 'Manage' }).click();
    await page.locator('#menu-tenantManagement > .MuiBackdrop-root').click();
    await page.getByRole('button', { name: /Submit/i }).click();
    await expect(page.getByText('SP_tenant_role')).toBeVisible();
  });
  test('remove a tenant Role', async ({ baseUrl, loggedInTenantPage: page }) => {
    await page.goto(`${baseUrl}ui/settings/role-management`);
    await page.getByRole('cell', { name: tenantRole.name }).click();
    await page.getByRole('button', { name: /delete role/i }).click();
    await page.getByLabel(`${tenantRole.name} *`).fill(tenantRole.name);
    await page.getByRole('button', { name: /Confirm/i }).click();
    await page.waitForTimeout(timeouts.default);
    await expect(page.getByText('SP_tenant_role')).not.toBeVisible();
  });
});
