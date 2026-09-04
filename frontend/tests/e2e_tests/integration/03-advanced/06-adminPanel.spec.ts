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
import { deviceStateCount, findMainTenant, getTenantDetail } from '../../utils/adminPanel';
import { getTokenFromStorage } from '../../utils/commands';

// The panel's own interactions are covered by the isolated `admin-panel` project. What is left here
// needs the state the preceding suites build up - an authorized device and a tenant session to
// compare the operator view against.
test.describe('Admin panel cross checks', () => {
  test.beforeEach(async ({ environment }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
  });

  test('reports the same device counts to the operator as to the tenant', async ({ adminBaseUrl, baseUrl, page, request }) => {
    const tenant = await findMainTenant(request, adminBaseUrl);
    const detail = await getTenantDetail(request, adminBaseUrl, tenant.id);
    // a partial aggregate is reported rather than failing outright, so surface it here
    expect(detail.warnings ?? []).toEqual([]);
    const { device_counts: counts } = detail;
    if (!counts) {
      throw new Error('the aggregated tenant detail came back without any device counts');
    }
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
    await expect(deviceStateCount(page, 'accepted')).toHaveText(String(counts.accepted));
    await expect(deviceStateCount(page, 'total')).toHaveText(String(counts.total));
  });
});
