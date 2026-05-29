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
import { execSync } from 'child_process';
import * as fs from 'fs';

import test, { expect } from '../../fixtures/fixtures';
import { isEnterpriseOrStaging } from '../../utils/commands';

const manifestArtifactApiUrl =
  'https://api.github.com/repos/mendersoftware/mender-server-enterprise/contents/backend/services/deployments/tests/data/test.manifest.mender?ref=main';
const manifestFileLocation = 'fixtures/test.manifest.mender';
const manifestYamlFileLocation = 'fixtures/manifest.yaml';

test.describe('Manifests', () => {
  test.beforeAll(async ({ baseUrl, browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    await page.goto(baseUrl);
    const features = await page.evaluate(() => (window as any).mender_environment?.features);
    if (!features?.hasManifestsEnabled) {
      console.log(`Manifest feature disabled - no artifact download...`);
      return;
    }

    console.log(`Downloading test manifest artifact via Github API...`);
    // download the test manifest artifact from the enterprise repo via GitHub API
    const { GITHUB_TOKEN: githubToken } = process.env;
    const response = await fetch(manifestArtifactApiUrl, { headers: { Accept: 'application/vnd.github.raw+json', Authorization: `token ${githubToken}` } });
    if (response.ok) {
      const buffer = await response.arrayBuffer();
      fs.writeFileSync(manifestFileLocation, Buffer.from(buffer));
      // Dump the yaml file (manifest.yaml) from the manifest mender file
      try {
        execSync(`mender-artifact dump --files fixtures/ ${manifestFileLocation}`, { stdio: ['inherit', 'pipe', 'pipe'] });
      } catch (e) {
        if (fs.existsSync(manifestYamlFileLocation)) {
          console.warn('mender-artifact dump exited non-zero, other worker created the file, continuing.');
        } else {
          console.error(e instanceof Error ? e.message : String(e));
          throw e;
        }
      }
    } else {
      console.warn(`Failed to download manifest artifact (${response.status}) - upload tests will be skipped`);
    }
  });

  test.beforeEach(async ({ environment, page }) => {
    const features = await page.evaluate(() => (window as any).mender_environment?.features);
    test.skip(!features || !features.hasManifestsEnabled, 'Manifests feature flag is not enabled');
    test.skip(!isEnterpriseOrStaging(environment), 'Manifests are only available in enterprise or staging environments');

    await page.locator('.leftFixed.leftNav').getByRole('link', { name: 'Software', exact: true }).click();
    await page.getByRole('tab', { name: /manifests/i }).click();
  });
  test('allows .mender manifest upload with tags and description', async ({ page }) => {
    await page.getByRole('button', { name: /upload a manifest/i }).click();
    const drawer = page.locator('.MuiDrawer-paper');
    await drawer.locator('.dropzone input').setInputFiles(manifestFileLocation);
    await drawer.getByPlaceholder(/add notes here/i).fill('uploaded via e2e');
    const tagsInput = drawer.getByPlaceholder(/add tags/i);
    await tagsInput.fill('e2e-tag');
    await tagsInput.press('Enter');
    await drawer.getByRole('button', { name: /^upload$/i }).click();
    await expect(page.getByRole('cell', { name: 'e2e-tag' })).toBeVisible();
  });

  test('shows the manifests tab and list', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: /name/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /latest modified/i })).toBeVisible();
  });

  test('shows manifest details when clicking a row', async ({ page }) => {
    await page.getByRole('cell', { name: 'test' }).click();
    await expect(page.getByText(`Manifest information for`)).toBeVisible();

    await expect(page.getByText(/^Signature/i)).toBeVisible();
    await expect(page.getByText(/Component types/i)).toBeVisible();

    await page.getByLabel(/close/i).click();
    await expect(page.getByText(`Manifest information for`)).not.toBeVisible();
  });

  test('allows editing manifest notes', async ({ page }) => {
    await page.getByRole('cell', { name: 'test' }).click();
    await expect(page.getByText('Manifest information for')).toBeVisible();
    const drawer = page.locator('.MuiDrawer-paper');
    const editButton = drawer.getByRole('heading', { name: 'Notes' }).locator('..').getByRole('button', { name: /edit/i });
    await editButton.click();
    const textField = drawer.getByRole('textbox');
    await textField.fill('e2e updated notes');
    await drawer.getByLabel(/confirm/i).click();
    await expect(page.getByText(/Manifest details were updated successfully/i)).toBeVisible();
    await expect(drawer.getByLabel(/confirm/i)).not.toBeVisible();
    await expect(drawer.getByText('e2e updated notes')).toBeVisible();
    await page.getByLabel(/close/i).click();
  });

  test('allows editing manifest tags', async ({ page }) => {
    await page.getByRole('cell', { name: 'test' }).click();
    await expect(page.getByText('Manifest information for')).toBeVisible();
    const drawer = page.locator('.MuiDrawer-paper');
    const editButton = drawer.getByRole('heading', { name: 'Tags' }).locator('..').getByRole('button', { name: /edit/i });
    await editButton.click();
    const tagsInput = drawer.locator('#tags-chip-select');
    await tagsInput.fill('e2e-new-tag');
    await tagsInput.press('Enter');
    await drawer.getByLabel(/confirm/i).click();
    await expect(page.getByText(/Manifest details were updated successfully/i)).toBeVisible();
    await expect(editButton).toBeVisible();
    await expect(drawer.getByText('e2e-new-tag')).toBeVisible();
    await page.getByLabel(/close/i).click();
  });

  test('allows removing an uploaded manifest', async ({ page }) => {
    const targetCell = page.getByRole('cell', { name: 'e2e-tag' });
    await expect(targetCell).toBeVisible();
    const targetRow = page.getByRole('row').filter({ has: targetCell });
    await targetRow.getByRole('checkbox').click();
    await page.click('.MuiSpeedDial-fab');
    await page.getByRole('menuitem', { name: 'Delete Manifest' }).click();
    await expect(page.getByText(/are you sure you want to remove/i)).toBeVisible();
    await page.getByRole('button', { name: 'Remove', exact: true }).click();
    await expect(page.getByRole('cell', { name: 'e2e-tag' })).not.toBeVisible();
  });
  test('allows .yaml manifest upload with tags and description', async ({ page }) => {
    await page.getByRole('button', { name: /upload a manifest/i }).click();
    const drawer = page.locator('.MuiDrawer-paper');
    await drawer.locator('.dropzone input').setInputFiles(manifestYamlFileLocation);
    await drawer.getByPlaceholder(/add notes here/i).fill('yaml e2e upload');
    const tagsInput = drawer.getByPlaceholder(/add tags/i);
    await tagsInput.fill('yaml-e2e-tag');
    await tagsInput.press('Enter');
    await drawer.getByRole('button', { name: /^upload$/i }).click();
    await expect(page.getByRole('cell', { name: 'yaml-e2e-tag' })).toBeVisible();
  });
});
