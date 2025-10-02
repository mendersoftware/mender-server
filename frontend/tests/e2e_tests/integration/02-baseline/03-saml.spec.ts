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
import dns from 'node:dns';

import test, { expect } from '../../fixtures/fixtures';
import { getTokenFromStorage, isEnterpriseOrStaging, isLoggedIn, startIdpServer } from '../../utils/commands';
import { timeouts } from '../../utils/constants';

dns.setDefaultResultOrder('ipv4first');

const samlSettings = {
  credentials: {
    chromium: 'saml.jackson@example.com',
    firefox: 'sam.l.jackson@example.com',
    webkit: 'samu.l.jackson@example.com'
  },
  idpUrl: 'http://localhost:7000/metadata'
};

let acsUrl = '';
let metadataLocation = '';

test.describe('SAML Login via sso/id/login', () => {
  test.afterAll(async ({ environment, baseUrl, browserName, request }, testInfo) => {
    if (testInfo.status === 'skipped' || !isEnterpriseOrStaging(environment)) {
      return;
    }
    const token = await getTokenFromStorage(baseUrl);
    const options = { headers: { Authorization: `Bearer ${token}` } };
    console.log(`Finished ${testInfo.title} with status ${testInfo.status}. Cleaning up.`);
    const response = await request.get(`${baseUrl}api/management/v1/useradm/users?email=${encodeURIComponent(samlSettings.credentials[browserName])}`, options);
    const users = await response.json();
    if (!response.ok() || !users.length) {
      console.log(`${samlSettings.credentials[browserName]} does not exist.`);
      return;
    }
    const { id: userId } = users[0];
    await request.delete(`${baseUrl}api/management/v1/useradm/users/${userId}`, options);
    console.log(`removed user ${samlSettings.credentials[browserName]}.`);
  });

  // Setups the SAML/SSO login with samltest.id Identity Provider
  test('Set up SAML', async ({ browserName, environment, baseUrl, page, request }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    // allow a lot of time to enter metadata + then some to handle uploading the config to the external service
    test.setTimeout(5 * timeouts.sixtySeconds + timeouts.fifteenSeconds);

    let idpServer;
    startIdpServer({}, server => (idpServer = server));
    await page.waitForTimeout(timeouts.oneSecond);
    const response = await request.get(samlSettings.idpUrl);
    idpServer.close();
    expect(response.ok()).toBeTruthy();
    const metadata = await response.text();
    await page.goto(`${baseUrl}ui/settings/organization`);
    const isInitialized = await page.getByText('Entity ID').isVisible();
    if (!isInitialized) {
      // Check input[type="checkbox"]
      await page.getByLabel(/Enable Single Sign-On/i).click();
      await page.getByRole('combobox').click();
      await page.getByRole('option', { name: 'SAML' }).click();
      // Click text=input with the text editor
      await page.getByText('input with the text editor').click();

      const textfield = await page.getByLabel(/editor content/i);
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
    await dialog.getByLabel('close').click();
    const token = await getTokenFromStorage(baseUrl);
    const options = { headers: { Authorization: `Bearer ${token}` } };
    const storedMetadataResponse = await request.get(`${baseUrl}api/management/v1/useradm/sso/idp/metadata`, options);
    const data = await storedMetadataResponse.json();
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
    const spMetadataResponse = await request.get(expectedSpMetaUrl, options);
    expect(spMetadataResponse.ok()).toBeTruthy();
    const spMetadata = await spMetadataResponse.text();
    expect(spMetadata).toContain('SPSSODescriptor');
    idpServer.close();
    await page.waitForTimeout(timeouts.oneSecond);
  });

  // Creates a user with login that matches Identity privder (samltest.id) user email
  test('Creates a user without a password', async ({ environment, baseUrl, browserName, page }) => {
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
  test('User can login via sso/login endpoint', async ({ environment, baseUrl, browser, browserName, page }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    test.setTimeout(3 * timeouts.fifteenSeconds);
    let idpServer;
    startIdpServer({ acsUrl, metadataLocation }, server => (idpServer = server));
    await page.waitForTimeout(timeouts.oneSecond);
    await page.goto(`${baseUrl}ui/help`);
    await page.goto(`${baseUrl}ui/settings`);
    await page.getByText(/organization/i).click();
    await page.getByText('View metadata in the text editor').waitFor({ timeout: timeouts.tenSeconds });
    let loginUrl = '';
    let loginThing = await page.locator('*:below(:text("Start URL"))').first();
    loginUrl = await loginThing.getAttribute('title');
    if (!loginUrl) {
      loginThing = await page.locator(':text("Start URL") + *').first();
      loginUrl = await loginThing.innerText();
    }
    console.log(`logging in via ${loginUrl} (using: ${samlSettings.credentials[browserName]})`);
    const context = await browser.newContext();
    const samlPage = await context.newPage();
    await samlPage.goto(loginUrl);
    // This screenshot saves the view right after the first redirection
    await samlPage.screenshot({ path: './test-results/saml-redirected.png' });

    await samlPage.getByLabel(/Subject NameID/i).clear();
    await samlPage.getByLabel(/Subject NameID/i).fill(samlSettings.credentials[browserName]);
    await samlPage.getByLabel(/E-Mail Address/i).clear();
    await samlPage.getByLabel(/E-Mail Address/i).fill(samlSettings.credentials[browserName]);
    await samlPage.getByRole('button', { name: /sign in/i }).click();
    // confirm we have logged in successfully
    await samlPage.screenshot({ path: './test-results/saml-logging-in-accept.png' });
    await isLoggedIn(samlPage);
    idpServer.close();
    await context.close();
  });
});
