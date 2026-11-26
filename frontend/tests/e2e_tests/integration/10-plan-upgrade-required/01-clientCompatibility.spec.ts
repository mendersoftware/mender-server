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
import type { APIRequestContext, Page, PlaywrightTestArgs } from '@playwright/test';
import { createHash, randomBytes, randomInt } from 'crypto';
import { existsSync, readFileSync, writeFileSync } from 'fs';

import type { CombinedFixtures } from '../../fixtures/fixtures';
import baseTest, { expect } from '../../fixtures/fixtures';
import { login } from '../../utils/commands';
import type { CompatibleClient } from '../../utils/compat';
import { compatibleClients, startCompatClients } from '../../utils/compat';
import { storageFolder, timeouts } from '../../utils/constants';

const md5 = (content: Buffer) => createHash('md5').update(content).digest('hex');

// Older client versions listed in fixtures/docker-compose.compat.yml are started at this late point of the run, so the earlier tests
// can rely on the single demo device.

const openDeviceByHostname = async (page: Page, hostname: string) => {
  await page.getByRole('button', { name: /find a device/i }).click();
  const searchDialog = page.getByRole('dialog');
  await searchDialog.getByRole('combobox').click();
  await page.getByRole('option', { name: /^Any/ }).click();
  const searchField = searchDialog.getByPlaceholder(/starting with/i);
  const result = searchDialog.getByRole('button', { name: hostname });
  // the client might not have sent its authentication request yet
  await expect(async () => {
    await searchField.fill('');
    await searchField.fill(hostname);
    await expect(result).toBeVisible({ timeout: timeouts.fiveSeconds });
  }).toPass({ timeout: 2 * timeouts.sixtySeconds });
  await result.click();
  await page.getByText(/device information/i).waitFor();
};

type CompatState = { deploymentId?: string; devices: Record<string, string> };
const compatStatePath = `${storageFolder}/compat-state.json`;
const readCompatState = (): CompatState => (existsSync(compatStatePath) ? JSON.parse(readFileSync(compatStatePath, 'utf-8')) : { devices: {} });
const storeAcceptedDevice = (hostname: string, deviceId: string) => {
  const state = readCompatState();
  writeFileSync(compatStatePath, JSON.stringify({ ...state, devices: { ...state.devices, [hostname]: deviceId } }));
};
const storeDeploymentId = (deploymentId: string) => writeFileSync(compatStatePath, JSON.stringify({ ...readCompatState(), deploymentId }));

// A single-file update that is compatible with the docker-client device type of the compat clients. The mender-demo-artifact can't be
// used, as 2.0.x clients fail to read Artifacts with payload provides ("type_info provides values not yet supported").
const artifactPath = 'fixtures/date.mender';
const artifactName = 'date';
const finishedDeviceStatusPattern = /Success|Fail|Aborted|No compatible artifact found|Already installed|Decommissioned|Skipped|Incompatible tier/;

// mender-connect versions before the simplified filetransfer protocol was introduced
const legacyConnectVersion = '3.0.0';

type DevauthDevice = { auth_sets: { id: string }[]; id: string; identity_data: Record<string, string>; status: string };

type LegacyDevice = { deviceId: string; headers: Record<string, string> };

type CompatWorkerFixtures = { legacyClient: CompatibleClient; legacyDevice: LegacyDevice };

type SetupArgs = Pick<CombinedFixtures, 'baseUrl' | 'password' | 'username'> & Pick<CompatWorkerFixtures, 'legacyClient'> & Pick<PlaywrightTestArgs, 'request'>;

const setupLegacyDevice = async ({ baseUrl, legacyClient, password, request, username }: SetupArgs): Promise<LegacyDevice> => {
  const { token } = await login(username, password, baseUrl, request);
  const headers = { Authorization: `Bearer ${token}` };
  const { hostname } = legacyClient;
  let device: DevauthDevice | undefined;
  await expect
    .poll(
      async () => {
        const response = await request.get(`${baseUrl}api/management/v2/devauth/devices`, { headers, params: { per_page: 500 } });
        const devices: DevauthDevice[] = response.ok() ? await response.json() : [];
        device = devices.find(({ identity_data }) => identity_data.hostname === hostname);
        return !!device;
      },
      { message: `waiting for device ${hostname} to show up`, timeout: 2 * timeouts.sixtySeconds }
    )
    .toBeTruthy();
  const { auth_sets: authSets, id: deviceId, status } = device!;
  if (status === 'pending') {
    const response = await request.put(`${baseUrl}api/management/v2/devauth/devices/${deviceId}/auth/${authSets[0].id}/status`, {
      headers,
      data: { status: 'accepted' }
    });
    expect(response.status()).toBe(204);
  }
  await expect
    .poll(
      async () => {
        const response = await request.get(`${baseUrl}api/management/v1/deviceconnect/devices/${deviceId}`, { headers });
        return response.ok() ? (await response.json()).status : response.status();
      },
      { message: `waiting for device ${hostname} to connect`, timeout: 2 * timeouts.sixtySeconds }
    )
    .toBe('connected');
  return { deviceId, headers };
};

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
const test = baseTest.extend<{}, CompatWorkerFixtures>({
  legacyClient: [
    // eslint-disable-next-line no-empty-pattern
    async ({}, use) => {
      const legacyClient = compatibleClients.find(({ connectVersion }) => connectVersion === legacyConnectVersion);
      expect(legacyClient, `fixtures/docker-compose.compat.yml should label a client with mender-connect ${legacyConnectVersion}`).toBeDefined();
      await use(legacyClient!);
    },
    { scope: 'worker' }
  ],
  legacyDevice: [
    async ({ baseUrl, legacyClient, password, playwright, username }, use) => {
      const request = await playwright.request.newContext({ ignoreHTTPSErrors: true });
      const legacyDevice = await setupLegacyDevice({ baseUrl, legacyClient, password, request, username });
      await request.dispose();
      await use(legacyDevice);
    },
    { scope: 'worker', timeout: 5 * timeouts.sixtySeconds }
  ]
});

test.describe('Client compatibility', () => {
  test.skip(({ isScheduledRun }) => !isScheduledRun, 'client compatibility tests are only run on scheduled pipelines');
  test.describe.configure({ timeout: 5 * timeouts.sixtySeconds });

  test('starts the compatible clients', async ({ baseUrl, tenantToken }) => {
    test.setTimeout(10 * timeouts.sixtySeconds);
    await startCompatClients(baseUrl, tenantToken);
  });

  for (const client of compatibleClients) {
    const { hostname } = client;
    test(`${client.image} can be authorized`, async ({ page }) => {
      await openDeviceByHostname(page, hostname);
      const deviceId = new URL(page.url()).searchParams.get('id');
      expect(deviceId, `no device id found in ${page.url()}`).toBeTruthy();
      const expandedDevice = page.locator('.expandedDevice');
      const acceptButton = expandedDevice.locator('.action-buttons').getByRole('button', { name: 'Accept', exact: true });
      await acceptButton.waitFor();
      if (await acceptButton.isEnabled()) {
        await acceptButton.click();
      }
      await expect(expandedDevice.getByRole('button', { name: /decommission device/i })).toBeVisible({ timeout: timeouts.tenSeconds });
      storeAcceptedDevice(hostname, deviceId!);
    });
  }

  test('deploys an artifact to the compatible clients', async ({ baseUrl, password, request, username }) => {
    const deviceIds = Object.values(readCompatState().devices);
    expect(deviceIds.length, 'none of the compatible clients got accepted').toBeGreaterThan(0);
    const { token } = await login(username, password, baseUrl, request);
    const headers = { Authorization: `Bearer ${token}` };

    const uploadResponse = await request.post(`${baseUrl}api/management/v1/deployments/artifacts`, {
      headers,
      multipart: { artifact: { name: 'date.mender', mimeType: 'application/octet-stream', buffer: readFileSync(artifactPath) } }
    });
    expect([201, 409], await uploadResponse.text()).toContain(uploadResponse.status());

    const deploymentResponse = await request.post(`${baseUrl}api/management/v1/deployments/deployments`, {
      headers,
      data: { artifact_name: artifactName, devices: deviceIds, name: 'client compatibility' }
    });
    expect(deploymentResponse.status(), await deploymentResponse.text()).toBe(201);
    const location = deploymentResponse.headers()['location'];
    storeDeploymentId(location.substring(location.lastIndexOf('/') + 1));
  });

  test('installs the deployed artifact on the compatible clients', async ({ baseUrl, page }) => {
    const { deploymentId, devices } = readCompatState();
    expect(deploymentId, 'no deployment was created for the compatible clients').toBeTruthy();
    await page.goto(`${baseUrl}ui/deployments?open=true&id=${deploymentId}`);
    const report = page.locator('.MuiDrawer-paper');
    // show all devices of the deployment on a single page
    await report.getByRole('combobox', { name: /rows per page/i }).click();
    await page.getByRole('option').last().click();
    const deviceRows = Object.entries(devices).map(([hostname, deviceId]) => ({
      hostname,
      row: report.getByRole('row').filter({ has: page.locator(`a[href$="devices?id=${deviceId}"]`) })
    }));
    // allTextContents doesn't wait for the rows, so devices that aren't listed yet count as unfinished
    const readStatuses = () =>
      Promise.all(
        deviceRows.map(async ({ hostname, row }) => ({ hostname, status: (await row.allTextContents()).join('').match(finishedDeviceStatusPattern)?.[0] }))
      );
    await expect
      .poll(async () => (await readStatuses()).filter(({ status }) => !status).map(({ hostname }) => hostname), {
        message: 'waiting for all compatible clients to finish the deployment',
        intervals: [timeouts.fiveSeconds],
        timeout: 4 * timeouts.sixtySeconds
      })
      .toEqual([]);
    const unsuccessful = (await readStatuses()).filter(({ status }) => status !== 'Success');
    expect(unsuccessful, 'the deployment did not succeed on all compatible clients').toEqual([]);
  });

  for (const client of compatibleClients) {
    const { hostname } = client;
    test(`${client.image} reports its inventory`, async ({ baseUrl, page }) => {
      const deviceId = readCompatState().devices[hostname];
      expect(deviceId, `${hostname} was not accepted`).toBeTruthy();
      await page.goto(`${baseUrl}ui/devices/accepted?id=${deviceId}&tab=inventory`);
      const expandedDevice = page.locator('.expandedDevice');
      const getAttributeValue = (attribute: string, value: string) =>
        expandedDevice
          .getByRole('row')
          .filter({ has: page.getByRole('cell', { name: attribute, exact: true }) })
          .getByRole('cell', { name: value, exact: true });
      await expect.soft(getAttributeValue('mender_client_version', client.version)).toBeVisible({ timeout: 3 * timeouts.sixtySeconds });
      await expect(getAttributeValue('artifact_name', artifactName), `${hostname} does not report the deployed artifact as installed`).toBeVisible({
        timeout: timeouts.sixtySeconds
      });
    });
  }
});

const downloadFile = (request: APIRequestContext, baseUrl: string, { deviceId, headers }: LegacyDevice, path: string) =>
  request.get(`${baseUrl}api/management/v1/deviceconnect/devices/${deviceId}/download`, { headers, params: { path } });

const uploadFile = (
  request: APIRequestContext,
  baseUrl: string,
  { deviceId, headers }: LegacyDevice,
  path: string,
  content: Buffer,
  { mode = '600', uid = '0', gid = '0' } = {}
) =>
  request.put(`${baseUrl}api/management/v1/deviceconnect/devices/${deviceId}/upload`, {
    headers,
    multipart: { path, mode, uid, gid, file: { name: path.substring(path.lastIndexOf('/') + 1), mimeType: 'application/octet-stream', buffer: content } },
    timeout: 2 * timeouts.sixtySeconds
  });

test.describe('filetransfer with a legacy mender-connect', () => {
  test.skip(({ isScheduledRun }) => !isScheduledRun, 'the legacy mender-connect client is only started on scheduled pipelines');
  // deviceconnect negotiates the filetransfer protocol version with mender-connect and must keep supporting older clients.
  // Since this is about the protocol these checks are API only.
  test('allows downloading a file', async ({ baseUrl, legacyDevice: device, request }) => {
    const path = '/etc/mender/mender.conf';
    const response = await downloadFile(request, baseUrl, device, path);
    expect(response.status(), await response.text()).toBe(200);
    const content = await response.body();
    expect(content.toString()).toContain('ServerURL');
    const responseHeaders = response.headers();
    expect(responseHeaders['content-disposition']).toEqual('attachment; filename="mender.conf"');
    expect(responseHeaders['content-type']).toEqual('application/octet-stream');
    expect(responseHeaders['x-men-file-path']).toEqual(path);
    expect(responseHeaders['x-men-file-size']).toEqual(`${content.length}`);
  });

  test('rejects downloads from relative paths', async ({ baseUrl, legacyDevice: device, request }) => {
    const response = await downloadFile(request, baseUrl, device, 'relative/path');
    expect(response.status(), await response.text()).toBe(400);
    expect((await response.json()).error).toEqual('path: must be absolute.');
  });

  test('reports downloads of missing files', async ({ baseUrl, legacyDevice: device, request }) => {
    const response = await downloadFile(request, baseUrl, device, '/does/not/exist');
    expect(response.status(), await response.text()).toBe(400);
    expect((await response.json()).error).toContain('/does/not/exist: no such file or directory');
  });

  test('transfers a large file intact in both directions', async ({ baseUrl, legacyDevice: device, request }) => {
    // big enough to span many chunks and exercise the acknowledgement sliding window of the legacy protocol
    const content = randomBytes(40 * 1024 * 1024);
    const uid = `${randomInt(100, 201)}`;
    const gid = `${randomInt(100, 201)}`;
    const fileName = `compat-${randomBytes(4).toString('hex')}`;
    const path = `/tmp/${fileName}`;

    const uploadResponse = await uploadFile(request, baseUrl, device, path, content, { mode: '600', uid, gid });
    expect(uploadResponse.status(), await uploadResponse.text()).toBe(201);

    const response = await downloadFile(request, baseUrl, device, path);
    expect(response.status()).toBe(200);
    const responseHeaders = response.headers();
    expect(responseHeaders['content-disposition']).toEqual(`attachment; filename="${fileName}"`);
    expect(responseHeaders['content-type']).toEqual('application/octet-stream');
    expect(responseHeaders['x-men-file-mode']).toEqual('600');
    expect(responseHeaders['x-men-file-uid']).toEqual(uid);
    expect(responseHeaders['x-men-file-gid']).toEqual(gid);
    expect(responseHeaders['x-men-file-path']).toEqual(path);
    expect(responseHeaders['x-men-file-size']).toEqual(`${content.length}`);
    expect(md5(await response.body())).toEqual(md5(content));
  });

  test('rejects uploads to relative paths', async ({ baseUrl, legacyDevice: device, request }) => {
    const response = await uploadFile(request, baseUrl, device, 'relative/path/dummy.txt', Buffer.from('dummy'));
    expect(response.status(), await response.text()).toBe(400);
    expect((await response.json()).error).toEqual('path: must be absolute.');
  });

  test('reports uploads to missing directories', async ({ baseUrl, legacyDevice: device, request }) => {
    const response = await uploadFile(request, baseUrl, device, '/does/not/exist/dummy.txt', Buffer.from('dummy'));
    expect(response.status(), await response.text()).toBe(400);
    expect((await response.json()).error).toContain('failed to create target file');
  });
});
