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
import dayjs from 'dayjs';
import isBetween from 'dayjs/plugin/isBetween.js';

import test, { expect } from '../fixtures/fixtures';
import { selectors, storagePath, timeouts } from '../utils/constants';

dayjs.extend(isBetween);

test.describe('Deployments', () => {
  test.use({ storageState: storagePath });

  test.beforeEach(async ({ baseUrl, loggedInPage: page }) => {
    await page.goto(`${baseUrl}ui/devices`);
    await page.waitForTimeout(timeouts.default);
    await page.goto(`${baseUrl}ui/releases`);
    await page.waitForTimeout(timeouts.default);
  });

  test('allows shortcut deployments', async ({ loggedInPage: page }) => {
    // create an artifact to download first
    await page.getByText(/mender-demo-artifact/i).click();
    await page.click('.MuiSpeedDial-fab');
    await page.click('[aria-label="deploy"]');
    await page.waitForSelector(selectors.deviceGroupSelect, { timeout: timeouts.fiveSeconds });
    const deviceGroupSelect = await page.getByPlaceholder(/select a device group/i);
    await deviceGroupSelect.focus();
    await deviceGroupSelect.fill('All');
    await page.click(`#deployment-device-group-selection-listbox li:has-text('All devices')`);
    const creationButton = await page.getByRole('button', { name: /create deployment/i });
    await creationButton.scrollIntoViewIfNeeded();
    await creationButton.click();
    await page.waitForSelector(selectors.deploymentListItem, { timeout: timeouts.tenSeconds });
    await page.getByRole('tab', { name: /finished/i }).click();
    await page.waitForSelector(selectors.deploymentListItemContent, { timeout: timeouts.sixtySeconds });
    const datetime = await page.getAttribute(`${selectors.deploymentListItemContent} time`, 'datetime');
    const time = dayjs(datetime);
    const earlier = dayjs().subtract(5, 'minutes');
    const now = dayjs();
    expect(time.isBetween(earlier, now));
  });

  test('allows shortcut device deployments', async ({ baseUrl, loggedInPage: page }) => {
    await page.goto(`${baseUrl}ui/devices`);
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
    await page.getByRole('tab', { name: /finished/i }).click();
    await page.waitForSelector(selectors.deploymentListItemContent, { timeout: timeouts.sixtySeconds });
    const datetime = await page.getAttribute(`${selectors.deploymentListItemContent} time`, 'datetime');
    const time = dayjs(datetime);
    const earlier = dayjs().subtract(5, 'minutes');
    const now = dayjs();
    expect(time.isBetween(earlier, now));
  });

  test('allows group deployments', async ({ loggedInPage: page }) => {
    await page.click(`a:has-text('Deployments')`);
    await page.click(`button:has-text('Create a deployment')`);

    await page.waitForSelector(selectors.releaseSelect, { timeout: timeouts.fiveSeconds });
    const releaseSelect = await page.getByPlaceholder(/select a release/i);
    await releaseSelect.focus();
    await releaseSelect.fill('mender');
    await page.click(`#deployment-release-selection-listbox li:has-text('mender-demo-artifact')`);

    await page.waitForSelector(selectors.deviceGroupSelect, { timeout: timeouts.fiveSeconds });
    const deviceGroupSelect = await page.getByPlaceholder(/select a device group/i);
    await deviceGroupSelect.focus();
    await deviceGroupSelect.fill('test');
    await page.click(`#deployment-device-group-selection-listbox li:has-text('testgroup')`);
    const creationButton = await page.getByRole('button', { name: /create deployment/i });
    await creationButton.scrollIntoViewIfNeeded();
    await creationButton.click();
    await expect(page.getByText(/Select a Release to deploy/i)).toHaveCount(0, { timeout: timeouts.tenSeconds });
    await page.waitForSelector(selectors.deploymentListItem, { timeout: timeouts.tenSeconds });
    await page.getByRole('tab', { name: /finished/i }).click();
    await page.waitForSelector(selectors.deploymentListItemContent, { timeout: timeouts.sixtySeconds });
  });
});
