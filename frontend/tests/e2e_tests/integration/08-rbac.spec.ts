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
import test, { expect } from '../fixtures/fixtures.ts';
import { isEnterpriseOrStaging, prepareNewPage } from '../utils/commands.ts';
import { releaseTag, selectors, timeouts } from '../utils/constants.ts';

const releaseRoles = [
  { name: 'test-releases-role', permissions: ['Read'], tag: undefined },
  { name: `test-manage-${releaseTag}-role`, permissions: ['Manage'], tag: releaseTag },
  { name: `test-ro-${releaseTag}-role`, permissions: ['Read'], tag: releaseTag }
];

test.describe('RBAC functionality', () => {
  test.describe('configuration', () => {
    test.beforeEach(async ({ baseUrl, loggedInPage: page }) => {
      await page.goto(`${baseUrl}ui/settings`);
      await page.getByText(/Global settings/i).waitFor();
      await page.getByText(/user management/i).click();
      await page.goto(`${baseUrl}ui/settings/user-management`);
      const isVisible = await page.getByRole('button', { name: /new user/i }).isVisible();
      if (!isVisible) {
        console.log('settings may not be loaded - move around');
        await page.goto(`${baseUrl}ui/help`);
        await page.goto(`${baseUrl}ui/settings/user-management`);
      }
    });
    test('allows access to user management', async ({ loggedInPage: page }) => {
      await page.getByRole('button', { name: /new user/i }).isVisible();
    });
    test('allows role creation for static groups', async ({ environment, loggedInPage: page }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      await page.getByText(/roles/i).click();
      await page.getByRole('button', { name: 'Add a role' }).click();
      let nameInput = await page.locator('label:has-text("name") >> ..');
      const dialog = await nameInput.locator('.. >> .. >> ..');
      nameInput = await nameInput.locator('input');
      await nameInput.click();
      await nameInput.fill('test-groups-role');
      await nameInput.press('Tab');
      await dialog.locator('#role-description').fill('some description');
      await dialog.getByText('Search groups​').click({ force: true });
      // we need to check the entire page here, since the selection list is rendered in a portal, so likely outside
      // of the dialog tree
      await page.getByRole('option', { name: 'testgroup' }).click();
      await dialog.getByText('Select​').nth(1).click({ force: true });
      await page.getByText('Configure').click();
      await page.press('body', 'Escape');
      await dialog.getByRole('button', { name: /submit/i }).scrollIntoViewIfNeeded();
      await dialog.getByRole('button', { name: /submit/i }).click();
    });
    test('allows role creation for release tags', async ({ environment, loggedInPage: page }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      await page.getByText(/roles/i).click();
      for (const { name, permissions, tag } of releaseRoles) {
        await page.getByText('Add a role').click();
        let nameInput = await page.locator('label:has-text("name") >> ..');
        const dialog = await nameInput.locator('.. >> .. >> ..');
        nameInput = await nameInput.locator('input');
        await nameInput.click();
        await nameInput.fill(name);
        await nameInput.press('Tab');
        await dialog.locator('#role-description').fill('some description');
        // we need to check the entire page here, since the selection list is rendered in a portal, so likely outside
        // of the dialog tree
        await dialog.getByText('Search release tags​').click({ force: true });
        if (tag) {
          await page.getByRole('option', { name: tag }).click();
        } else {
          await page.getByRole('option', { name: /All releases/i }).click({ force: true });
        }
        await dialog.getByText('Select​').first().click({ force: true });
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
    test('allows user creation', async ({ environment, loggedInPage: page, password, username }) => {
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
  });
  test.describe('has working RBAC limitations for', () => {
    test('device groups', async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      const page = await prepareNewPage({ baseUrl, browser, password, username: `limited-${username}` });
      await page.getByRole('link', { name: /devices/i }).click({ force: true, timeout: timeouts.tenSeconds });
      await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
      // the created role does have permission to configure devices, so the section should be visible
      await page.getByText(/configuration/i).click();
      await page.getByText(/Device configuration/i).waitFor({ timeout: timeouts.tenSeconds });
    });
    test('read-only all releases', async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      const page = await prepareNewPage({ baseUrl, browser, password, username: `limited-ro-releases-${username}` });
      await page.getByRole('link', { name: /releases/i }).click({ force: true, timeout: timeouts.tenSeconds });
      // there should be multiple releases present
      await expect(page.getByText('1-2 of 2')).toBeVisible();
      // the created role doesn't have permission to upload artifacts, so the button shouldn't be visible
      await expect(page.getByRole('button', { name: /upload/i })).not.toBeVisible();
    });
    test('read-only tagged releases', async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      const page = await prepareNewPage({ baseUrl, browser, password, username: `limited-ro-${releaseTag}-${username}` });
      await page.getByRole('link', { name: /releases/i }).click({ force: true, timeout: timeouts.tenSeconds });
      // there should be only one release tagged with the releaseTag
      await expect(page.getByText('1-1 of 1')).toBeVisible();
      // the created role doesn't have permission to upload artifacts, so the button shouldn't be visible
      await expect(page.getByRole('button', { name: /upload/i })).not.toBeVisible();
    });
    test('manage tagged releases', async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      const page = await prepareNewPage({ baseUrl, browser, password, username: `limited-manage-${releaseTag}-${username}` });
      await page.getByRole('link', { name: /releases/i }).click({ force: true, timeout: timeouts.tenSeconds });
      // there should be only one release tagged with the releaseTag
      await expect(page.getByText('1-1 of 1')).toBeVisible();
      // the created role does have permission to upload artifacts, so the button should be visible
      await expect(page.getByRole('button', { name: /upload/i })).toBeVisible();
    });
  });
});
