// Copyright 2022 Northern.tech AS
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
import axios from 'axios';
import dns from 'node:dns';
import * as https from 'node:https';

import test, { expect } from '../fixtures/fixtures.ts';
import { getTokenFromStorage, isEnterpriseOrStaging, isLoggedIn, startIdpServer } from '../utils/commands.ts';
import { storagePath, timeouts } from '../utils/constants.ts';

dns.setDefaultResultOrder('ipv4first');

const samlSettings = {
  credentials: {
    chromium: 'saml.jackson@example.com',
    firefox: 'sam.l.jackson@example.com',
    webkit: 'samu.l.jackson@example.com'
  },
  idpUrl: 'http://localhost:7000/metadata'
};

const httpsAgent = new https.Agent({ rejectUnauthorized: false });
const defaultHeaders = { 'Content-Type': 'application/json' };

let acsUrl = '';
let metadataLocation = '';

test.describe('SAML Login via sso/id/login', () => {
  test.use({ storageState: storagePath });
  test.afterAll(async ({ environment, baseUrl, browserName }, testInfo) => {
    if (testInfo.status === 'skipped' || !isEnterpriseOrStaging(environment)) {
      return;
    }
    const token = await getTokenFromStorage(baseUrl);
    const requestInfo = { headers: { ...defaultHeaders, Authorization: `Bearer ${token}` }, httpsAgent, method: 'GET' };
    console.log(`Finished ${testInfo.title} with status ${testInfo.status}. Cleaning up.`);
    const response = await axios({
      ...requestInfo,
      url: `${baseUrl}api/management/v1/useradm/users?email=${encodeURIComponent(samlSettings.credentials[browserName])}`
    });
    if (response.status >= 300 || !response.data.length) {
      console.log(`${samlSettings.credentials[browserName]} does not exist.`);
      return;
    }
    const { id: userId } = response.data[0];
    await axios({
      ...requestInfo,
      url: `${baseUrl}api/management/v1/useradm/users/${userId}`,
      method: 'DELETE'
    });
    console.log(`removed user ${samlSettings.credentials[browserName]}.`);
  });

  // Setups the SAML/SSO login with samltest.id Identity Provider
  test('Set up SAML', async ({ browserName, environment, baseUrl, loggedInPage: page }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    // allow a lot of time to enter metadata + then some to handle uploading the config to the external service
    test.setTimeout(5 * timeouts.sixtySeconds + timeouts.fifteenSeconds);

    let idpServer;
    startIdpServer({}, server => (idpServer = server));
    await page.waitForTimeout(timeouts.oneSecond);
    const { data: metadata, status } = await axios({ url: samlSettings.idpUrl, method: 'GET' });
    idpServer.close();
    expect(status).toBeGreaterThanOrEqual(200);
    expect(status).toBeLessThan(300);
    await page.goto(`${baseUrl}ui/settings/organization-and-billing`);
    const isInitialized = await page.getByText('Entity ID').isVisible();
    if (!isInitialized) {
      // Check input[type="checkbox"]
      await page.getByLabel(/Enable Single Sign-On/i).click();
      await page.getByRole('combobox').click();
      await page.getByRole('option', { name: 'SAML' }).click();
      // Click text=input with the text editor
      await page.getByText('input with the text editor').click();

      const textfield = await page.locator('[aria-label="Editor content\\;Press Alt\\+F1 for Accessibility Options\\."]');
      const cleanedMetaData = metadata.replace(/(?:\r\n|\r|\n)/g, '');
      if (browserName === 'firefox') {
        await textfield.pressSequentially(cleanedMetaData);
      } else {
        await textfield.fill(cleanedMetaData);
      }
      console.log('typing metadata done.');
      // The screenshot saves the view of the typed metadata
      await page.screenshot({ 'path': './test-results/saml-edit-saving.png' });
      // Click text=Save >> nth=1
      await page.getByText('Save').nth(1).click();
      await page.getByText('Entity ID').waitFor();
    }

    await page.getByText('View metadata in the text editor').click();
    // Click text=Download file
    const [download] = await Promise.all([page.waitForEvent('download'), page.getByRole('button', { name: /download file/i }).click()]);
    const downloadTargetPath = await download.path();
    expect(downloadTargetPath).toBeTruthy();
    const dialog = await page.locator('text=SAML metadata >> .. >> ..');
    await dialog.locator('data-testid=CloseIcon').click();
    const token = await getTokenFromStorage(baseUrl);
    const requestInfo = { method: 'GET', headers: { ...defaultHeaders, Authorization: `Bearer ${token}` }, httpsAgent };
    const { data } = await axios({ ...requestInfo, url: `${baseUrl}api/management/v1/useradm/sso/idp/metadata` });
    const metadataId = data[0].id;
    console.log(`looking for config info for metadata id: ${metadataId}`);
    const expectedLoginUrl = `${baseUrl}api/management/v1/useradm/auth/sso/${metadataId}/login`;
    const loginUrl = await page.getByText(expectedLoginUrl);
    await loginUrl.waitFor();
    await expect(loginUrl).toBeVisible();
    const expectedAcsUrl = `${baseUrl}api/management/v1/useradm/auth/sso/${metadataId}/acs`;
    await expect(page.getByText(expectedAcsUrl)).toBeVisible();
    const expectedSpMetaUrl = `${baseUrl}api/management/v1/useradm/sso/sp/metadata/${metadataId}`;
    await expect(page.getByText(expectedSpMetaUrl)).toBeVisible();
    acsUrl = expectedAcsUrl;
    metadataLocation = expectedSpMetaUrl;

    const { data: spMetadata, status: spDataStatus } = await axios({ ...requestInfo, url: expectedSpMetaUrl });
    expect(spDataStatus).toBeGreaterThanOrEqual(200);
    expect(spDataStatus).toBeLessThan(300);
    expect(spMetadata).toContain('SPSSODescriptor');
    idpServer.close();
    await page.waitForTimeout(timeouts.oneSecond);
  });

  // Creates a user with login that matches Identity privder (samltest.id) user email
  test('Creates a user without a password', async ({ environment, baseUrl, browserName, loggedInPage: page }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    await page.goto(`${baseUrl}ui/settings/user-management`);
    const userExists = await page.getByText(samlSettings.credentials[browserName]).isVisible();
    if (userExists) {
      console.log(`${samlSettings.credentials[browserName]} already exists.`);
      return;
    }
    await page.getByRole('button', { name: /new user/i }).click();
    await page.getByPlaceholder(/Email/i).click();
    await page.getByPlaceholder(/Email/i).fill(samlSettings.credentials[browserName]);
    // Click text=Create user
    await page.getByRole('button', { name: /Create user/i }).click();
    await page.screenshot({ path: './test-results/user-created.png' });
    await page.getByText('The user was created successfully.').waitFor();
  });

  // This test calls auth/sso/${id}/login, where id is the id of the identity provider
  // and verifies that login is successful.
  test('User can login via sso/login endpoint', async ({ environment, baseUrl, browser, browserName, loggedInPage }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    test.setTimeout(3 * timeouts.fifteenSeconds);
    let idpServer;
    startIdpServer({ acsUrl, metadataLocation }, server => (idpServer = server));
    await loggedInPage.waitForTimeout(timeouts.oneSecond);
    await loggedInPage.goto(`${baseUrl}ui/help`);
    await loggedInPage.goto(`${baseUrl}ui/settings`);
    await loggedInPage.getByText(/organization/i).click();
    await loggedInPage.getByText('View metadata in the text editor').waitFor({ timeout: timeouts.tenSeconds });
    let loginUrl = '';
    let loginThing = await loggedInPage.locator('*:below(:text("Start URL"))').first();
    loginUrl = await loginThing.getAttribute('title');
    if (!loginUrl) {
      loginThing = await loggedInPage.locator(':text("Start URL") + *').first();
      loginUrl = await loginThing.innerText();
    }
    console.log(`logging in via ${loginUrl} (using: ${samlSettings.credentials[browserName]})`);
    const context = await browser.newContext();
    const page = await context.newPage();
    await page.goto(loginUrl);
    // This screenshot saves the view right after the first redirection
    await page.screenshot({ path: './test-results/saml-redirected.png' });

    await page.getByLabel(/Subject NameID/i).clear();
    await page.getByLabel(/Subject NameID/i).fill(samlSettings.credentials[browserName]);
    await page.getByLabel(/E-Mail Address/i).clear();
    await page.getByLabel(/E-Mail Address/i).fill(samlSettings.credentials[browserName]);
    await page.getByRole('button', { name: /sign in/i }).click();
    // confirm we have logged in successfully
    await page.screenshot({ path: './test-results/saml-logging-in-accept.png' });
    await isLoggedIn(page);
    idpServer.close();
  });
});
