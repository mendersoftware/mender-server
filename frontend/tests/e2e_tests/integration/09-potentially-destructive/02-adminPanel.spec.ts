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
import { adminPanelApiUrl, findSecondaryTenant, getTenantDetail, isAdminPanelAvailable, listTenants } from '../../utils/adminPanel';
import { login } from '../../utils/commands';

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
  test.beforeEach(async ({ baseUrl, browserName, environment, request }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
    test.skip(browserName !== 'chromium', 'tenant wide mutations only need to be exercised once');
    test.skip(!(await isAdminPanelAvailable(request, baseUrl)), 'the admin panel is not routed in this environment');
  });

  test('propagates a raised device limit through to the tenant', async ({ baseUrl, password, request }) => {
    const tenant = await findSecondaryTenant(request, baseUrl);
    const response = await request.put(adminPanelApiUrl(baseUrl, `tenants/${tenant.id}/device-limits`), {
      data: [{ name: 'max_devices', value: raisedDeviceLimit }]
    });
    expect(response.ok()).toBeTruthy();

    // the tenant's own session has to be told about the new ceiling, not just the operator
    const { token } = await login(secondaryTenantUser, password, baseUrl, request);
    const limitResponse = await request.get(`${baseUrl}api/management/v2/devauth/limits/max_devices`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    expect(limitResponse.ok()).toBeTruthy();
    const { limit } = await limitResponse.json();
    expect(limit).toEqual(raisedDeviceLimit);
  });

  test('locks a suspended tenant out and lets it back in again', async ({ baseUrl, password, request }) => {
    const tenant = await findSecondaryTenant(request, baseUrl);
    expect(await canLogIn(secondaryTenantUser, password, baseUrl, request)).toBeTruthy();

    const suspension = await request.put(adminPanelApiUrl(baseUrl, `tenants/${tenant.id}/status`), { data: { status: 'suspended' } });
    expect(suspension.ok()).toBeTruthy();

    try {
      const suspended = await getTenantDetail(request, baseUrl, tenant.id);
      expect(suspended.status).toEqual('suspended');
      // suspending is the operator's kill switch, it has to actually cut the tenant's users off
      expect(await canLogIn(secondaryTenantUser, password, baseUrl, request)).toBeFalsy();
    } finally {
      // leave the tenant usable for whatever runs next, even if the assertions above failed
      const reactivation = await request.put(adminPanelApiUrl(baseUrl, `tenants/${tenant.id}/status`), { data: { status: 'active' } });
      expect(reactivation.ok()).toBeTruthy();
    }

    const reactivated = await getTenantDetail(request, baseUrl, tenant.id);
    expect(reactivated.status).toEqual('active');
    expect(await canLogIn(secondaryTenantUser, password, baseUrl, request)).toBeTruthy();
  });

  test('moves a tenant onto a different plan', async ({ baseUrl, request }) => {
    const tenant = await findSecondaryTenant(request, baseUrl);
    expect(tenant.plan).toEqual('enterprise');

    try {
      const response = await request.put(adminPanelApiUrl(baseUrl, `tenants/${tenant.id}`), { data: { plan: 'professional' } });
      expect(response.ok()).toBeTruthy();

      const updated = await getTenantDetail(request, baseUrl, tenant.id);
      expect(updated.plan).toEqual('professional');
      // the plan filter the dashboard breakdown is built on has to pick the change up
      const { tenants: professional } = await listTenants(request, baseUrl, { per_page: 100, plan: 'professional' });
      expect(professional.map(({ id }) => id)).toContain(tenant.id);
    } finally {
      const reset = await request.put(adminPanelApiUrl(baseUrl, `tenants/${tenant.id}`), { data: { plan: tenant.plan } });
      expect(reset.ok()).toBeTruthy();
    }

    const restored = await getTenantDetail(request, baseUrl, tenant.id);
    expect(restored.plan).toEqual('enterprise');
  });
});
