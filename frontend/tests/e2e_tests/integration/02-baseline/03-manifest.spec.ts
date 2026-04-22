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

import test, { expect } from '../../fixtures/fixtures';
import { getTokenFromStorage, isEnterpriseOrStaging } from '../../utils/commands';

const manifestArtifactApiUrl =
  'https://api.github.com/repos/mendersoftware/mender-server-enterprise/contents/backend/services/deployments/tests/data/test.manifest.mender?ref=main';
const manifestFileLocation = 'fixtures/test.manifest.mender';

test.describe('Manifests', () => {
  test.beforeAll(async ({ baseUrl, browser, request }) => {
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
    } else {
      console.warn(`Failed to download manifest artifact (${response.status}) - upload tests will be skipped`);
    }

    const token = getTokenFromStorage(baseUrl);
    const fileBuffer = fs.readFileSync(manifestFileLocation);
    const upload = await request.post(`${baseUrl}api/management/v1alpha1/deployments/manifests`, {
      headers: { Authorization: `Bearer ${token}` },
      multipart: {
        artifact: { name: 'test.manifest.mender', mimeType: 'application/octet-stream', buffer: fileBuffer }
      }
    });
    if (!upload.ok()) {
      console.warn(`Failed to upload manifest artifact via API (${upload.status()})`);
    }
  });

  test.beforeEach(async ({ environment, page }) => {
    const features = await page.evaluate(() => (window as any).mender_environment?.features);
    test.skip(!features || !features.hasManifestsEnabled, 'Manifests feature flag is not enabled');
    test.skip(!isEnterpriseOrStaging(environment), 'Manifests are only available in enterprise or staging environments');

    await page.locator('.leftFixed.leftNav').getByRole('link', { name: 'Software', exact: true }).click();
    await page.getByRole('tab', { name: /manifests/i }).click();
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
});
