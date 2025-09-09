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
import test, { expect } from '../../fixtures/fixtures';
import { expectedArtifactName, selectors, timeouts } from '../../utils/constants';

const fileName = `${expectedArtifactName}.mender`;
const fileLocation = `fixtures/${fileName}`;

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
    await page.waitForSelector(`css=${selectors.deviceListItem} >> text=/original/`, { timeout: 2 * timeouts.sixtySeconds });
    const element = await page.textContent(selectors.deviceListItem);
    expect(element.includes('original')).toBeTruthy();
    await page.locator(`css=${selectors.deviceListItem} div:last-child`).last().click();
    await page.getByText(/Device information for/i).waitFor();
    await expect(page.getByText('Authentication status')).toBeVisible();
  });

  test('allows file uploads', async ({ page }) => {
    await navbar.getByRole('link', { name: /Releases/i }).click();
    const uploadButton = await page.getByRole('button', { name: /upload/i });
    await uploadButton.click();
    const drawer = page.locator(`.MuiDialog-paper`);
    await drawer.locator('.dropzone input').setInputFiles(fileLocation);
    await drawer.getByRole('button', { name: /Upload/i }).click();
    await page.getByText(/last modified/i).waitFor();
  });

  test('allows shortcut device deployments', async ({ page }) => {
    await navbar.getByRole('link', { name: /devices/i }).click();
    // create an artifact to download first
    await page.getByText(/original/i).click();
    await page.click('.MuiSpeedDial-fab');
    await page.click('[aria-label="create-deployment"]');
    await page.waitForSelector(selectors.releaseSelect, { timeout: timeouts.fiveSeconds });
    const releaseSelect = await page.getByPlaceholder(/select a release/i);
    await releaseSelect.focus();
    await releaseSelect.fill('mender-demo');
    await page.click(`#deployment-release-selection-listbox li`);
    await page.getByRole('button', { name: 'Clear' }).click();
    const textContent = await releaseSelect.textContent();
    expect(textContent).toBeFalsy();
    await releaseSelect.focus();
    await releaseSelect.fill('mender-demo');
    await page.click(`#deployment-release-selection-listbox li`);
    const creationButton = await page.getByRole('button', { name: /create deployment/i });
    await creationButton.scrollIntoViewIfNeeded();
    await creationButton.click();
    await expect(page.getByText(/Select a Release to deploy/i)).toHaveCount(0, { timeout: timeouts.tenSeconds });
    await page.waitForSelector(selectors.deploymentListItemContent, { timeout: timeouts.sixtySeconds });
  });

  test('shows delta generation', async ({ page }) => {
    await navbar.getByRole('link', { name: /Releases/i }).click();
    await page.getByRole('tab', { name: /delta/i }).click();
    await page.getByText(/to version/i).waitFor({ timeout: timeouts.sixtySeconds });
  });
});
