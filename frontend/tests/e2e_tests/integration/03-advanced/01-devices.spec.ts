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
import * as fs from 'fs';
import md5 from 'md5';

import test, { expect } from '../../fixtures/fixtures';
import { isEnterpriseOrStaging } from '../../utils/commands';
import { expectedArtifactName, selectors, timeouts } from '../../utils/constants';

const fileName = `${expectedArtifactName}.mender`;
const rootfs = 'rootfs-image.version';

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
      await pendingMessage.click();
      await page.click(selectors.deviceListCheckbox);
      await page.click('.MuiSpeedDial-fab');
      await page.click('[aria-label="accept"]');
    }
    await page.locator(`input:near(:text("Status:"))`).first().click({ force: true });
    await page.click(`css=.MuiPaper-root >> text=/Accepted/i`);
    await page.waitForSelector(`css=${selectors.deviceListItem} >> text=/original/`, { timeout: 2 * timeouts.sixtySeconds });
    const element = await page.textContent(selectors.deviceListItem);
    expect(element.includes('original')).toBeTruthy();
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await page.getByText(/Device information for/i).waitFor();
    await expect(page.getByText('Authentication status')).toBeVisible();
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
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await page.getByText(/troubleshooting/i).click();
    // the deviceconnect connection might not be established right away
    await page.waitForSelector(`text=/Session status/i`, { timeout: timeouts.tenSeconds });
    await page.locator('.dropzone input').setInputFiles(`fixtures/${fileName}`);
    await page.click(selectors.placeholderExample, { clickCount: 3 });
    await page.getByPlaceholder(/installed-by-single-file/i).fill(`/tmp/${fileName}`);
    await page.getByRole('button', { name: /upload/i }).click();
    await page.getByText(/Upload successful/i).waitFor({ timeout: timeouts.fiveSeconds });
    await page.getByRole('tab', { name: /download/i }).click();
    await page.getByPlaceholder(/\/home\/mender/i).fill(`/tmp/${fileName}`);
    const [download] = await Promise.all([page.waitForEvent('download'), page.click('button:text("Download"):below(:text("file on the device"))')]);
    const downloadTargetPath = await download.path();
    const newFile = await fs.readFileSync(downloadTargetPath);
    const testFile = await fs.readFileSync(`fixtures/${fileName}`);
    expect(md5(newFile)).toEqual(md5(testFile));
  });

  test('can be found', async ({ demoDeviceName, page }) => {
    const searchField = await page.getByPlaceholder(/search devices/i);
    await searchField.fill(demoDeviceName);
    await page.waitForSelector(selectors.deviceListItem);
    const slideOut = await page.locator('.MuiPaper-root');
    await expect(slideOut.locator(`:text("${demoDeviceName}"):below(:text("clear search"))`)).toBeVisible();
    await expect(slideOut.getByText('1-1 of 1')).toBeVisible();
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await page.getByText(/device information/i).waitFor();
    await expect(page.getByText(/Authorization sets/i)).toBeVisible();
    await page.click('[aria-label="close"]');
    await expect(page.getByText(/table options/i)).toBeVisible();
    await page.getByText(/releases/i).click();
    await searchField.press('Enter');
    await expect(page.getByText(/device found/i)).toBeVisible();
  });

  test('can be filtered', async ({ browserName, demoDeviceName, page }) => {
    test.setTimeout(2 * timeouts.fifteenSeconds);
    await page.getByRole('button', { name: /filters/i }).click();
    await page.getByLabel(/attribute/i).fill(rootfs);
    const nameInput = await page.getByLabel(/value/i);
    await nameInput.fill(demoDeviceName);
    await page.waitForTimeout(timeouts.default);
    await nameInput.press('Enter');
    if (browserName === 'webkit') {
      await page.waitForTimeout(timeouts.fiveSeconds);
    }
    const filterChip = await page.getByRole('button', { name: `${rootfs} = ${demoDeviceName}` });
    await filterChip.waitFor({ timeout: timeouts.fiveSeconds });
    await expect(filterChip).toBeVisible();
    const resetButton = await page.getByRole('button', { name: /clear filter/i });
    await expect(resetButton).toBeEnabled();
    await page.waitForSelector(selectors.deviceListItem);
  });

  test('can be filtered into non-existence by numerical comparison', async ({ environment, page }) => {
    test.skip(!isEnterpriseOrStaging(environment), 'not available in OS');
    test.setTimeout(timeouts.fifteenSeconds);
    await page.getByRole('button', { name: /filters/i }).click();
    await page.getByText(/professional/i).waitFor({ state: 'hidden' }); // assume once the plan indicator tag is gone, filters can be used without problems
    await page.getByLabel(/attribute/i).fill('mem_total_kB');
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
    await page.getByRole('button', { name: /filters/i }).click();
    await page.getByLabel(/attribute/i).fill(rootfs);
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
    const pagination = await page.getByText('1-1');
    await pagination.waitFor({ timeout: timeouts.default });
    await expect(pagination).toBeVisible();
  });
});
