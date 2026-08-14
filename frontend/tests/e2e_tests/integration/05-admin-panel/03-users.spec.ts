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
import { adminPanelUrl, findMainTenant, isAdminPanelAvailable } from '../../utils/adminPanel';

test.describe('Admin panel users', () => {
  test.beforeEach(async ({ baseUrl, environment, request }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
    test.skip(!(await isAdminPanelAvailable(request, baseUrl)), 'the admin panel is not routed in this environment');
  });

  test('lists the users of every tenant by default', async ({ baseUrl, page, spTenantUsername, username }) => {
    await page.goto(adminPanelUrl(baseUrl, 'users'));
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();
    // users of separate tenants side by side is what the tenant facing UI can never show
    await expect(page.getByRole('gridcell', { name: username })).toBeVisible();
    await expect(page.getByRole('gridcell', { name: spTenantUsername })).toBeVisible();
  });

  test('filters by email and clears the filter again', async ({ baseUrl, page, spTenantUsername, username }) => {
    await page.goto(adminPanelUrl(baseUrl, 'users'));
    await page.getByLabel('Email').fill(username);
    await page.getByRole('button', { name: 'Search' }).click();
    await expect(page.getByRole('gridcell', { name: username })).toBeVisible();
    await expect(page.getByRole('gridcell', { name: spTenantUsername })).not.toBeVisible();

    await page.getByRole('button', { name: 'Clear' }).click();
    await expect(page.getByLabel('Email')).toHaveValue('');
    await expect(page.getByRole('gridcell', { name: spTenantUsername })).toBeVisible();
  });

  test('filters down to the users of a single tenant', async ({ baseUrl, page, request, spTenantUsername, username }) => {
    const tenant = await findMainTenant(request, baseUrl);
    await page.goto(adminPanelUrl(baseUrl, 'users'));
    await page.getByLabel('Tenant ID').fill(tenant.id);
    await page.getByRole('button', { name: 'Search' }).click();
    await expect(page.getByRole('gridcell', { name: username })).toBeVisible();
    await expect(page.getByRole('gridcell', { name: spTenantUsername })).not.toBeVisible();
  });

  test('only accepts an sso subject together with its provider', async ({ baseUrl, page }) => {
    await page.goto(adminPanelUrl(baseUrl, 'users'));
    const ssoSubject = page.getByLabel('SSO Subject');
    // a subject means nothing without knowing which provider issued it
    await expect(ssoSubject).toBeDisabled();

    await page.getByLabel('SSO Provider').click();
    await page.getByRole('option', { name: 'google' }).click();
    await expect(ssoSubject).toBeEnabled();

    await ssoSubject.fill('someone@example.com');
    await page.getByLabel('SSO Provider').click();
    await page.getByRole('option', { name: 'Any' }).click();
    // dropping the provider has to drop the subject with it, rather than leave a dangling filter
    await expect(ssoSubject).toBeDisabled();
    await expect(ssoSubject).toHaveValue('');
  });

  test('carries the tenant over when arriving from a tenant', async ({ baseUrl, page, request }) => {
    const tenant = await findMainTenant(request, baseUrl);
    await page.goto(adminPanelUrl(baseUrl, `tenants/${tenant.id}`));
    await page.getByRole('link', { name: 'View Users' }).click();
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();
    // the operator lands on a list that is already narrowed to the tenant they came from
    await expect(page.getByLabel('Tenant ID')).toHaveValue(tenant.id);
  });
});
