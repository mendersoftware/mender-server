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
import type { APIRequestContext } from '@playwright/test';

import test, { expect } from '../../fixtures/fixtures';
import { adminPanelApiUrl, findSecondaryTenant, getTenantDetail } from '../../utils/adminPanel';
import { login } from '../../utils/commands';
import { timeouts } from '../../utils/constants';

// the user the runner creates together with the plain `secondary` tenant
const secondaryTenantUser = 'demo-secondary@example.com';

const raisedDeviceLimit = 42;

/** logging in is the cheapest way to find out whether a tenant still grants its users access */
const canLogIn = async (username: string, password: string, baseUrl: string, request: APIRequestContext): Promise<boolean> => {
  try {
    const { token } = await login(username, password, baseUrl, request);
    return !!token;
  } catch {
    // a suspended tenant makes useradm reject the credentials outright
    return false;
  }
};

// These drive operator actions that reach past the panel and change what a tenant is allowed to do,
// so they stay on the plain `secondary` tenant and put back whatever they changed.
test.describe('Admin panel operator actions', () => {
  test.beforeEach(async ({ browserName, environment }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
    test.skip(browserName !== 'chromium', 'tenant wide mutations only need to be exercised once');
  });

  test('propagates a raised device limit through to the tenant', async ({ adminBaseUrl, baseUrl, password, page, request }) => {
    const tenant = await findSecondaryTenant(request, adminBaseUrl);
    const previousLimit = (await getTenantDetail(request, adminBaseUrl, tenant.id)).device_limits?.max_devices?.value ?? -1;

    await page.goto(`${adminBaseUrl}tenants/${tenant.id}`);
    await expect(page.getByRole('heading', { name: `Tenant: ${tenant.name}` })).toBeVisible();
    const standardLimit = page.getByRole('row', { name: /^standard/ }).getByRole('textbox');
    await expect(standardLimit).toHaveValue(String(previousLimit));

    try {
      await standardLimit.fill(String(raisedDeviceLimit));
      await standardLimit.blur();
      await page.getByRole('button', { name: 'Save' }).click();
      await expect(page.getByText('Tenant updated successfully')).toBeVisible({ timeout: timeouts.tenSeconds });

      // the tenant's own session has to be told about the new ceiling, not just the operator
      const { token } = await login(secondaryTenantUser, password, baseUrl, request);
      const limitResponse = await request.get(`${baseUrl}api/management/v2/devauth/limits/max_devices`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      expect(limitResponse.ok()).toBeTruthy();
      const { limit } = await limitResponse.json();
      expect(limit).toEqual(raisedDeviceLimit);
    } finally {
      // leave the tenant on the limit it came with, even if the assertions above failed
      const restore = await request.put(adminPanelApiUrl(adminBaseUrl, `tenants/${tenant.id}/device-limits`), {
        data: [{ name: 'max_devices', value: previousLimit }]
      });
      expect(restore.ok()).toBeTruthy();
    }
  });

  test('locks a suspended tenant out and lets it back in again', async ({ adminBaseUrl, baseUrl, password, page, request }) => {
    const tenant = await findSecondaryTenant(request, adminBaseUrl);
    expect(await canLogIn(secondaryTenantUser, password, baseUrl, request)).toBeTruthy();

    await page.goto(`${adminBaseUrl}tenants/${tenant.id}`);
    const status = page.getByRole('combobox', { name: 'Select status' });
    await expect(status).toHaveText('active');

    try {
      await status.click();
      await page.getByRole('option', { name: 'suspended' }).click();
      await page.getByRole('button', { name: 'Save' }).click();
      await expect(page.getByText('Tenant updated successfully')).toBeVisible({ timeout: timeouts.tenSeconds });

      const suspended = await getTenantDetail(request, adminBaseUrl, tenant.id);
      expect(suspended.status).toEqual('suspended');
      // suspending is the operator's kill switch, it has to actually cut the tenant's users off
      expect(await canLogIn(secondaryTenantUser, password, baseUrl, request)).toBeFalsy();
    } finally {
      // reactivate through the API - this has to happen even if the form never got that far
      const reactivation = await request.put(adminPanelApiUrl(adminBaseUrl, `tenants/${tenant.id}/status`), { data: { status: 'active' } });
      expect(reactivation.ok()).toBeTruthy();
    }

    const reactivated = await getTenantDetail(request, adminBaseUrl, tenant.id);
    expect(reactivated.status).toEqual('active');
    expect(await canLogIn(secondaryTenantUser, password, baseUrl, request)).toBeTruthy();
  });
});
