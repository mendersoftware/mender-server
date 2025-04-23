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
import test, { expect } from '../fixtures/fixtures.ts';
import { baseUrlToDomain, getTokenFromStorage, prepareCookies, prepareNewPage, processLoginForm } from '../utils/commands.ts';
import { spStoragePath, switchTenantStoragePath, timeouts } from '../utils/constants.ts';

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
  test.use({ storageState: spStoragePath });
  test.beforeEach(async ({ page, environment }) => {
    test.skip(environment !== 'enterprise', 'not available in OS');
    await page
      .locator('.leftFixed.leftNav')
      .getByRole('link', { name: /Tenants/i })
      .click();
  });
  test('tenant creation', async ({ page }) => {
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
  test('tenant edit', async ({ page }) => {
    await page.getByText('View details').click();
    await page.getByRole('button', { name: /edit device limit/i }).click();
    await page.getByLabel(/Set device limit/i).fill('12');
    await page.getByRole('button', { name: /save/i }).click();
    await page.getByLabel('close').click();
    await expect(page.getByText('0/12')).toBeVisible();
  });
  test('tenant removal', async ({ page }) => {
    await page.getByText('View details').click();
    await page.getByRole('button', { name: /delete tenant/i }).click();
    const confirmInput = page.getByRole('textbox', { name: /delete/i });
    await confirmInput.fill('delete');
    await page.getByRole('button', { name: /Confirm/i }).click();
  });
  test('create a tenant Role', async ({ baseUrl, page }) => {
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
  test('remove a tenant Role', async ({ baseUrl, page }) => {
    await page.goto(`${baseUrl}ui/settings/role-management`);
    await page.getByRole('cell', { name: tenantRole.name }).click();
    await page.getByRole('button', { name: /delete role/i }).click();
    await page.getByLabel(`${tenantRole.name} *`).fill(tenantRole.name);
    await page.getByRole('button', { name: /Confirm/i }).click();
    await page.waitForTimeout(timeouts.default);
    await expect(page.getByText('SP_tenant_role')).not.toBeVisible();
  });
});

test.describe('Multi tenant access', () => {
  const secondaryUser = 'demo-secondary@example.com';
  const tenantIdDescriptor = 'Tenant ID:';
  let userId = '';
  test('allows adding users to tenants', async ({ baseUrl, browser, browserName, environment, page: loggedInPage, request, password }) => {
    test.skip('enterprise' !== environment || browserName !== 'chromium');
    await loggedInPage.goto(`${baseUrl}ui/settings`);
    await loggedInPage.getByRole('link', { name: /user management/i }).click();
    const hasUserAlready = await loggedInPage.getByText(secondaryUser).isVisible();
    test.skip(hasUserAlready, `${secondaryUser} was added in a previous run, but success notification wasn't caught`);
    const page = await prepareNewPage({ baseUrl, browser, password, request, username: secondaryUser });
    await page.goto(`${baseUrl}ui/settings/my-account`);
    await page
      .getByText(/User ID/i)
      .locator('..')
      .locator('..')
      .getByRole('button', { name: /copy to clipboard/i })
      .click({ force: true });
    const content = await page.evaluateHandle(() => navigator.clipboard.readText());
    userId = await content.jsonValue();
    expect(userId).toBeTruthy();
    await page.getByText(/help/i).click();
    await page.getByRole('button', { name: secondaryUser }).click();
    await expect(page.getByText(/switch organization/i)).not.toBeVisible();

    await loggedInPage.getByRole('button', { name: /new user/i }).click();
    const passwordInput = await loggedInPage.getByPlaceholder(/password/i);
    const emailUuidInput = await loggedInPage.getByPlaceholder(/email/i);
    await emailUuidInput.click();
    await emailUuidInput.fill(userId);
    await expect(passwordInput).not.toBeVisible();
    await loggedInPage.getByRole('button', { name: /add user/i }).click();
    await loggedInPage.getByText('The user was added successfully.').waitFor();
    await expect(loggedInPage.getByText('The user was added successfully.')).toBeVisible();
    await loggedInPage.screenshot({ path: './test-results/switch-user-added.png' });

    await page.reload();
    await page.getByRole('button', { name: secondaryUser }).click();
    await expect(page.getByRole('button', { name: /switch organization/i })).toBeVisible();
    await page.context().close();
  });
  test('allows switching tenants', async ({ baseUrl, browser, browserName, environment, page: loggedInPage, password, request }) => {
    test.skip('enterprise' !== environment || browserName !== 'chromium');
    await loggedInPage.goto(`${baseUrl}ui/settings`);
    await loggedInPage.getByRole('link', { name: /user management/i }).click();
    const hasUserAlready = await loggedInPage.getByText(secondaryUser).isVisible();
    if (!hasUserAlready) {
      await loggedInPage.getByRole('link', { name: /organization/i }).click();
      await loggedInPage.getByText('test').click();
      const content = await loggedInPage.evaluateHandle(() => navigator.clipboard.readText());
      const tenantInfo = await content.jsonValue();
      const tenantId = tenantInfo.substring(tenantInfo.indexOf(tenantIdDescriptor) + tenantIdDescriptor.length).trim();
      const token = await getTokenFromStorage(baseUrl);
      const options = { headers: { Authorization: `Bearer ${token}` }, data: { tenant_ids: [tenantId] } };
      const response = await request.post(`${baseUrl}api/management/v1/useradm/users/${userId}/assign`, options);
      expect(response.ok()).toBeTruthy();
      await loggedInPage.getByRole('link', { name: /user management/i }).click();
      await loggedInPage.screenshot({ path: './test-results/switch-user-list.png' });
    }
    // here we can't use prepareNewPage as it sets the initial JWT to be used on every page init
    const domain = baseUrlToDomain(baseUrl);
    let newContext = await browser.newContext({ storageState: switchTenantStoragePath });
    newContext = await prepareCookies(newContext, domain, '');
    const page = await newContext.newPage();
    await page.goto(`${baseUrl}ui/`);
    await processLoginForm({ username: secondaryUser, password, page, environment });
    await page.getByRole('button', { name: secondaryUser }).click();
    await expect(page.getByRole('menuitem', { name: /secondary/i })).toBeVisible();
    await page.screenshot({ path: './test-results/switch-try-switch.png' });
    await page.getByRole('button', { name: /switch organization/i }).click({ force: true });
    const tenantSwitch = await page.getByRole('menuitem', { name: /test/i });
    await tenantSwitch.waitFor({ timeout: timeouts.default });
    await tenantSwitch.click();
    await page.waitForTimeout(timeouts.default);
    await page.getByRole('button', { name: secondaryUser }).click();
    await expect(page.getByRole('menuitem', { name: /secondary/i })).not.toBeVisible();
    await expect(page.getByRole('menuitem', { name: /test/i })).toBeVisible();
    await page.screenshot({ path: './test-results/switch-post-switch.png' });

    await loggedInPage.goto(`${baseUrl}ui/settings`);
    await loggedInPage.getByRole('link', { name: /user management/i }).click();
    await loggedInPage.getByText(secondaryUser).click();
    await loggedInPage.getByRole('button', { name: /delete user/i }).click();
    await expect(loggedInPage.getByText(/delete user\?/i)).toBeVisible();
    await loggedInPage
      .getByRole('button', { name: /delete user/i })
      .last()
      .click();

    await page.getByRole('menuitem', { name: /log out/i }).click();
    await processLoginForm({ username: secondaryUser, password, page, environment });
    await page.getByRole('button', { name: secondaryUser }).click();
    await expect(page.getByText(/switch organization/i)).not.toBeVisible();
    await newContext.close();
  });
});
