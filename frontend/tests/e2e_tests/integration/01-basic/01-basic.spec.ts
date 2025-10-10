// Copyright 2021 Northern.tech AS
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
import { expect } from '@playwright/test';

import test from '../../fixtures/fixtures';
import { emptyStorageState } from '../../utils/constants';

test.describe('Basic functionality checks', () => {
  test.use({ storageState: { ...emptyStorageState } });
  test.describe('basic window checks', () => {
    test('get the global window object', async ({ baseUrl, context, page }) => {
      page = await context.newPage();
      await page.goto(`${baseUrl}ui/`);
      const theWindow = await page.evaluate(() => window.innerWidth);
      expect(theWindow).toBeDefined();
    });
    test('get the document object', async ({ page }) => {
      const documentCharset = await page.evaluate(() => document.charset);
      expect(documentCharset).toBeDefined();
      expect(documentCharset).toEqual('UTF-8');
    });
    test('get the title', async ({ baseUrl, context, page }) => {
      page = await context.newPage();
      await page.goto(`${baseUrl}ui/`);
      await expect(page).toHaveTitle(/Mender/i);
    });
  });
});

test.describe('Overall layout and structure', () => {
  let navbar;
  test.beforeEach(async ({ page }) => {
    navbar = page.locator('.leftFixed.leftNav');
  });
  test('shows the left navigation', async () => {
    await expect(navbar.getByRole('link', { name: /Dashboard/i })).toBeVisible();
    await expect(navbar.getByRole('link', { name: /Devices/i })).toBeVisible();
    await expect(navbar.getByRole('link', { name: /Releases/i })).toBeVisible();
    await expect(navbar.getByRole('link', { name: /Deployments/i })).toBeVisible();
  });
  test('has clickable header buttons', async () => {
    await expect(navbar.getByRole('link', { name: /Dashboard/i })).toBeVisible();
    await navbar.getByRole('link', { name: /Dashboard/i }).click();
    await navbar.getByRole('link', { name: /Devices/i }).click();
    await navbar.getByRole('link', { name: /Releases/i }).click();
    await navbar.getByRole('link', { name: /Deployments/i }).click();
  });
});
