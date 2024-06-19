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

import test, { expect } from '../fixtures/fixtures';
import { compareImages, isEnterpriseOrStaging } from '../utils/commands';
import { selectors, storagePath, timeouts } from '../utils/constants';

const terminalReferenceFileMap = {
  default: 'terminalContent.png',
  webkit: 'terminalContent-webkit.png'
};

const rootfs = 'rootfs-image.version';

test.describe('Device details', () => {
  test.use({ storageState: storagePath });

  test('has basic inventory', async ({ demoDeviceName, loggedInPage: page }) => {
    await page.click(`.leftNav :text('Devices')`);
    await page.click(`${selectors.deviceListItem} div:last-child`);
    await page.click(`text=/inventory/i`);
    expect(await page.isVisible(`css=.expandedDevice >> text=Linux`)).toBeTruthy();
    expect(await page.isVisible(`css=.expandedDevice >> text=mac`)).toBeTruthy();
    expect(await page.isVisible(`css=.expandedDevice >> text=${demoDeviceName}`)).toBeTruthy();
  });

  test('can be found', async ({ demoDeviceName, loggedInPage: page }) => {
    const searchField = await page.getByPlaceholder(/search devices/i);
    await searchField.fill(demoDeviceName);
    await page.waitForSelector(selectors.deviceListItem);
    const slideOut = await page.locator('.MuiPaper-root');
    expect(await slideOut.locator(`:text("${demoDeviceName}"):below(:text("clear search"))`).isVisible()).toBeTruthy();
    expect(await slideOut.getByText('1-1 of 1').isVisible()).toBeTruthy();
    await page.click(selectors.deviceListItem);
    await page.waitForSelector('text=/device information/i');
    expect(await page.getByText(/Authorization sets/i).isVisible()).toBeTruthy();
    await page.click('[aria-label="close"]');
    expect(await page.getByText(/table options/i).isVisible()).toBeTruthy();
    await page.getByText(/releases/i).click();
    await searchField.press('Enter');
    expect(await page.getByText(/device found/i).isVisible()).toBeTruthy();
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
    expect(await page.getByRole('button', { name: `${rootfs} = ${demoDeviceName}` }).isVisible({ timeout: timeouts.default })).toBeTruthy();
    await page.waitForSelector(selectors.deviceListItem);
  });

  test('can be filtered into non-existence', async ({ environment, loggedInPage: page }) => {
    test.skip(!isEnterpriseOrStaging(environment), 'not available in OS');
    test.setTimeout(2 * timeouts.fifteenSeconds);
    await page.click(`.leftNav :text('Devices')`);
    await page.getByRole('button', { name: /filters/i }).click();
    await page.getByLabel(/attribute/i).fill(rootfs);
    await page.getByText(/equals/i).click();
    await page.getByText(`doesn't exist`).click();
    await page.waitForTimeout(timeouts.fiveSeconds);
    expect(await page.getByRole('button', { name: `${rootfs} doesn't exist` }).isVisible()).toBeTruthy();
    expect(await page.getByText('No devices found').isVisible()).toBeTruthy();
    await page.getByText(/clear filter/i).click();
    await page.waitForSelector(selectors.deviceListItem);
    const paginationVisible = await page.getByText('1-1').isVisible({ timeout: timeouts.default });
    expect(paginationVisible).toBeTruthy();
  });

  test('can open a terminal', async ({ browserName, loggedInPage: page }) => {
    await page.click(`.leftNav :text('Devices')`);
    await page.click(`${selectors.deviceListItem} div:last-child`);
    await page.click(`text=/troubleshooting/i`);
    // the deviceconnect connection might not be established right away
    await page.waitForSelector('text=/Session status/i', { timeout: timeouts.tenSeconds });
    const connectionButton = await page.getByRole('button', { name: /connect/i });
    await connectionButton.first().click();
    await page.waitForSelector(`text=Connection with the device established`, { timeout: timeouts.tenSeconds });
    expect(await page.isVisible('.terminal.xterm .xterm-screen')).toBeTruthy();

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
      const screenShotPath = path.join(__dirname, '..', 'test-results', 'diffs', 'terminalContent-actual.png');
      await elementHandle.screenshot({ path: screenShotPath });

      const expectedPath = path.join(__dirname, '..', 'fixtures', terminalReferenceFileMap[browserName] ?? terminalReferenceFileMap.default);
      const { pass } = compareImages(expectedPath, screenShotPath);
      expect(pass).toBeTruthy();

      const terminalText = await page.locator(selectors.terminalText);
      await terminalText.fill('top');
      await page.keyboard.press('Enter');
      await page.waitForTimeout(timeouts.default);

      await elementHandle.screenshot({ path: screenShotPath });
      const { pass: pass2 } = compareImages(expectedPath, screenShotPath);
      expect(pass2).not.toBeTruthy();
    }
  });
});
