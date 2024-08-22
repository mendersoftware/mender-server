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
import * as path from 'path';
import { fileURLToPath } from 'url';

import test, { expect } from '../fixtures/fixtures.ts';
import { compareImages, isEnterpriseOrStaging } from '../utils/commands.ts';
import { selectors, storagePath, timeouts } from '../utils/constants.ts';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const terminalReferenceFileMap = {
  default: 'terminalContent.png',
  webkit: 'terminalContent-webkit.png'
};

const rootfs = 'rootfs-image.version';

test.describe('Device details', () => {
  test.use({ storageState: storagePath });

  test('has basic inventory', async ({ demoDeviceName, loggedInPage: page }) => {
    await page.click(`.leftNav :text('Devices')`);
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await page.getByText(/inventory/i).click();
    await expect(page.locator(`css=.expandedDevice >> text=Linux`)).toBeVisible();
    await expect(page.locator(`css=.expandedDevice >> text=mac`)).toBeVisible();
    await expect(page.locator(`css=.expandedDevice >> text=${demoDeviceName}`)).toBeVisible();
  });

  test('can be found', async ({ demoDeviceName, loggedInPage: page }) => {
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

  test('can be filtered', async ({ browserName, demoDeviceName, loggedInPage: page }) => {
    test.setTimeout(2 * timeouts.fifteenSeconds);
    await page.click(`.leftNav :text('Devices')`);
    await page.getByRole('button', { name: /filters/i }).click();
    await page.getByLabel(/attribute/i).fill(rootfs);
    const nameInput = await page.getByLabel(/value/i);
    await nameInput.fill(demoDeviceName);
    await page.waitForTimeout(timeouts.oneSecond);
    await nameInput.press('Enter');
    if (browserName === 'webkit') {
      await page.waitForTimeout(timeouts.fiveSeconds);
    }
    const filterChip = await page.getByRole('button', { name: `${rootfs} = ${demoDeviceName}` });
    await filterChip.waitFor({ timeout: timeouts.fiveSeconds });
    await expect(filterChip).toBeVisible();
    await page.waitForSelector(selectors.deviceListItem);
  });

  test('can be filtered into non-existence', async ({ environment, loggedInPage: page }) => {
    test.skip(!isEnterpriseOrStaging(environment), 'not available in OS');
    test.setTimeout(2 * timeouts.fifteenSeconds);
    await page.click(`.leftNav :text('Devices')`);
    await page.getByRole('button', { name: /filters/i }).click();
    await page.getByLabel(/attribute/i).fill(rootfs);
    await page.getByText(/equals/i).click();
    await page.waitForTimeout(timeouts.default);
    await page.getByRole('option', { name: `doesn't exist` }).click();
    const filterChip = await page.getByRole('button', { name: `${rootfs} doesn't exist` });
    await filterChip.waitFor({ timeout: timeouts.fiveSeconds });
    await expect(filterChip).toBeVisible();
    await expect(page.getByText('No devices found')).toBeVisible();
    await page.getByText(/clear filter/i).click();
    await page.waitForSelector(selectors.deviceListItem);
    const pagination = await page.getByText('1-1');
    await pagination.waitFor({ timeout: timeouts.default });
    await expect(pagination).toBeVisible();
  });

  test('can open a terminal', async ({ browserName, loggedInPage: page }) => {
    await page.click(`.leftNav :text('Devices')`);
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await page.getByText(/troubleshooting/i).click();
    // the deviceconnect connection might not be established right away
    await page.getByText(/Session status/i).waitFor({ timeout: timeouts.tenSeconds });
    const connectionButton = await page.getByRole('button', { name: /connect/i });
    await connectionButton.first().click();
    await page.getByText('Connection with the device established').waitFor({ timeout: timeouts.tenSeconds });
    await expect(page.locator('.terminal.xterm .xterm-screen')).toBeVisible();

    // the terminal content might take a bit to get painted - thus the waiting
    await page.click(selectors.terminalElement, { timeout: timeouts.default });

    // the terminal content differs a bit depending on the device id, thus the higher threshold allowed
    // NB! without the screenshot-name argument the options don't seem to be applied
    // NB! screenshots should only be taken by running the docker composition (as in CI) - never in open mode,
    // as the resizing option on `allowSizeMismatch` only pads the screenshot with transparent pixels until
    // the larger size is met (when diffing screenshots of multiple sizes) and does not scale to fit!
    const elementHandle = await page.$(selectors.terminalElement);
    expect(elementHandle).toBeTruthy();
    if (['chromium', 'webkit'].includes(browserName)) {
      // this should ensure a repeatable position across test runners
      await page.locator('.MuiDrawer-paper').hover();
      await page.mouse.wheel(0, -100);
      await elementHandle.scrollIntoViewIfNeeded();

      const screenShotPath = path.join(__dirname, '..', 'test-results', 'diffs', 'terminalContent-actual.png');
      await elementHandle.screenshot({ path: screenShotPath });

      const expectedPath = path.join(__dirname, '..', 'fixtures', terminalReferenceFileMap[browserName] ?? terminalReferenceFileMap.default);
      const { pass } = compareImages(expectedPath, screenShotPath);
      expect(pass).toBeTruthy();

      const terminalText = await page.locator(`css=${selectors.terminalText}`);
      await terminalText.fill('top');
      await page.keyboard.press('Enter');
      await page.waitForTimeout(timeouts.default);

      await elementHandle.screenshot({ path: screenShotPath });
      const { pass: pass2 } = compareImages(expectedPath, screenShotPath);
      expect(pass2).not.toBeTruthy();
    }
  });
});
