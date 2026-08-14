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
import { findMainTenant, getTenantDetail, listTenants, listUsers } from '../../utils/adminPanel';
import { getTokenFromStorage } from '../../utils/commands';

// The panel's own interactions are covered by the isolated `admin-panel` project. What is left here
// needs the state the preceding suites build up - an authorized device and a tenant session to
// compare the operator view against.
test.describe('Admin panel cross checks', () => {
  test.beforeEach(async ({ environment }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
  });

  test('counts the tenants of this installation on the dashboard', async ({ adminBaseUrl, page, request }) => {
    const { total } = await listTenants(request, adminBaseUrl, { per_page: 1 });
    // the runner creates the main tenant plus a service provider and a plain secondary one
    expect(total).toBeGreaterThanOrEqual(3);
    const { total: enterpriseTotal } = await listTenants(request, adminBaseUrl, { per_page: 1, plan: 'enterprise' });

    await page.goto(adminBaseUrl);
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();

    // the tiles are driven off `X-Total-Count`, so a dropped header shows up as a zeroed dashboard
    await expect(page.getByText('Total Tenants').locator('..').locator('h3')).toHaveText(String(total));

    // `tenantadm create-org` defaults to the enterprise plan, so the breakdown has to account for them
    await expect(page.getByText('Tenants by Plan')).toBeVisible();
    const enterpriseRow = page.getByText('Enterprise', { exact: true }).locator('..');
    await expect(enterpriseRow.getByText(String(enterpriseTotal), { exact: true })).toBeVisible();
    // NB the user tile is left out on purpose - the user store reports no total to count with
  });

  test('aggregates the device counts deviceauth reports for a tenant', async ({ adminBaseUrl, baseUrl, page, request }) => {
    const tenant = await findMainTenant(request, adminBaseUrl);
    const detail = await getTenantDetail(request, adminBaseUrl, tenant.id);
    // a partial aggregate is reported rather than failing outright, so surface it here
    expect(detail.warnings ?? []).toEqual([]);
    const { device_counts: counts } = detail;
    if (!counts) {
      throw new Error('the aggregated tenant detail came back without any device counts');
    }

    // the panel has no total of its own, it sums the individual auth set states
    expect(counts.total).toEqual(counts.accepted + counts.pending + counts.preauthorized + counts.rejected + counts.noauth);
    // the device the device tests authorized has to be part of the aggregate
    expect(counts.accepted).toBeGreaterThan(0);

    // cross check the operator view against what the tenant's own session is told
    const token = await getTokenFromStorage(baseUrl);
    const response = await request.get(`${baseUrl}api/management/v2/devauth/devices/count`, {
      headers: { Authorization: `Bearer ${token}` },
      params: { status: 'accepted' }
    });
    expect(response.ok()).toBeTruthy();
    const { count } = await response.json();
    expect(counts.accepted).toEqual(count);

    await page.goto(`${adminBaseUrl}tenants/${tenant.id}`);
    await expect(page.getByRole('heading', { name: `Tenant: ${tenant.name}` })).toBeVisible();
    const deviceTile = (state: string) => page.getByText(state, { exact: true }).locator('..');
    await expect(deviceTile('accepted').locator('h6')).toHaveText(String(counts.accepted));
    // every state the total is summed from has to be accounted for in the breakdown
    await expect(deviceTile('noauth').locator('h6')).toHaveText(String(counts.noauth));
  });

  test('sees the users of every tenant at once', async ({ adminBaseUrl, request, spTenantUsername, username }) => {
    const users = await listUsers(request, adminBaseUrl, { per_page: 100 });
    const emails = users.map(({ email }) => email);
    // the tenant facing UI is scoped to one tenant, the panel is what spans them
    expect(emails).toContain(username);
    expect(emails).toContain(spTenantUsername);

    const tenant = await findMainTenant(request, adminBaseUrl);
    const scoped = await listUsers(request, adminBaseUrl, { per_page: 100, tenant_id: tenant.id });
    const scopedEmails = scoped.map(({ email }) => email);
    expect(scopedEmails).toContain(username);
    expect(scopedEmails).not.toContain(spTenantUsername);
  });
});
