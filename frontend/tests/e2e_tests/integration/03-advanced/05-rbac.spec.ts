// Copyright 2022 Northern.tech AS
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
import { v4 as uuid } from 'uuid';

import test, { expect } from '../../fixtures/fixtures';
import { isEnterpriseOrStaging, prepareNewPage } from '../../utils/commands';
import { releaseTag, selectors, timeouts } from '../../utils/constants';

const releaseRoles = [
  { name: 'test-releases-role', permissions: ['Read'], tag: undefined },
  { name: `test-manage-${releaseTag}-role`, permissions: ['Manage'], tag: releaseTag },
  { name: `test-ro-${releaseTag}-role`, permissions: ['Read'], tag: releaseTag }
];

test.describe('RBAC functionality', () => {
  test.beforeEach(async ({ baseUrl, environment, page }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    await page.goto(`${baseUrl}ui/settings`);
    await page.getByText(/Global settings/i).waitFor();
    await page.getByText(/user management/i).click();
    const isVisible = await page.getByRole('button', { name: /new user/i }).isVisible();
    if (!isVisible) {
      console.log('settings may not be loaded - move around');
      await page.goto(`${baseUrl}ui/help`);
      await page.goto(`${baseUrl}ui/settings/user-management`);
    }
  });
  test.describe('configuration', () => {
    test('allows role creation for static groups', async ({ page }) => {
      await page.getByText(/roles/i).click();
      await page.getByRole('button', { name: 'Add a role' }).click();
      const dialog = await page.locator('.MuiPaper-root >> form');
      const nameInput = await dialog.getByLabel('name');
      await nameInput.click();
      await nameInput.fill('test-groups-role');
      await nameInput.press('Tab');
      await dialog.getByLabel(/description/i).fill('some description');
      await dialog.getByLabel(/Search groups/i).click({ force: true });
      // we need to check the entire page here, since the selection list is rendered in a portal, so likely outside
      // of the dialog tree
      await page.getByRole('option', { name: 'testgroup' }).click();
      await dialog.locator(`[id="mui-component-select-groups.0.uiPermissions"]`).click();
      await page.getByText('Configure').click();
      await page.press('body', 'Escape');
      await dialog.getByRole('button', { name: /submit/i }).scrollIntoViewIfNeeded();
      await dialog.getByRole('button', { name: /submit/i }).click();
      await page.getByText(/role was created/i).waitFor();
    });
    test('allows role creation for release tags', async ({ page }) => {
      await page.getByText(/roles/i).click();
      for (const { name, permissions, tag } of releaseRoles) {
        await page.getByText('Add a role').click();
        const dialog = await page.locator('.MuiPaper-root >> form');
        const nameInput = await dialog.getByLabel('name');
        await nameInput.click();
        await nameInput.fill(name);
        await nameInput.press('Tab');
        // we need to check the entire page here, since the selection list is rendered in a portal, so likely outside
        // of the dialog tree
        await dialog.getByLabel(/Search release tags/i).click({ force: true });
        if (tag) {
          await page.getByRole('option', { name: tag }).click();
        } else {
          await page.getByRole('option', { name: /All releases/i }).click({ force: true });
        }
        await dialog.locator(`[id="mui-component-select-releases.0.uiPermissions"]`).click();
        for await (const permission of permissions) {
          await page.getByRole('option', { name: permission }).click();
        }
        await page.press('body', 'Escape');
        await dialog.getByRole('button', { name: /submit/i }).scrollIntoViewIfNeeded();
        await dialog.getByRole('button', { name: /submit/i }).click();
        await page.getByText('The role was created successfully.').waitFor();
        await page.waitForTimeout(timeouts.default);
      }
    });
  });
  test('TEMP: user creation timeout stress probe', async ({ environment, page, password, username }) => {
    // This temporary test attempts to reproduce the recurring hang in the user creation endpoint.
    test.setTimeout(3 * timeouts.sixtySeconds);
    for (let i = 0; i < 10; i++) {
      const usernamePrefixed = `${uuid()}-${username}`;

      const userCreations = [
        { user: `limited-${usernamePrefixed}`, role: 'test-groups-role' },
        { user: `lim-ro-rel-${usernamePrefixed}`, role: releaseRoles[0].name },
        { user: `lim-man-${releaseTag}-${usernamePrefixed}`, role: releaseRoles[1].name },
        { user: `lim-ro-${releaseTag}-${usernamePrefixed}`, role: releaseRoles[2].name }
      ];
      for (const { user, role } of userCreations) {
        await page.getByRole('button', { name: /new user/i }).click();
        await page.getByPlaceholder(/email/i).click();
        await page.getByPlaceholder(/email/i).fill(user);
        await page.getByPlaceholder(/Password/i).click();
        await page.getByPlaceholder(/Password/i).fill(password);
        if (isEnterpriseOrStaging(environment)) {
          await page.getByRole('combobox', { name: /admin/i }).click();
          // first we need to deselect the default admin role
          await page.getByRole('option', { name: 'Admin' }).click();
          await page.getByRole('option', { name: role }).scrollIntoViewIfNeeded();
          await page.getByRole('option', { name: role }).click();
          await page.press('body', 'Escape');
        }
        await page.getByText(/Create user/i).click();
        await page.getByText('The user was created successfully.').waitFor();
      }
      for (let j = 0; j < userCreations.length; j++) {
        await page
          .getByRole('button', { name: /view details/i })
          .nth(2)
          .click();
        await page.getByRole('button', { name: /delete user/i }).click();
        await page.getByRole('button', { name: /delete user/i }).click();
      }
    }
  });
  test('allows user creation', async ({ environment, page, password, username }) => {
    const userCreations = [
      { user: `limited-${username}`, role: 'test-groups-role' },
      { user: `limited-ro-releases-${username}`, role: releaseRoles[0].name },
      { user: `limited-manage-${releaseTag}-${username}`, role: releaseRoles[1].name },
      { user: `limited-ro-${releaseTag}-${username}`, role: releaseRoles[2].name }
    ];
    for (const { user, role } of userCreations) {
      await page.getByRole('button', { name: /new user/i }).click();
      await page.getByPlaceholder(/email/i).click();
      await page.getByPlaceholder(/email/i).fill(user);
      await page.getByPlaceholder(/Password/i).click();
      await page.getByPlaceholder(/Password/i).fill(password);
      if (isEnterpriseOrStaging(environment)) {
        await page.getByRole('combobox', { name: /admin/i }).click();
        // first we need to deselect the default admin role
        await page.getByRole('option', { name: 'Admin' }).click();
        await page.getByRole('option', { name: role }).scrollIntoViewIfNeeded();
        await page.getByRole('option', { name: role }).click();
        await page.press('body', 'Escape');
      }
      await page.getByText(/Create user/i).click();
      await page.getByText('The user was created successfully.').waitFor();
      await page.waitForTimeout(timeouts.default);
    }
  });
  test.describe('has working RBAC limitations for', () => {
    let navbar;
    test.beforeEach(async ({ page }) => {
      navbar = page.locator('.leftFixed.leftNav');
    });
    test('device groups', async ({ baseUrl, browser, password, request, username }) => {
      const page = await prepareNewPage({ baseUrl, browser, password, request, username: `limited-${username}` });
      const navigationButton = page.getByRole('link', { name: /devices/i });
      await navigationButton.waitFor({ timeout: timeouts.tenSeconds });
      await navigationButton.click({ force: true });
      await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
      // the created role does have permission to configure devices, so the section should be visible
      await page.getByText(/configuration/i).click();
      await page.getByText(/Device configuration/i).waitFor({ timeout: timeouts.tenSeconds });
      await page.context().close();
    });
    test('read-only all releases', async ({ baseUrl, browser, password, request, username }) => {
      const page = await prepareNewPage({ baseUrl, browser, password, request, username: `limited-ro-releases-${username}` });
      const navigationButton = navbar.getByRole('link', { name: 'Releases', exact: true });
      await navigationButton.waitFor({ timeout: timeouts.tenSeconds });
      await navigationButton.click({ force: true });
      // there should be multiple releases present
      await expect(page.getByText('1-2 of 2')).toBeVisible();
      // the created role doesn't have permission to upload artifacts, so the button shouldn't be visible
      await expect(page.getByRole('button', { name: /upload/i })).not.toBeVisible();
      await page.getByRole('checkbox').first().click();
      await expect(page.getByLabel(/release-actions/i)).not.toBeVisible();
      await page.context().close();
    });
    test('read-only tagged releases', async ({ baseUrl, browser, password, request, username }) => {
      const page = await prepareNewPage({ baseUrl, browser, password, request, username: `limited-ro-${releaseTag}-${username}` });
      const navigationButton = navbar.getByRole('link', { name: 'Releases', exact: true });
      await navigationButton.waitFor({ timeout: timeouts.tenSeconds });
      await navigationButton.click({ force: true });
      // there should be only one release tagged with the releaseTag
      await expect(page.getByText('1-1 of 1')).toBeVisible();
      // the created role doesn't have permission to upload artifacts, so the button shouldn't be visible
      await expect(page.getByRole('button', { name: /upload/i })).not.toBeVisible();
      await page.context().close();
    });
    test('manage tagged releases', async ({ baseUrl, browser, password, request, username }) => {
      const page = await prepareNewPage({ baseUrl, browser, password, request, username: `limited-manage-${releaseTag}-${username}` });
      const navigationButton = navbar.getByRole('link', { name: 'Releases', exact: true });
      await navigationButton.waitFor({ timeout: timeouts.tenSeconds });
      await navigationButton.click({ force: true });
      // there should be only one release tagged with the releaseTag
      await expect(page.getByText('1-1 of 1')).toBeVisible();
      // the created role does have permission to upload artifacts, so the button should be visible
      await expect(page.getByRole('button', { name: /upload/i })).toBeVisible();
      await page.context().close();
    });
  });
});
