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
import { execSync } from 'child_process';
import dayjs from 'dayjs';
import isBetween from 'dayjs/plugin/isBetween.js';
import * as fs from 'fs';
import { parse } from 'yaml';

import test, { expect } from '../../fixtures/fixtures';
import { getTokenFromStorage, isEnterpriseOrStaging, tagRelease } from '../../utils/commands';
import { expectedArtifactName, releaseTag, selectors, timeouts } from '../../utils/constants';

dayjs.extend(isBetween);

const fileName = `${expectedArtifactName}.mender`;
const demoArtifactLocation = `https://dgsbl4vditpls.cloudfront.net/${fileName}`;
const fileLocation = `fixtures/${fileName}`;

test.describe('Files', () => {
  let navbar;
  test.beforeAll(async () => {
    // download a fresh version of the demo artifact
    const response = await fetch(demoArtifactLocation);
    const buffer = await response.arrayBuffer();
    fs.writeFileSync(fileLocation, Buffer.from(buffer));
  });
  test.beforeEach(async ({ browserName, page }) => {
    navbar = page.locator('.leftFixed.leftNav');
    await navbar.getByRole('link', { name: /Releases/i }).click({ force: browserName === 'webkit' });
  });

  test('allows file removal', async ({ page, environment }) => {
    if (!isEnterpriseOrStaging(environment)) {
      const uploadButton = await page.getByRole('button', { name: /upload/i });
      await uploadButton.click();
      const drawer = page.locator(`.MuiDialog-paper`);
      await drawer.locator('.dropzone input').setInputFiles(fileLocation);
      await drawer.getByRole('button', { name: /Upload/i }).click();
      await page.getByText(/mender-demo-artifact/i).waitFor();
    }
    await page.getByRole('checkbox').first().click();
    await page.click('.MuiSpeedDial-fab');
    await page.getByLabel(/delete release/i).click();
    await expect(page.getByText(/will be deleted/i)).toBeVisible();
    await page.getByRole('button', { name: /delete/i }).click();
    await expect(page.getByText(/There are no Releases yet/i)).toBeVisible();
  });

  test('allows file uploads', async ({ page }) => {
    const uploadButton = await page.getByRole('button', { name: /upload/i });
    await uploadButton.click();
    const drawer = page.locator(`.MuiDialog-paper`);
    await drawer.locator('.dropzone input').setInputFiles(fileLocation);
    await drawer.getByRole('button', { name: /Upload/i }).click();
    await page.getByText(/last modified/i).waitFor();
  });

  test.describe('downloads', () => {
    test.describe.configure({ retries: 2 });
    test('allows artifact downloads', async ({ demoArtifactVersion, page, request }) => {
      await page.getByText(/mender-demo-artifact/i).click();
      await page.click('.expandButton');
      const downloadButton = await page.getByText(/download artifact/i);
      await expect(downloadButton).toBeVisible();
      const downloadPromise = page.waitForEvent('download');
      await downloadButton.click();
      const download = await downloadPromise;
      let downloadTargetPath;
      const downloadError = await download.failure();
      if (downloadError) {
        const downloadUrl = download.url();
        const response = await request.get(downloadUrl);
        const fileData = await response.body();
        downloadTargetPath = `./${download.suggestedFilename()}`;
        fs.writeFileSync(downloadTargetPath, fileData);
      } else {
        downloadTargetPath = await download.path();
      }
      const stdout = execSync(`mender-artifact read --no-progress ${downloadTargetPath}`);
      const artifactInfo = parse(stdout.toString());
      // Parse artifact header to check that artifact name matches
      const artifactName = artifactInfo['Mender Artifact'].Name;
      expect(artifactName).toMatch(/^mender-demo-artifact/);
      const versionInfo = artifactName.substring(artifactName.indexOf(expectedArtifactName) + expectedArtifactName.length + 1);
      expect(versionInfo).toEqual(demoArtifactVersion.artifactVersion);
      const { 'data-partition.mender-demo-artifact.version': updateVersion } = artifactInfo.Updates[0].Provides;
      expect(updateVersion).toEqual(demoArtifactVersion.updateVersion);
    });
  });

  test('allows artifact generation', async ({ baseUrl, browserName, page, request }) => {
    const hasTaggedRelease = await page.getByText(/customRelease/i).isVisible();
    if (hasTaggedRelease) {
      return;
    }
    const releaseName = 'terminalImage';
    const uploadButton = await page.getByRole('button', { name: /upload/i });
    await uploadButton.click();
    await page.locator('.MuiDialog-paper .dropzone input').setInputFiles(`fixtures/terminalContent.png`);
    await page.getByPlaceholder(/installed-by-single-file/i).fill(`/usr/src`);
    const deviceTypeInput = await page.getByLabel(/Release name/i);
    await deviceTypeInput.clear();
    await deviceTypeInput.fill(releaseName);
    await page.getByLabel(/Device types/i).fill(`all-of-them,`);
    await page.getByRole('button', { name: /next/i }).click();
    await page.getByRole('button', { name: /upload artifact/i }).click();
    await page.getByText('1-2 of 2').waitFor();
    const token = await getTokenFromStorage(baseUrl);
    await tagRelease(releaseName, 'customRelease', baseUrl, token, request);
    await page.waitForTimeout(timeouts.oneSecond); // some extra time for the release to be tagged in the backend
    await page.keyboard.press('Escape');
    await page.reload();
    await navbar.getByRole('link', { name: /Releases/i }).click({ force: browserName === 'webkit' });
    await expect(page.getByText(/customRelease/i)).toBeVisible();
  });

  test('allows release notes manipulation', async ({ page }) => {
    await page.getByText(/terminalimage/i).click();
    await expect(page.getByRole('heading', { name: /Release notes/i })).toBeVisible();
    const hasNotes = await page.getByText('foo notes');
    if (await hasNotes.isVisible()) {
      return;
    }
    // layout based locators are not an option here, since the edit button is also visible on the nearby tags section
    // and the selector would get confused due to the proximity - so instead we loop over all the divs
    await page
      .locator('div')
      .filter({ hasText: /^Add release notes here Edit$/i })
      .getByRole('button')
      .click();
    const input = await page.getByPlaceholder(/release notes/i);
    await input.fill('foo notes');
    await page.getByRole('button', { name: 'confirm' }).click();
    await expect(input).not.toBeVisible();
    await expect(hasNotes).toBeVisible();
  });

  test('allows release tags manipulation', async ({ baseUrl, page }) => {
    const alreadyTagged = await page.getByText(selectors.releaseTags).isVisible();
    test.skip(alreadyTagged, 'looks like the release was tagged already');
    await page.getByText(/demo-artifact/i).click();
    await expect(page.getByRole('heading', { name: /Release notes/i })).toBeVisible();
    await expect(page.getByRole('button', { name: 'some' })).not.toBeVisible();
    // layout based locators are not an option here, since the edit button is also visible on the nearby release notes section
    // and the selector would get confused due to the proximity - so instead we loop over all the divs
    const theDiv = await page
      .locator('div')
      .filter({ has: page.getByRole('heading', { name: /tags/i }), hasNotText: /notes/i })
      .filter({ has: page.getByRole('button', { name: /edit/i }) });
    const editButton = await theDiv.getByRole('button', { name: /edit/i });
    await editButton.click();
    const input = await page.getByPlaceholder(/enter release tags/i);
    await input.pressSequentially('some,tags', { delay: 300 });
    await page.getByRole('button', { name: 'confirm' }).click();
    await page.waitForTimeout(timeouts.oneSecond);
    await expect(input).not.toBeVisible();
    await page.goto(`${baseUrl}ui/releases`);
    await page.getByText(selectors.releaseTags).waitFor({ timeout: timeouts.default });
    await expect(page.getByText(selectors.releaseTags)).toBeVisible();
  });

  test('allows release tags reset', async ({ page }) => {
    await page.getByText(/demo-artifact/i).click();
    const theDiv = await page
      .locator('div')
      .filter({ has: page.getByRole('heading', { name: /tags/i }), hasNotText: /notes/ })
      .filter({ has: page.getByRole('button', { name: /edit/i }) });
    const editButton = await theDiv.getByRole('button', { name: /edit/i });
    await editButton.click();
    const alreadyTagged = await page.getByRole('button', { name: 'some' }).isVisible();
    if (alreadyTagged) {
      for await (const name of ['some', 'tags']) {
        const foundTag = await page.getByRole('button', { name });
        if (!(await foundTag.isVisible())) {
          continue;
        }
        await foundTag.getByLabel('tags-delete').click();
      }
      await page.getByRole('button', { name: 'confirm' }).click();
      await page.getByPlaceholder(/add release tags/i).waitFor({ timeout: timeouts.oneSecond });
      await expect(page.getByPlaceholder(/add release tags/i)).toBeVisible();
      await editButton.click();
    }
    await page.getByPlaceholder(/enter release tags/i).pressSequentially(releaseTag, { delay: 100 });
    await page.getByRole('button', { name: 'confirm' }).click();
    await page.getByLabel(/close/i).click();
    await page.waitForTimeout(timeouts.default);
    await page.getByText('Upload').isVisible({ timeout: timeouts.default });
    await page.screenshot({ path: './test-results/releasetag-reset.png' });
    await expect(page.getByText(releaseTag, { exact: false })).toBeVisible();
  });

  test('allows release tags filtering', async ({ page }) => {
    await expect(page.getByText(releaseTag.toLowerCase())).toBeVisible();
    await page.getByPlaceholder(/select tags/i).fill('foo,');
    const releasesNote = await page.getByText(/There are no Releases*/i);
    releasesNote.waitFor({ timeout: timeouts.default });
    await page.getByText(/mender-demo-artifact*/i).waitFor({ timeout: timeouts.default, state: 'detached' });
    await page.getByText(/Clear filter/i).click();
    await page.getByText(/mender-demo-artifact*/i).waitFor();
    await expect(page.getByText(releaseTag.toLowerCase())).toBeVisible();
    await page.getByPlaceholder(/select tags/i).fill(`${releaseTag.toLowerCase()},`);
    await page.getByText(/mender-demo-artifact*/i).waitFor({ timeout: timeouts.default });
    await expect(releasesNote).not.toBeVisible();
  });

  // test('allows uploading custom file creations', () => {
  //   cy.exec('mender-artifact write rootfs-image -f core-image-full-cmdline-qemux86-64.ext4 -t qemux86-64 -n release1 -o qemux86-64_release_1.mender')
  //     .then(result => {
  //       expect(result.code).to.be.equal(0)
  //         const encoding = 'base64'
  //         const fileName = 'qemux86-64_release_1.mender'
  //         cy.readFile(fileName, encoding).then(fileContent => {
  //           cy.get('.dropzone input')
  //             .upload({ fileContent, fileName, encoding, mimeType: 'application/octet-stream' })
  //             .wait(10000) // give some extra time for the upload
  //         })
  //       })
  // })
});
