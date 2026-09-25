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
import * as fs from 'fs';

import test, { expect } from '../../fixtures/fixtures';
import { timeouts } from '../../utils/constants';
import { acceptPendingDevice, navigateTo, selectReleaseByName, triggerDeploymentCreation } from '../../utils/utils.ts';

const rtosArtifactUrl = 'https://raw.githubusercontent.com/mendersoftware/mender-orchestrator-support/main/demo/premade-artifacts/rtos-v2.mender';
const rtosArtifactLocation = 'fixtures/rtos-v2.mender';
const manifestLocation = 'fixtures/manifest-rtos.yaml';

const targetComponentVersion = 'rtos-v2';
const manifestName = 'system-core-rtos-v2';

test.describe('Orchestrator device', () => {
  test.beforeAll(async () => {
    if (fs.existsSync(rtosArtifactLocation)) {
      return;
    }
    const response = await fetch(rtosArtifactUrl);
    if (!response.ok) {
      throw new Error(`Failed to download the rtos artifact (${response.status}) - cannot deploy a manifest without it`);
    }
    fs.writeFileSync(rtosArtifactLocation, Buffer.from(await response.arrayBuffer()));
  });

  test.beforeEach(async ({ page }) => {
    const features = await page.evaluate(() => (window as any).mender_environment?.features);
    test.skip(!features || !features.hasManifestsEnabled, 'Manifests feature flag is not enabled');
  });

  test('lists the components reported by the orchestrator', async ({ page }) => {
    test.setTimeout(8 * timeouts.sixtySeconds);
    await acceptPendingDevice(page);
    await page.getByText(/qemu/i).click();

    const systemTab = page.getByRole('tab', { name: /system/i });
    await systemTab.waitFor({ timeout: 4 * timeouts.sixtySeconds });
    await systemTab.click();

    await expect(page.getByText(/System information/i)).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /component type/i })).toBeVisible();
    expect(await page.getByRole('table').getByRole('row').count()).toBeGreaterThan(1);
  });

  test('uploads the artifact and the manifest referring to it', async ({ page }) => {
    test.setTimeout(6 * timeouts.sixtySeconds);
    await navigateTo(page, 'software');
    await page.getByRole('button', { name: 'Upload an artifact' }).click();
    const artifactDialog = page.locator('.MuiDialog-paper');
    await artifactDialog.locator('.dropzone input').setInputFiles(rtosArtifactLocation);
    await artifactDialog.getByRole('button', { name: 'Upload artifact' }).click();
    await expect(page.getByText(targetComponentVersion)).toBeVisible({ timeout: 2 * timeouts.sixtySeconds });

    await page.getByRole('tab', { name: /manifests/i }).click();
    await page.getByRole('button', { name: /upload a manifest/i }).click();
    const manifestDrawer = page.locator('.MuiDrawer-paper');
    await manifestDrawer.locator('.dropzone input').setInputFiles(manifestLocation);
    await manifestDrawer.getByRole('button', { name: /^upload$/i }).click();
    await expect(page.getByRole('cell', { name: manifestName })).toBeVisible({ timeout: timeouts.sixtySeconds });
  });

  test('deploys the manifest and sees the component versions change', async ({ page }) => {
    test.setTimeout(20 * timeouts.sixtySeconds);
    await navigateTo(page, 'devices');
    await page.getByText(/qemu/i).click();
    await page.getByRole('button', { name: 'device-actions' }).click();
    await page.getByRole('menuitem', { name: /Create deployment for this/i }).click();

    await selectReleaseByName(page, manifestName);
    await triggerDeploymentCreation(page, expect(page.getByText(/Select software to deploy/i)).toHaveCount(0, { timeout: timeouts.tenSeconds }));
    await page.getByText('finished').click();
    await page
      .getByRole('listitem')
      .first()
      .waitFor({ timeout: 10 * timeouts.sixtySeconds });

    await expect(async () => {
      await navigateTo(page, 'devices');
      await page.getByRole('tab', { name: /system/i }).click();
      await expect(page.getByRole('cell', { name: targetComponentVersion }).first()).toBeVisible();
    }).toPass({ timeout: 6 * timeouts.sixtySeconds });
  });
});
