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
import * as path from 'path';

import test, { expect } from '../../fixtures/fixtures';
import { extractArtifactFromDevice, modifyArtifactChecksum } from '../../utils/commands';
import { selectors, timeouts } from '../../utils/constants';
import { selectReleaseByName, triggerDeploymentCreation } from '../../utils/utils.ts';

const qemuDeviceType = 'qemux86-64';

test.describe('Devices', () => {
  let navbar;

  test.beforeEach(async ({ page }) => {
    navbar = page.locator('.leftFixed.leftNav');
  });

  test('can authorize a device', async ({ browserName, page }) => {
    // allow twice the device interaction time + roughly a regular test execution time
    test.setTimeout(4 * timeouts.sixtySeconds + timeouts.fifteenSeconds);
    await navbar.getByRole('link', { name: /Devices/i }).click({ force: browserName === 'webkit' });
    let hasAcceptedDevice = false;
    try {
      await page.waitForSelector(`css=${selectors.deviceListItem}`, { timeout: timeouts.default });
      hasAcceptedDevice = await page.isVisible(selectors.deviceListItem);
    } catch {
      console.log(`no accepted device present so far`);
    }
    if (!hasAcceptedDevice) {
      const pendingMessage = await page.getByText(/pending authorization/i);
      await pendingMessage.waitFor({ timeout: 3 * timeouts.sixtySeconds });
      await pendingMessage.click();
      await page.click(selectors.deviceListCheckbox);
      await page.click('.MuiSpeedDial-fab');
      await page.click('[aria-label="accept"]');
    }
    await page.locator(`input:near(:text("Status:"))`).first().click({ force: true });
    await page.click(`css=.MuiPaper-root >> text=/Accepted/i`);
    await page.waitForSelector(`css=${selectors.deviceListItem} >> text=/${qemuDeviceType}/`, { timeout: 2 * timeouts.sixtySeconds });
    const element = await page.textContent(selectors.deviceListItem);
    expect(element.includes(qemuDeviceType)).toBeTruthy();
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await page.getByText(/Device information for/i).waitFor();
    await expect(page.getByText('Authentication status')).toBeVisible();
  });

  test('extract delta', async ({ page }) => {
    test.setTimeout(12 * timeouts.sixtySeconds);
    const extractedArtifactPath = path.resolve('fixtures/extracted-snapshot.mender');
    const modifiedArtifactPath = path.resolve('fixtures/modified-snapshot.mender');
    const { CLIENT_IP: targetHost = 'qemu-client' } = process.env;

    try {
      await extractArtifactFromDevice(targetHost, extractedArtifactPath);
      await modifyArtifactChecksum(extractedArtifactPath, modifiedArtifactPath);
    } catch (error) {
      console.error('Failed to extract and modify artifact:', error);
      throw error;
    }
    expect(fs.existsSync(extractedArtifactPath)).toBeTruthy();
    expect(fs.existsSync(modifiedArtifactPath)).toBeTruthy();

    await navbar.getByRole('link', { name: 'Releases', exact: true }).click();
    const uploadButton = await page.getByRole('button', { name: /upload/i });
    await uploadButton.click();
    const drawer = page.locator(`.MuiDialog-paper`);

    await drawer.locator('.dropzone input').setInputFiles(extractedArtifactPath);
    await drawer.getByRole('button', { name: /Upload/i }).click();
    await page.waitForTimeout(timeouts.sixtySeconds);
    await page.getByText(/last modified/i).waitFor();

    await page.getByRole('button', { name: /upload/i }).click();
    const drawer2 = page.locator(`.MuiDialog-paper`);
    await drawer2.locator('.dropzone input').setInputFiles(modifiedArtifactPath);
    await drawer2.getByRole('button', { name: /Upload/i }).click();
    await page.waitForTimeout(timeouts.sixtySeconds);
    await page.getByText(/last modified/i).waitFor();

    if (fs.existsSync(extractedArtifactPath)) {
      fs.unlinkSync(extractedArtifactPath);
    }
    if (fs.existsSync(modifiedArtifactPath)) {
      fs.unlinkSync(modifiedArtifactPath);
    }
  });

  test('allows shortcut device deployments', async ({ page }) => {
    test.setTimeout(12 * timeouts.sixtySeconds);
    await navbar.getByRole('link', { name: /devices/i }).click();
    await page.getByText(/qemu/i).click();
    await page.click('.MuiSpeedDial-fab');
    await page.click('[aria-label="create-deployment"]');

    await selectReleaseByName(page, 'snapshot-test');
    await triggerDeploymentCreation(page, expect(page.getByText(/Select a Release to deploy/i)).toHaveCount(0, { timeout: timeouts.tenSeconds }));
    await page.getByText('finished').click();
    await page
      .getByRole('listitem')
      .first()
      .waitFor({ timeout: 10 * timeouts.sixtySeconds });
  });

  test('allows shortcut device deployments 2', async ({ page }) => {
    await navbar.getByRole('link', { name: /devices/i }).click();
    await page.getByText(/qemu/i).click();
    await page.click('.MuiSpeedDial-fab');
    await page.click('[aria-label="create-deployment"]');

    await selectReleaseByName(page, 'snapshot-modified');

    await page.getByRole('button', { name: /advanced options/i }).click();
    await page.getByRole('checkbox', { name: /delta artifacts/i }).click();
    await triggerDeploymentCreation(page, expect(page.getByText(/Select a Release to deploy/i)).toHaveCount(0, { timeout: timeouts.tenSeconds }));
    await page.getByRole('listitem').first().waitFor({ timeout: timeouts.sixtySeconds });
  });

  test('shows delta generation', async ({ page }) => {
    await navbar.getByRole('link', { name: /deployments/i }).click();
    const pageContent = page.locator('.rightFluid.container');
    const listItem = pageContent.getByRole('listitem').first();
    await listItem.waitFor({ timeout: timeouts.sixtySeconds });
    await listItem.getByRole('button', { name: /view details/i }).click();
    await page.waitForTimeout(timeouts.default);
    await page.getByRole('button', { name: /close/i }).click();
    await page.getByText('finished').click();
    await page.getByRole('listitem').first().waitFor({ timeout: timeouts.sixtySeconds });
    await navbar.getByRole('link', { name: /Releases/i }).click();
    await page.getByRole('tab', { name: /delta/i }).click();
    await page.getByText(/to version/i).waitFor({ timeout: timeouts.sixtySeconds });
  });

  test('opens & closes delta generation details', async ({ page }) => {
    await navbar.getByRole('link', { name: /Releases/i }).click();
    await page.getByRole('tab', { name: /delta/i }).click();
    await page.getByText(qemuDeviceType).click();
    await expect(page.getByText('Delta Artifact information')).toBeVisible();
    await page.waitForTimeout(timeouts.default);
    await page
      .getByRole('table')
      .getByText(/pending/i)
      .waitFor({ state: 'hidden', timeout: 2 * timeouts.sixtySeconds });
    const detailsColumns = await page.getByRole('cell', { name: '-', exact: true }).all();
    expect(detailsColumns.length).toBeLessThanOrEqual(1); // allow a single column to be empty in case there was no delta generated
    await page.getByRole('button', { name: /close/i }).click();
    await expect(page.getByText('Delta Artifact information')).not.toBeVisible();
  });
});
