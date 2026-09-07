// Copyright 2025 Northern.tech AS
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
import type { Page } from '@playwright/test';
import * as fs from 'fs';
import md5 from 'md5';

import type { TestEnvironment } from '../../fixtures/fixtures';
import test, { expect } from '../../fixtures/fixtures';
import { isEnterpriseOrStaging } from '../../utils/commands';
import { expectedArtifactName, selectors, timeouts } from '../../utils/constants';

const fileName = `${expectedArtifactName}.mender`;
const rootfs = 'rootfs-image.version';
const macPattern = /^([0-9a-f]{2}:){5}[0-9a-f]{2}$/i;

// rely on `.MuiCollapse-entered` as signal for animation completion to ensure the filters are actionable to prevent flakiness in slower CI browsers
const openFilters = async (page: Page) => {
  await page.getByRole('button', { name: /filters/i }).click();
  await expect(page.locator('.filter-wrapper')).toHaveClass(/MuiCollapse-entered/);
};

const openDeviceDetails = (page: Page) => page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();

const skipUnlessTestDevicesAvailable = async ({ environment, page }: { environment: TestEnvironment; page: Page }) => {
  test.skip(environment !== 'staging', 'test devices are only available on hosted Mender');
  const features = await page.evaluate(() => (window as any).mender_environment?.features);
  test.skip(!features?.hasDeviceFlags, 'the device flags feature is not enabled on this deployment');
};

test.describe('Devices', () => {
  let navbar;
  test.beforeEach(async ({ browserName, page }) => {
    navbar = page.locator('.leftFixed.leftNav');
    await navbar.getByRole('link', { name: /Devices/i }).click({ force: browserName === 'webkit' });
  });

  test('can authorize a device', async ({ page }) => {
    // allow twice the device interaction time + roughly a regular test execution time
    test.setTimeout(2 * timeouts.sixtySeconds + timeouts.fifteenSeconds);
    let hasAcceptedDevice = false;
    try {
      await page.waitForSelector(`css=${selectors.deviceListItem}`, { timeout: timeouts.default });
      hasAcceptedDevice = await page.isVisible(selectors.deviceListItem);
    } catch {
      console.log(`no accepted device present so far`);
    }
    if (!hasAcceptedDevice) {
      const pendingMessage = await page.getByText(/pending authorization/i);
      await pendingMessage.waitFor({ timeout: timeouts.sixtySeconds });
      await pendingMessage
        .locator('..')
        .getByRole('button', { name: /view details/i })
        .click();
      await page.click(selectors.deviceListCheckbox);
      await page.click('.MuiSpeedDial-fab');
      await page.click('[aria-label="accept"]');
    }
    await page.locator(`input:near(:text("Status:"))`).first().click({ force: true });
    await page.click(`css=.MuiPaper-root >> text=/Accepted/i`);
    await page.waitForSelector(`css=${selectors.deviceListItem} >> text=/original/`, { timeout: 2 * timeouts.sixtySeconds });
    const element = await page.textContent(selectors.deviceListItem);
    expect(element.includes('original')).toBeTruthy();
    await openDeviceDetails(page);
    await page.getByText(/Device information for/i).waitFor();
    await expect(page.getByText('Authentication status')).toBeVisible();
  });

  test('is marked as test device once accepted by a trial account', async ({ environment, page }) => {
    await skipUnlessTestDevicesAvailable({ environment, page });
    await expect(page.getByRole('button', { name: /trial plan/i })).toBeVisible({ timeout: timeouts.tenSeconds });
    await openDeviceDetails(page);
    const testDeviceChip = page.getByRole('button', { name: 'Test device', exact: true });
    await expect(testDeviceChip).toBeVisible({ timeout: timeouts.tenSeconds });
    await testDeviceChip.click();
    await expect(page.locator('.MuiTooltip-tooltip').getByText(/1\/\d+ test devices set/)).toBeVisible({ timeout: timeouts.tenSeconds });
  });

  test('can have its test device status unset and set again', async ({ environment, page }) => {
    await skipUnlessTestDevicesAvailable({ environment, page });
    await openDeviceDetails(page);
    const testDeviceChip = page.getByRole('button', { name: 'Test device', exact: true });
    const deviceActions = page.getByRole('button', { name: 'device-actions' });
    const snackbar = page.getByText(/device updated successfully/i);
    await expect(testDeviceChip).toBeVisible({ timeout: timeouts.tenSeconds });

    await deviceActions.click();
    await expect(page.locator('[aria-label="remove-test-device"]')).toBeVisible();
    await expect(page.locator('[aria-label="set-test-device"]')).toHaveCount(0);
    await page.click('[aria-label="remove-test-device"]');
    const removalDialog = page.getByRole('dialog');
    await expect(removalDialog.getByRole('heading', { name: 'Remove as test device?' })).toBeVisible();
    await removalDialog.getByRole('button', { name: 'Remove as test device', exact: true }).click();
    await expect(snackbar).toBeVisible();
    await expect(testDeviceChip).not.toBeVisible();

    await deviceActions.click();
    await expect(page.locator('[aria-label="set-test-device"]')).toBeVisible();
    await expect(page.locator('[aria-label="remove-test-device"]')).toHaveCount(0);
    await page.click('[aria-label="set-test-device"]');
    const additionDialog = page.getByRole('dialog');
    await expect(additionDialog.getByRole('heading', { name: 'Set as test device?' })).toBeVisible();
    await additionDialog.getByRole('button', { name: 'Set as test device', exact: true }).click();
    await expect(snackbar).toBeVisible();
    await expect(testDeviceChip).toBeVisible();
  });

  test('can group a device', async ({ page }) => {
    const groupList = await page.locator('.grouplist');
    const wasGrouped = await groupList.getByText('testgroup').isVisible();
    test.skip(wasGrouped, 'looks like the device was grouped already, continue with the remaining tests');
    await page.click(selectors.deviceListCheckbox);
    await page.click('.MuiSpeedDial-fab');
    await page.click('[aria-label="group-add"]');
    await page.getByLabel(/type to create new/i).fill('testgroup');
    await page.click('.MuiDialogTitle-root');
    const groupCreation = await page.getByRole('button', { name: /create group/i });
    const groupExtension = await page.getByRole('button', { name: /add to group/i });
    await groupCreation.or(groupExtension).first().click();
    await groupList.getByText('testgroup').waitFor();
    await expect(groupList.getByText('testgroup')).toBeVisible();
    await groupList.getByText('All devices');
    await page.click(selectors.deviceListCheckbox);
    await groupList.getByText('testgroup').click();
    await expect(page.locator(`css=${selectors.deviceListItem} >> text=/original/`)).toBeVisible();
  });

  test('allows file transfer', async ({ browserName, environment, page }) => {
    // TODO adjust test to better work with webkit, for now it should be good enough to assume file transfers work there too if the remote terminal works
    test.skip(!isEnterpriseOrStaging(environment) || ['webkit'].includes(browserName));
    await openDeviceDetails(page);
    await page.getByText(/troubleshooting/i).click();
    // the deviceconnect connection might not be established right away
    await page.waitForSelector(`text=/Session status/i`, { timeout: timeouts.tenSeconds });
    await page.locator('.dropzone input').setInputFiles(`fixtures/${fileName}`);
    await page.click(selectors.placeholderExample, { clickCount: 3 });
    await page.getByPlaceholder(/installed-by-single-file/i).fill(`/tmp/${fileName}`);
    await page.getByRole('button', { name: 'Upload' }).click();
    await page.getByText(/Upload successful/i).waitFor({ timeout: timeouts.fiveSeconds });
    await page.getByRole('tab', { name: /download/i }).click();
    await page.getByPlaceholder(/\/home\/mender/i).fill(`/tmp/${fileName}`);
    await expect(async () => {
      console.log('trying to download...');
      const downloadPromise = page.waitForEvent('download', { timeout: timeouts.default });
      await page.click('button:text("Download"):below(:text("file on the device"))');
      const download = await downloadPromise;
      const failure = await download.failure();
      if (failure) {
        throw new Error(`Download failed: ${failure}`);
      }
      const downloadTargetPath = await download.path();
      const newFile = fs.readFileSync(downloadTargetPath);
      const testFile = fs.readFileSync(`fixtures/${fileName}`);
      expect(md5(newFile)).toEqual(md5(testFile));
    }).toPass({
      intervals: [timeouts.oneSecond, timeouts.fiveSeconds, timeouts.tenSeconds],
      timeout: 2 * timeouts.fifteenSeconds
    });
  });

  test('can be found', async ({ demoDeviceSoftware, page }) => {
    const features = await page.evaluate(() => (window as any).mender_environment?.features);
    test.skip(features?.hasNewSearch, 'the environment is configured to use new search enpoint');
    const searchField = await page.getByPlaceholder(/search devices/i);
    await searchField.fill(demoDeviceSoftware.slice(0, 6));
    await page.waitForSelector(selectors.deviceListItem);
    const slideOut = page.locator('.MuiPaper-root');
    await expect(slideOut.locator(`:text("${demoDeviceSoftware}")`)).toBeVisible();
    await expect(slideOut.getByText('1-1 of 1')).toBeVisible();
    await openDeviceDetails(page);
    await page.getByText(/device information/i).waitFor();
    await expect(page.getByText(/Authentication sets/i)).toBeVisible();
    await page.click('[aria-label="close"]');
    await expect(page.getByText(/table options/i)).toBeVisible();
    await page.locator('.leftFixed.leftNav').getByRole('link', { name: 'Software', exact: true }).click();
    await searchField.press('Enter');
    await expect(page.getByText(/device found/i)).toBeVisible();
  });

  test('can be found through the new search', async ({ page }) => {
    const features = await page.evaluate(() => (window as any).mender_environment?.features);
    test.skip(!features?.hasNewSearch, 'the environment is configured to use the legacy search');
    await page.waitForSelector(selectors.deviceListItem);
    const mac = await page.locator(selectors.deviceListItem).first().getByText(macPattern).innerText();

    await page.getByPlaceholder(/find a device/i).click();
    const searchDialog = page.getByRole('dialog');
    const searchField = searchDialog.getByPlaceholder(/starting with/i);
    await searchField.fill('nonExistentDevicePrefix');
    await expect(searchDialog.getByText(/no matching devices found/i)).toBeVisible({ timeout: timeouts.tenSeconds });
    await searchField.fill(mac.slice(0, 4));
    const result = searchDialog.getByText(mac);
    await result.waitFor({ timeout: timeouts.tenSeconds });
    await result.click();
    await expect(searchField).not.toBeVisible();
    await page.getByText(/device information/i).waitFor();
    await expect(page.locator('.expandedDevice')).toContainText(mac);
  });

  test('can be filtered', async ({ browserName, demoDeviceSoftware, page }) => {
    test.setTimeout(2 * timeouts.fifteenSeconds);
    await openFilters(page);
    await page.getByLabel(/attribute/i).fill(rootfs);
    const nameInput = await page.getByLabel(/value/i);
    await nameInput.fill(demoDeviceSoftware);
    await page.waitForTimeout(timeouts.default);
    await nameInput.press('Enter');
    if (browserName === 'webkit') {
      await page.waitForTimeout(timeouts.fiveSeconds);
    }
    const filterChip = await page.getByRole('button', { name: `${rootfs} = ${demoDeviceSoftware}` });
    await filterChip.waitFor({ timeout: timeouts.fiveSeconds });
    await expect(filterChip).toBeVisible();
    const resetButton = await page.getByRole('button', { name: /clear filter/i });
    await expect(resetButton).toBeEnabled();
    await page.waitForSelector(selectors.deviceListItem);
  });

  test('can be filtered into non-existence by numerical comparison', async ({ environment, page }) => {
    test.skip(!isEnterpriseOrStaging(environment), 'not available in OS');
    test.setTimeout(timeouts.fifteenSeconds);
    await openFilters(page);
    await page.getByText(/professional/i).waitFor({ state: 'hidden' }); // assume once the plan indicator tag is gone, filters can be used without problems
    await page.getByLabel(/attribute/i).fill('mem_total_kB');
    await page.keyboard.press('Enter');
    await page.getByText(/equals/i).click();
    await page.waitForTimeout(timeouts.default);
    await page.getByRole('option', { name: '>', exact: true }).click();
    const nameInput = await page.getByLabel(/value/i);
    await nameInput.fill('1000000000');
    await page.waitForTimeout(timeouts.default);
    await nameInput.press('Enter');
    await page.getByText('No devices found').waitFor({ timeout: timeouts.fiveSeconds });
  });

  test('can be filtered into non-existence', async ({ environment, page }) => {
    test.skip(!isEnterpriseOrStaging(environment), 'not available in OS');
    test.setTimeout(2 * timeouts.fifteenSeconds);
    await openFilters(page);
    await page.getByLabel(/attribute/i).fill(rootfs);
    await page.keyboard.press('Enter');
    await page.getByText(/equals/i).click();
    await page.waitForTimeout(timeouts.default);
    await page.getByRole('option', { name: `doesn't exist`, exact: true }).click();
    await page.getByRole('button', { name: /Add rule/i }).waitFor();
    await page.getByRole('button', { name: /Add rule/i }).click();
    await expect(page.getByRole('button', { name: `${rootfs} doesn't exist` })).toBeVisible();
    await page.getByText('No devices found').waitFor({ timeout: timeouts.fiveSeconds });
    await expect(page.getByText('No devices found')).toBeVisible();
    await page.getByText(/clear filter/i).click();
    await page.waitForSelector(selectors.deviceListItem);
    const pagination = await page.getByText('1-1 of 1');
    await pagination.waitFor({ timeout: timeouts.default });
    await expect(pagination).toBeVisible();
  });
});
