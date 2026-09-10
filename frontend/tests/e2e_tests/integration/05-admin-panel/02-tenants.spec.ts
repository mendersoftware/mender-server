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
import { adminPanelApiUrl, findSecondaryTenant, getTenantDetail, tenantNames } from '../../utils/adminPanel';
import { timeouts } from '../../utils/constants';

const gridRow = '.MuiDataGrid-row';

test.describe('Admin panel tenants', () => {
  test.beforeEach(async ({ environment }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
  });

  test('searches the tenant list and opens what it finds', async ({ adminBaseUrl, page }) => {
    await page.goto(`${adminBaseUrl}tenants`);
    const mainRow = page.getByRole('gridcell', { name: tenantNames.main, exact: true });
    const secondaryRow = page.getByRole('gridcell', { name: tenantNames.secondary, exact: true }).first();
    await expect(mainRow).toBeVisible();
    await expect(secondaryRow).toBeVisible();

    const search = page.getByPlaceholder('Search tenants...');
    await search.fill(tenantNames.main);
    await expect(mainRow).toBeVisible();
    await expect(secondaryRow).not.toBeVisible();

    await search.fill('there-is-no-such-tenant');
    await expect(page.getByText('No tenants found')).toBeVisible();
    await expect(page.locator(gridRow)).toHaveCount(0);

    await page.getByRole('button', { name: 'Clear search' }).click();
    await expect(secondaryRow).toBeVisible();
    // the whole row is clickable, not just the id link
    await secondaryRow.click();
    await expect(page.getByRole('heading', { name: `Tenant: ${tenantNames.secondary}` })).toBeVisible();
  });

  test('guards the fields that must not be edited by accident', async ({ adminBaseUrl, page, request }) => {
    const tenant = await findSecondaryTenant(request, adminBaseUrl);
    await page.goto(`${adminBaseUrl}tenants/${tenant.id}`);
    await expect(page.getByRole('heading', { name: `Tenant: ${tenant.name}` })).toBeVisible();

    const trialCheckbox = page.getByRole('checkbox', { name: 'Trial' });
    const expiration = page.getByLabel('Trial Expiration');
    // an expiration only means something on a trial, so the field follows the checkbox
    await expect(trialCheckbox).not.toBeChecked();
    await expect(expiration).toHaveCount(0);
    await trialCheckbox.check();
    await expect(expiration).toBeVisible();
    await trialCheckbox.uncheck();
    await expect(expiration).toHaveCount(0);

    const customerId = page.getByLabel('Customer ID');
    // billing automation hangs off these, so they are not editable by accident
    await expect(customerId).toBeDisabled();
    await page.getByRole('button', { name: 'Edit Subscription' }).click();
    await expect(customerId).toBeEnabled();
    await page.getByRole('button', { name: 'Lock' }).click();
    await expect(customerId).toBeDisabled();
  });

  test('saves an edited tenant and reports back', async ({ adminBaseUrl, page, request }) => {
    const tenant = await findSecondaryTenant(request, adminBaseUrl);
    const editedName = `${tenant.name}-edited`;
    await page.goto(`${adminBaseUrl}tenants/${tenant.id}`);
    const name = page.getByLabel('Name');
    await expect(name).toHaveValue(tenant.name);

    try {
      await name.fill(editedName);
      await page.getByRole('button', { name: 'Save' }).click();
      // saving fans out over the profile, limit and delta endpoints before it reports success
      await expect(page.getByText('Tenant updated successfully')).toBeVisible({ timeout: timeouts.tenSeconds });

      const updated = await getTenantDetail(request, adminBaseUrl, tenant.id);
      expect(updated.name).toEqual(editedName);
    } finally {
      // put the name back so the lookup other specs rely on keeps working
      const restore = await request.put(adminPanelApiUrl(adminBaseUrl, `tenants/${tenant.id}`), { data: { name: tenant.name } });
      expect(restore.ok()).toBeTruthy();
    }

    const restored = await getTenantDetail(request, adminBaseUrl, tenant.id);
    expect(restored.name).toEqual(tenant.name);
  });
});
