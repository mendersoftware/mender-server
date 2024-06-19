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
import test, { expect } from '../fixtures/fixtures';
import { isEnterpriseOrStaging, isLoggedIn, processLoginForm } from '../utils/commands';
import { releaseTag, selectors, timeouts } from '../utils/constants';

const releaseRoles = [
  { name: 'test-releases-role', permissions: ['Read'], tag: undefined },
  { name: `test-manage-${releaseTag}-role`, permissions: ['Manage'], tag: releaseTag },
  { name: `test-ro-${releaseTag}-role`, permissions: ['Read'], tag: releaseTag }
];

test.describe('RBAC functionality', () => {
  test.beforeEach(async ({ baseUrl, loggedInPage: page }) => {
    await page.goto(`${baseUrl}ui/settings`);
    await page.waitForSelector('text=/Global settings/i');
    await page.click('text=/user management/i');
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
    await page.click('text=/roles/i');
    await page.click('text=Add a role');
    let nameInput = await page.locator('label:has-text("name") >> ..');
    const dialog = await nameInput.locator('.. >> .. >> ..');
    nameInput = await nameInput.locator('input');
    await nameInput.click();
    await nameInput.fill('test-groups-role');
    await nameInput.press('Tab');
    await dialog.locator('#role-description').fill('some description');
    await dialog.locator('text=Search groups​').click({ force: true });
    // we need to check the entire page here, since the selection list is rendered in a portal, so likely outside
    // of the dialog tree
    await page.locator('li[role="option"]:has-text("testgroup")').click();
    await dialog.locator('text=Select​').nth(1).click({ force: true });
    await page.locator('text=Configure').click();
    await page.press('body', 'Escape');
    await dialog.getByRole('button', { name: /submit/i }).scrollIntoViewIfNeeded();
    await dialog.getByRole('button', { name: /submit/i }).click();
  });

  test('allows role creation for release tags', async ({ environment, loggedInPage: page }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    await page.click('text=/roles/i');
    for (const { name, permissions, tag } of releaseRoles) {
      await page.click('text=Add a role');
      let nameInput = await page.locator('label:has-text("name") >> ..');
      const dialog = await nameInput.locator('.. >> .. >> ..');
      nameInput = await nameInput.locator('input');
      await nameInput.click();
      await nameInput.fill(name);
      await nameInput.press('Tab');
      await dialog.locator('#role-description').fill('some description');
      // we need to check the entire page here, since the selection list is rendered in a portal, so likely outside
      // of the dialog tree
      await dialog.locator('text=Search release tags​').click({ force: true });
      if (tag) {
        await page.locator(`li[role="option"]:has-text("${tag}")`).click();
      } else {
        await page.locator(`li[role="option"]:has-text("All releases")`).click({ force: true });
      }
      await dialog.locator('text=Select​').first().click({ force: true });
      await Promise.all(permissions.map(async permission => await page.locator(`li[role="option"]:has-text("${permission}")`).click()));
      await page.press('body', 'Escape');
      await dialog.getByRole('button', { name: /submit/i }).scrollIntoViewIfNeeded();
      await dialog.getByRole('button', { name: /submit/i }).click();
      await page.waitForSelector('text=The role was created successfully.');
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
        await page.getByRole('option', { name: role }).click();
        await page.press('body', 'Escape');
      }
      await page.click(`text=/Create user/i`);
      await page.waitForSelector('text=The user was created successfully.');
    }
  });

  test('has working RBAC groups limitations', async ({ baseUrl, browser, environment, password, username }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    const context = await browser.newContext();
    const page = await context.newPage();
    await page.goto(baseUrl);
    // enter valid username and password of the new user
    await processLoginForm({ username: `limited-${username}`, password, page, environment });
    await isLoggedIn(page);
    await page.click(`.leftNav :text('Devices')`);
    await page.click(`${selectors.deviceListItem} div:last-child`);
    // the created role does have permission to configure devices, so the section should be visible
    await page.click(`text=/configuration/i`);
    await page.waitForSelector('text=/Device configuration/i', { timeout: timeouts.tenSeconds });
  });

  test.describe('has working RBAC release limitations', () => {
    test('read-only all releases', async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      const context = await browser.newContext();
      const page = await context.newPage();
      await page.goto(baseUrl);
      // enter valid username and password of the new user
      await processLoginForm({ username: `limited-ro-releases-${username}`, password, page, environment });
      await isLoggedIn(page);
      await page.getByText(/releases/i).click({ timeout: timeouts.tenSeconds });
      // there should be multiple releases present
      expect(await page.getByText('1-2 of 2')).toBeVisible();
      // the created role doesn't have permission to upload artifacts, so the button shouldn't be visible
      expect(await page.getByRole('button', { name: /upload/i })).not.toBeVisible();
    });
    test('read-only tagged releases', async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      const context = await browser.newContext();
      const page = await context.newPage();
      await page.goto(baseUrl);
      // enter valid username and password of the new user
      await processLoginForm({ username: `limited-ro-${releaseTag}-${username}`, password, page, environment });
      await isLoggedIn(page);
      await page.getByText(/releases/i).click({ timeout: timeouts.tenSeconds });
      // there should be only one release tagged with the releaseTag
      expect(await page.getByText('1-1 of 1')).toBeVisible();
      // the created role doesn't have permission to upload artifacts, so the button shouldn't be visible
      expect(await page.getByRole('button', { name: /upload/i })).not.toBeVisible();
    });
    test('manage tagged releases', async ({ baseUrl, browser, environment, password, username }) => {
      test.skip(!isEnterpriseOrStaging(environment));
      const context = await browser.newContext();
      const page = await context.newPage();
      await page.goto(baseUrl);
      // enter valid username and password of the new user
      await processLoginForm({ username: `limited-manage-${releaseTag}-${username}`, password, page, environment });
      await isLoggedIn(page);
      await page.getByText(/releases/i).click({ timeout: timeouts.tenSeconds });
      // there should be only one release tagged with the releaseTag
      expect(await page.getByText('1-1 of 1')).toBeVisible();
      // the created role does have permission to upload artifacts, so the button should be visible
      expect(await page.getByRole('button', { name: /upload/i })).toBeVisible();
    });
  });
});
