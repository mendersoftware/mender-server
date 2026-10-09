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
import type { APIRequestContext, Page } from '@playwright/test';
import dayjs from 'dayjs';
import isBetween from 'dayjs/plugin/isBetween.js';
import * as fs from 'fs';

import test, { expect } from '../../fixtures/fixtures';
import { getTokenFromStorage, isEnterpriseOrStaging, startClient, tenantTokenRetrieval } from '../../utils/commands';
import { releaseTag, selectors, timeouts } from '../../utils/constants';
import { locateReleaseByName, navigateTo, selectReleaseByName, triggerDeploymentCreation } from '../../utils/utils.ts';

dayjs.extend(isBetween);

const stressTestClientPath = './mender-stress-test-client';
// a dedicated mac prefix keeps the simulated devices apart from the demo device & other stress test clients
const uniformClientMacPrefix = 'fe';
const uniformClientCount = 6;
const uniformGroup = 'uniformgroup';

type AuthSet = { id: string; status: string };
type DevauthDevice = { auth_sets: AuthSet[]; id: string; identity_data: { mac?: string } };

const getUniformClientDevices = async (baseUrl: string, request: APIRequestContext, headers: Record<string, string>): Promise<DevauthDevice[]> => {
  const response = await request.get(`${baseUrl}api/management/v2/devauth/devices?per_page=500`, { headers });
  const devices: DevauthDevice[] = await response.json();
  return devices.filter(({ identity_data }) => identity_data.mac?.startsWith(`${uniformClientMacPrefix}:`));
};

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
  test('check time filters before deployment', async ({ page }) => {
    await navigateTo(page, 'deployments');
    await page.getByRole('tab', { name: /finished/i }).click();
    await checkTimeFilter(page, 'From');
    await checkTimeFilter(page, 'To', true);
  });
  test('advanced deployment release filters', async ({ page }) => {
    const releaseName = 'mender-demo-artifact';
    await navigateTo(page, 'deployments');
    await page.click(`button:has-text('Create a deployment')`);
    await page.getByRole('button', { name: 'Select software' }).click();

    await page.getByRole('button', { name: 'Advanced filter' }).click();
    await page.getByRole('combobox', { name: 'Select tags...' }).click();
    await page.getByRole('option', { name: releaseTag }).click();
    await page.waitForTimeout(timeouts.default);
    await expect(locateReleaseByName(page, releaseName)).toBeVisible();

    await page.getByRole('button', { name: 'Clear all' }).click();
    await page.getByRole('combobox', { name: 'Select Artifact type' }).click();
    await page.getByRole('option', { name: 'directory' }).click();
    await page.waitForTimeout(timeouts.default);
    await expect(locateReleaseByName(page, releaseName)).toBeVisible();
    await page.getByRole('button', { name: 'Clear all' }).click();
  });

  test('ensure release page filters are not used on deployment creation', async ({ page }) => {
    await navigateTo(page, 'software');
    await page.getByPlaceholder(/select tags/i).click();
    await page.getByRole('option', { name: releaseTag.toLowerCase() }).click();
    await page.keyboard.press('Escape');
    await page.waitForTimeout(timeouts.default);
    await navigateTo(page, 'deployments');
    await page
      .getByRole('button', { name: /create a deployment/i })
      .first()
      .click();
    await page.getByRole('button', { name: 'Select software' }).click();
    await expect(locateReleaseByName(page, 'mender-demo-artifact')).toBeVisible();
  });
  test('allows shortcut deployments', async ({ page }) => {
    await navigateTo(page, 'software');
    test.setTimeout(6 * timeouts.sixtySeconds);
    // create an artifact to download first
    await page.getByText(/mender-demo-artifact/i).click();
    await page.getByRole('button', { name: 'release-actions' }).click();
    await page.getByRole('menuitem', { name: /Create a deployment for this/i }).click();
    await page.waitForSelector(selectors.deviceGroupSelect, { timeout: timeouts.fiveSeconds });
    const deviceGroupSelect = await page.getByPlaceholder(/select a device group/i);
    await deviceGroupSelect.focus();
    await deviceGroupSelect.fill('All');
    await page.click(`#deployment-device-group-selection-listbox li:has-text('All devices')`);
    await triggerDeploymentCreation(page, page.getByRole('listitem').first().waitFor({ timeout: timeouts.tenSeconds }));
    await page.getByRole('tab', { name: /finished/i }).click();
    const pageContent = page.locator('.rightFluid.container');
    const listItem = pageContent.getByRole('listitem').first();
    await listItem.waitFor({ timeout: 5 * timeouts.sixtySeconds });
    const datetime = await listItem.locator('time').last().getAttribute('datetime');
    const time = dayjs(datetime);
    const earlier = dayjs().subtract(5, 'minutes');
    const now = dayjs();
    expect(time.isBetween(earlier, now));
  });

  test('allows shortcut device deployments', async ({ page }) => {
    test.setTimeout(6 * timeouts.sixtySeconds);
    await navigateTo(page, 'devices');
    await page.getByText(/original/i).click();
    await expect(page.getByText(/device information for/i)).toBeVisible();
    await page.getByRole('button', { name: 'device-actions' }).click();
    await page.getByRole('menuitem', { name: /Create deployment for this/i }).click();

    await selectReleaseByName(page, 'mender-demo-artifact');
    await triggerDeploymentCreation(page, expect(page.getByText(/Select software to deploy/i)).toHaveCount(0, { timeout: timeouts.tenSeconds }));
    await page.getByRole('tab', { name: /finished/i }).click();
    const pageContent = page.locator('.rightFluid.container');
    const listItem = pageContent.getByRole('listitem').first();
    await listItem.waitFor({ timeout: 5 * timeouts.sixtySeconds });
    const datetime = await listItem.locator('time').last().getAttribute('datetime');
    const time = dayjs(datetime);
    const earlier = dayjs().subtract(5, 'minutes');
    const now = dayjs();
    expect(time.isBetween(earlier, now));
    await checkTimeFilter(page, 'From', true);
    await checkTimeFilter(page, 'To', true);
  });

  test('allows group deployments', async ({ page, staticGroupName }) => {
    test.setTimeout(6 * timeouts.sixtySeconds);
    await navigateTo(page, 'deployments');
    await page.click(`button:has-text('Create a deployment')`);

    await selectReleaseByName(page, 'mender-demo-artifact');

    await page.waitForSelector(selectors.deviceGroupSelect, { timeout: timeouts.fiveSeconds });
    const deviceGroupSelect = await page.getByPlaceholder(/select a device group/i);
    await deviceGroupSelect.focus();
    await deviceGroupSelect.fill('test');
    await page.click(`#deployment-device-group-selection-listbox li:has-text('${staticGroupName}')`);
    await triggerDeploymentCreation(page, expect(page.getByText(/Select software to deploy/i)).toHaveCount(0, { timeout: timeouts.tenSeconds }));
    await page.getByRole('tab', { name: /finished/i }).click();
    const pageContent = page.locator('.rightFluid.container');
    const listItem = pageContent.getByRole('listitem').first();
    await listItem.waitFor({ timeout: 5 * timeouts.sixtySeconds });
  });

  test('allows deployment filtering by name', async ({ demoDeviceName, page }) => {
    await navigateTo(page, 'deployments');
    await page.getByRole('tab', { name: /Finished/i }).click();
    await page.getByPlaceholder(/group or device/i).click();
    await page.getByPlaceholder(/group or device/i).fill(demoDeviceName);
    await page.getByRole('listitem').first().waitFor({ timeout: timeouts.fiveSeconds });
    const deployments = await page.getByRole('listitem').all();
    expect(deployments.length).toBeTruthy();
  });

  test('deployment pagination', async ({ baseUrl, page, request }) => {
    const token = await getTokenFromStorage(baseUrl);
    const pendingDeploymentRequests = Array.from({ length: 60 }, (_, index) => ({
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
      try {
        await page.getByRole('button', { name: 'next' }).click({ noWaitAfter: true, force: true });
      } catch {
        // the pagination component may re-render between clicks, momentarily hiding the button - just retry
        continue;
      }
    }
    await expect(page.getByText(/-([5|6]\d) of \1/)).toBeVisible(); // depending on the deployment speed of other tests there might be slightly more than 50
    await expect(page.getByText(/queued to start/i).first()).toBeVisible();
  });

  test('allows uniform phased deployments', async ({ baseUrl, environment, page, request }) => {
    test.skip(!isEnterpriseOrStaging(environment), 'phased deployments are not available in OS');
    test.skip(!fs.existsSync(stressTestClientPath), 'requires the mender-stress-test-client to simulate a device fleet');
    test.setTimeout(6 * timeouts.sixtySeconds);
    const headers = { Authorization: `Bearer ${getTokenFromStorage(baseUrl)}` };
    const tenantToken = await tenantTokenRetrieval(baseUrl, page);
    const client = await startClient(baseUrl, tenantToken, uniformClientCount, [`--mac-address-prefix=${uniformClientMacPrefix}`]);
    try {
      let devices: DevauthDevice[] = [];
      await expect(async () => {
        devices = await getUniformClientDevices(baseUrl, request, headers);
        expect(devices).toHaveLength(uniformClientCount);
      }).toPass({ timeout: 2 * timeouts.sixtySeconds });
      await Promise.all(
        devices.flatMap(({ auth_sets, id }) =>
          auth_sets
            .filter(({ status }) => status === 'pending')
            .map(authSet =>
              request.put(`${baseUrl}api/management/v2/devauth/devices/${id}/auth/${authSet.id}/status`, { data: { status: 'accepted' }, headers })
            )
        )
      );
      const deviceIds = devices.map(({ id }) => id);
      // accepted devices only show up in the inventory with a slight delay, so the grouping may have to be repeated
      await expect(async () => {
        await request.patch(`${baseUrl}api/management/v1/inventory/groups/${uniformGroup}/devices`, { data: deviceIds, headers });
        const response = await request.get(`${baseUrl}api/management/v1/inventory/groups/${uniformGroup}/devices?per_page=500`, { headers });
        expect(await response.json()).toHaveLength(uniformClientCount);
      }).toPass({ timeout: timeouts.sixtySeconds });

      await navigateTo(page, 'deployments');
      await page
        .getByRole('button', { name: /create a deployment/i })
        .first()
        .click();
      await selectReleaseByName(page, 'mender-demo-artifact');
      const deviceGroupSelect = page.getByPlaceholder(/select a device group/i);
      await deviceGroupSelect.focus();
      await deviceGroupSelect.fill(uniformGroup);
      await page.click(`#deployment-device-group-selection-listbox li:has-text('${uniformGroup}')`);
      await page.getByText(/show advanced options/i).click();
      await page.getByRole('checkbox', { name: /select a rollout pattern/i }).check();
      await page
        .getByRole('combobox')
        .filter({ hasText: /^custom$/i })
        .click();
      await page.getByText(/uniform/i).click();
      await page.getByRole('radio', { name: /by number of devices/i }).check();
      const batchSizeInput = page
        .locator('table')
        .filter({ hasText: /first phase begins/i })
        .getByRole('textbox')
        .first();
      await batchSizeInput.clear();
      await batchSizeInput.fill('2');
      await batchSizeInput.press('Tab');
      const deploymentId = await triggerDeploymentCreation(
        page,
        expect(page.getByText(/Select software to deploy/i)).toHaveCount(0, { timeout: timeouts.tenSeconds })
      );
      await page.goto(`${baseUrl}ui/deployments?open=true&id=${deploymentId}`);
      await expect(page.getByText('Uniform, 2 devices per phase')).toBeVisible({ timeout: timeouts.tenSeconds });
      await expect(page.getByText('Phase 1', { exact: true })).toBeVisible();
      await expect(page.getByText('Phase 3', { exact: true })).toBeVisible();
      await expect(page.getByText('(Final phase)')).toHaveCount(1);
    } finally {
      client.kill();
    }
  });
});
