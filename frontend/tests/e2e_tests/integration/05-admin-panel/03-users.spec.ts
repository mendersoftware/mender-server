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
import { findMainTenant } from '../../utils/adminPanel';

test.describe('Admin panel users', () => {
  test.beforeEach(async ({ environment }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
  });

  test('filters the cross tenant user list by email and by tenant', async ({ adminBaseUrl, page, request, spTenantUsername, username }) => {
    const tenant = await findMainTenant(request, adminBaseUrl);
    const mainUser = page.getByRole('gridcell', { name: username, exact: true });
    const spUser = page.getByRole('gridcell', { name: spTenantUsername, exact: true });

    await page.goto(`${adminBaseUrl}users`);
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();
    await expect(mainUser).toBeVisible();

    // a user of an entirely different tenant is one filter away
    await page.getByLabel('Email').fill(spTenantUsername);
    await page.getByRole('button', { name: 'Search' }).click();
    await expect(spUser).toBeVisible();
    await expect(mainUser).not.toBeVisible();

    await page.getByRole('button', { name: 'Clear' }).click();
    await expect(page.getByLabel('Email')).toHaveValue('');
    await expect(mainUser).toBeVisible();

    await page.getByLabel('Tenant ID').fill(tenant.id);
    await page.getByRole('button', { name: 'Search' }).click();
    await expect(mainUser).toBeVisible();
    await expect(spUser).not.toBeVisible();
  });

  test('only accepts an sso subject together with its provider', async ({ adminBaseUrl, page }) => {
    await page.goto(`${adminBaseUrl}users`);
    const ssoSubject = page.getByLabel('SSO Subject');
    // a subject means nothing without knowing which provider issued it
    await expect(ssoSubject).toBeDisabled();

    // the provider select has no accessible name of its own yet, so it is addressed by position
    const ssoProviderSelect = page.getByRole('combobox').first();
    await ssoProviderSelect.click();
    await page.getByRole('option', { name: 'google' }).click();
    await expect(ssoSubject).toBeEnabled();

    await ssoSubject.fill('someone@example.com');
    await ssoProviderSelect.click();
    await page.getByRole('option', { name: 'Any' }).click();
    // dropping the provider has to drop the subject with it, rather than leave a dangling filter
    await expect(ssoSubject).toBeDisabled();
    await expect(ssoSubject).toHaveValue('');
  });

  test('carries the tenant over when arriving from a tenant', async ({ adminBaseUrl, page, request }) => {
    const tenant = await findMainTenant(request, adminBaseUrl);
    await page.goto(`${adminBaseUrl}tenants/${tenant.id}`);
    await page.getByRole('link', { name: 'View Users' }).click();
    await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();
    // the operator lands on a list that is already narrowed to the tenant they came from
    await expect(page.getByLabel('Tenant ID')).toHaveValue(tenant.id);
  });
});
