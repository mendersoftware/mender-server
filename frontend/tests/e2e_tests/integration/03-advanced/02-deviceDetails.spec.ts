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

import test, { expect } from '../../fixtures/fixtures';
import { compareImages } from '../../utils/commands';
import { selectors, timeouts } from '../../utils/constants';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const terminalReferenceFileMap = {
  default: 'terminalContent.png',
  webkit: 'terminalContent-webkit.png'
};

test.describe('Device details', () => {
  test.beforeEach(async ({ baseUrl, page }) => {
    await page.goto(`${baseUrl}ui/devices`);
  });
  test('has basic inventory', async ({ demoDeviceName, page }) => {
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await page.getByText(/inventory/i).click();
    const expandedDevice = await page.locator(`css=.expandedDevice`);
    await expect(expandedDevice.getByText('Linux')).toBeVisible();
    await expect(expandedDevice.getByText(/hostname/).first()).toBeVisible();
    await expandedDevice.getByRole('tab', { name: /software/i }).click();
    await expect(expandedDevice.getByText(demoDeviceName)).toBeVisible();
  });

  test.describe('Terminal interactions', () => {
    test.afterEach(async ({ page }) => {
      const terminalTextInput = await page.locator(selectors.terminalText);
      await terminalTextInput.fill('exit');
      await terminalTextInput.press('Enter');
    });
    test('can open a terminal', async ({ browserName, page }) => {
      await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
      await page.getByText(/troubleshooting/i).click();
      // the deviceconnect connection might not be established right away + polling interval is 10s on device details
      await page.getByText(/Session status/i).waitFor({ timeout: 3 * timeouts.tenSeconds });
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
      const elementHandle = await page.locator(selectors.terminalElement);
      expect(elementHandle).toBeTruthy();
      if (['chromium', 'webkit'].includes(browserName)) {
        // this should ensure a repeatable position across test runners
        await page.locator('.MuiDrawer-paper').hover();
        await page.mouse.wheel(0, -100);
        await elementHandle.scrollIntoViewIfNeeded();

        const screenShotPath = path.join(__dirname, '..', 'test-results', 'diffs', 'terminalContent-actual.png');
        await elementHandle.screenshot({ path: screenShotPath });

        const expectedPath = path.join(__dirname, '..', '..', 'fixtures', terminalReferenceFileMap[browserName] ?? terminalReferenceFileMap.default);
        const { pass } = compareImages(expectedPath, screenShotPath);
        expect(pass).toBeTruthy();

        const terminalTextInput = await page.locator(selectors.terminalText);
        await terminalTextInput.fill('top');
        await terminalTextInput.press('Enter');
        await page.waitForTimeout(timeouts.default);

        await elementHandle.screenshot({ path: screenShotPath });
        const { pass: pass2 } = compareImages(expectedPath, screenShotPath);
        expect(pass2).not.toBeTruthy();
        await terminalTextInput.press('q');
      }
    });
  });

  test('can trigger on device updates', async ({ page }) => {
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await page.getByText(/troubleshooting/i).click();
    // the deviceconnect connection might not be established right away
    await page.getByText(/Session status/i).waitFor({ timeout: timeouts.tenSeconds });
    const connectionButton = await page.getByRole('button', { name: /connect/i });
    await connectionButton.first().click();
    await page.getByText('Connection with the device established').waitFor({ timeout: timeouts.tenSeconds });
    const quickActionMenu = await page.getByText(/quick commands/i);
    await quickActionMenu.scrollIntoViewIfNeeded();
    await quickActionMenu.click();
    const updateLoadingIndicator = page.locator('li .miniLoaderContainer');
    await expect(updateLoadingIndicator).not.toBeVisible();
    await page.getByRole('menuitem', { name: 'Trigger update check' }).click();
    await expect(updateLoadingIndicator).toBeVisible();
    await updateLoadingIndicator.waitFor({ state: 'hidden', timeout: timeouts.tenSeconds });
    await expect(updateLoadingIndicator).not.toBeVisible();
  });
});
