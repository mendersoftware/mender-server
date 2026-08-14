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
import { findSecondaryTenant } from '../../utils/adminPanel';

test.describe('Admin panel navigation', () => {
  test.beforeEach(async ({ environment }) => {
    test.skip(environment !== 'enterprise', 'the admin panel is only deployed alongside enterprise installations');
  });

  test('opens on the dashboard', async ({ adminBaseUrl, page }) => {
    await page.goto(adminBaseUrl);
    await expect(page.getByText('Mender Admin Panel')).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
  });

  test('moves between the sections through the drawer', async ({ adminBaseUrl, page }) => {
    await page.goto(adminBaseUrl);
    const sections = [
      { heading: 'Tenants', nav: 'Tenants', path: 'tenants' },
      { heading: 'Users', nav: 'Users', path: 'users' },
      { heading: 'Send Notice', nav: 'Notices', path: 'notices' }
    ];
    for (const { heading, nav, path } of sections) {
      await page.getByRole('button', { name: nav, exact: true }).click();
      await expect(page.getByRole('heading', { name: heading })).toBeVisible();
      // the panel is a client side router, the address bar has to keep up with it
      await expect(page).toHaveURL(`${adminBaseUrl}${path}`);
    }
    await page.getByRole('button', { name: 'Dashboard', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
  });

  test('serves a deep linked tenant and finds its way back', async ({ adminBaseUrl, page, request }) => {
    const tenant = await findSecondaryTenant(request, adminBaseUrl);
    // a reload on a nested route has to be served by the app rather than 404 from the web server
    await page.goto(`${adminBaseUrl}tenants/${tenant.id}`);
    await expect(page.getByRole('heading', { name: `Tenant: ${tenant.name}` })).toBeVisible();

    await page.getByRole('link', { name: 'Back to Tenants' }).click();
    await expect(page.getByRole('heading', { name: 'Tenants' })).toBeVisible();
    await expect(page).toHaveURL(`${adminBaseUrl}tenants`);
  });
});
