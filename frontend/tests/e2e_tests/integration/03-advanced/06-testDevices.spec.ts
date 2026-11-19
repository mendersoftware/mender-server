// Copyright 2026 Northern.tech AS
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

import type { TestEnvironment } from '../../fixtures/fixtures';
import test, { expect } from '../../fixtures/fixtures';
import { selectors, timeouts } from '../../utils/constants';
import { selectReleaseByName, triggerDeploymentCreation } from '../../utils/utils';

const openDeviceDetails = (page: Page) => page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();

const skipUnlessTestDevicesAvailable = async (environment: TestEnvironment) =>
  test.skip(environment !== 'staging', 'test devices are only available on hosted Mender');

test.describe('Test devices', () => {
  let navbar;
  test.beforeEach(async ({ environment, page }) => {
    await skipUnlessTestDevicesAvailable(environment);
    navbar = page.locator('.leftFixed.leftNav');
    await navbar.getByRole('link', { name: /Devices/i }).click();
  });

  test('is marked as test device once accepted by a trial account', async ({ page }) => {
    await expect(page.getByRole('button', { name: /trial plan/i })).toBeVisible({ timeout: timeouts.tenSeconds });
    await openDeviceDetails(page);
    const testDeviceChip = page.getByRole('button', { name: 'Test device', exact: true });
    await expect(testDeviceChip).toBeVisible({ timeout: timeouts.tenSeconds });
    await testDeviceChip.click();
    await expect(page.locator('.MuiTooltip-tooltip').getByText(/[1-5]\/\d+ test devices set/)).toBeVisible({ timeout: timeouts.tenSeconds });
  });

  let deploymentId = '';
  test('can have its test device status unset and then be slow to deploy', async ({ baseUrl, page }) => {
    await openDeviceDetails(page);
    const testDeviceChip = page.getByRole('button', { name: 'Test device', exact: true });
    const deviceActions = page.getByRole('button', { name: 'device-actions' });
    const snackbar = page.getByText(/device updated successfully/i);
    await expect(testDeviceChip).toBeVisible({ timeout: timeouts.tenSeconds });

    await deviceActions.click();
    await expect(page.locator('[aria-label="remove-test-device"]')).toBeVisible();
    await expect(page.locator('[aria-label="set-test-device"]')).toHaveCount(0);
    await page.getByRole('menuitem', { name: /Remove as test device/i }).click();
    const removalDialog = page.getByRole('dialog');
    await expect(removalDialog.getByRole('heading', { name: 'Remove as test device?' })).toBeVisible();
    await removalDialog.getByRole('button', { name: 'Remove as test device', exact: true }).click();
    await expect(snackbar).toBeVisible();
    await expect(testDeviceChip).not.toBeVisible();
    await deviceActions.click();
    await page.getByRole('menuitem', { name: /Create deployment for this/i }).click();

    await selectReleaseByName(page, 'mender-demo-artifact');
    deploymentId = await triggerDeploymentCreation(
      page,
      expect(page.getByText(/Select a Release to deploy/i)).toHaveCount(0, { timeout: timeouts.tenSeconds })
    );
    await page.waitForTimeout(2 * timeouts.fifteenSeconds);
    await page.goto(`${baseUrl}ui/deployments?open=true&id=${deploymentId}`);
    const deploymentDetails = page.getByRole('dialog');
    await expect(deploymentDetails.getByText('awaiting devices')).toBeVisible();
  });

  test('can have its test device status set again & pick up the deployment', async ({ baseUrl, page }) => {
    await openDeviceDetails(page);
    const testDeviceChip = page.getByRole('button', { name: 'Test device', exact: true });
    const deviceActions = page.getByRole('button', { name: 'device-actions' });

    await deviceActions.click();
    await expect(page.locator('[aria-label="set-test-device"]')).toBeVisible();
    await expect(page.locator('[aria-label="remove-test-device"]')).toHaveCount(0);
    await page.getByRole('menuitem', { name: /Set as test device/i }).click();
    const additionDialog = page.getByRole('dialog');
    await expect(additionDialog.getByRole('heading', { name: 'Set as test device?' })).toBeVisible();
    await additionDialog.getByRole('button', { name: 'Set as test device', exact: true }).click();
    await expect(testDeviceChip).toBeVisible();
    await page.goto(`${baseUrl}ui/deployments?open=true&id=${deploymentId}`);
    const deploymentDetails = page.getByRole('dialog');
    const pendingDeviceIndicator = deploymentDetails.getByText('awaiting devices');
    await pendingDeviceIndicator.waitFor({ state: 'hidden', timeout: timeouts.fifteenSeconds });
    await expect(pendingDeviceIndicator).not.toBeVisible();
  });
});
