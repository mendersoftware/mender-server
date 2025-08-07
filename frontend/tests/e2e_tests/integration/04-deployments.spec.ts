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
import type { Page } from '@playwright/test';
import dayjs from 'dayjs';
import isBetween from 'dayjs/plugin/isBetween.js';

import test, { expect } from '../fixtures/fixtures';
import { getTokenFromStorage } from '../utils/commands';
import { releaseTag, selectors, timeouts } from '../utils/constants';

dayjs.extend(isBetween);

const checkTimeFilter = async (page: Page, name: string, isSetToday?: boolean) => {
  const input = page.getByRole('group', { name });
  if (isSetToday) {
    await page.waitForTimeout(timeouts.oneSecond); // wait a little as sometimes the rendering hasn't fully finished when the following runs
    const shownDate = await input.textContent(); // will be shown as `YYYY-MM-DD${name}`
    await expect(shownDate).toContain(dayjs().format('YYYY-MM-DD'));
  }
  await expect(input).not.toHaveClass(/Mui-error/);
};

test.describe('Deployments', () => {
  test.beforeEach(async ({ baseUrl, page }) => {
    await page.goto(`${baseUrl}ui/devices`);
    await page.waitForTimeout(timeouts.default);
    await page.goto(`${baseUrl}ui/releases`);
    await page.waitForTimeout(timeouts.default);
  });
  test('check time filters before deployment', async ({ baseUrl, page }) => {
    await page.goto(`${baseUrl}ui/deployments`);
    await page.getByRole('tab', { name: /finished/i }).click();
    await checkTimeFilter(page, 'From');
    await checkTimeFilter(page, 'To', true);
  });
  test('ensure release page filters are not used on deployment creation', async ({ baseUrl, page }) => {
    await page.getByPlaceholder(/select tags/i).fill(`${releaseTag.toLowerCase()},`);
    await page.goto(`${baseUrl}ui/deployments`);
    await page.getByRole('button', { name: /create a deployment/i }).click();
    await page.waitForSelector(selectors.releaseSelect, { timeout: timeouts.fiveSeconds });
    const releaseSelect = await page.getByPlaceholder(/select a release/i);
    await releaseSelect.focus();
    await expect(page.locator(`#deployment-release-selection-listbox li:has-text('mender-demo-artifact')`)).toBeVisible();
  });
  test('allows shortcut deployments', async ({ page }) => {
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

  test('allows shortcut device deployments', async ({ baseUrl, page }) => {
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
    await checkTimeFilter(page, 'From', true);
    await checkTimeFilter(page, 'To', true);
    await page.waitForSelector(selectors.deploymentListItemContent, { timeout: timeouts.sixtySeconds });
    const datetime = await page.getAttribute(`${selectors.deploymentListItemContent} time`, 'datetime');
    const time = dayjs(datetime);
    const earlier = dayjs().subtract(5, 'minutes');
    const now = dayjs();
    expect(time.isBetween(earlier, now));
  });

  test('allows group deployments', async ({ page }) => {
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
  test('deployment pagination', async ({ baseUrl, page, request }) => {
    const token = await getTokenFromStorage(baseUrl);
    const pendingDeploymentRequests = Array.from({ length: 50 }, (_, index) => ({
      artifact_name: 'terminalImage',
      all_devices: true,
      max_devices: index,
      name: `deployment-${index + 1}`
    })).map(deployment =>
      request.post(`${baseUrl}api/management/v1/deployments/deployments`, { data: deployment, headers: { Authorization: `Bearer ${token}` } })
    );
    await Promise.all(pendingDeploymentRequests);
    await page.goto(`${baseUrl}ui/deployments`);
    await expect(page.getByText(/rows/i)).toBeVisible();
    await page.getByText(/rows/i).scrollIntoViewIfNeeded();
    // 10 clicks as anything leading outside of the 50 + something releases present (considering the 10 item page size)
    for (let clickAttempt = 0; clickAttempt < 10; clickAttempt++) {
      const paginationButton = page.getByRole('button', { name: 'next' });
      await paginationButton.click({ noWaitAfter: true, force: true });
    }
    await expect(page.getByText(/queued to start/i).first()).toBeVisible();
  });
});
