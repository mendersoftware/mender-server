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
import { Browser, BrowserContext, Page } from '@playwright/test';
import { runServer } from '@sidewinder1138/saml-idp';
import axios from 'axios';
import { spawn } from 'child_process';
import * as fs from 'fs';
import * as https from 'https';
import { jwtDecode } from 'jwt-decode';
import { authenticator } from 'otplib';
import * as path from 'path';
import pixelmatch from 'pixelmatch';
import { PNG } from 'pngjs';
import { fileURLToPath } from 'url';
import { v4 as uuid } from 'uuid';

import { selectors, storagePath } from './constants.ts';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export const getPeristentLoginInfo = () => {
  let loginInfo;
  try {
    const content = fs.readFileSync('loginInfo.json', 'utf8');
    loginInfo = JSON.parse(content);
    return loginInfo;
  } catch {
    loginInfo = { username: process.env.STAGING_USER ?? `${uuid()}@example.com`, password: process.env.STAGING_PASSWORD ?? uuid() };
  }
  fs.writeFileSync('loginInfo.json', JSON.stringify(loginInfo));
  return loginInfo;
};

export const getStorageState = location => {
  let storageState;
  try {
    const content = fs.readFileSync(location, 'utf8');
    storageState = JSON.parse(content);
    return storageState;
  } catch {
    storageState = { username: process.env.STAGING_USER ?? `${uuid()}@example.com`, password: process.env.STAGING_PASSWORD ?? uuid() };
  }
  return storageState;
};

export const getTokenFromStorage = (baseUrl: string) => {
  const originUrl = baseUrl.endsWith('/') ? baseUrl.substring(0, baseUrl.length - 1) : baseUrl;
  const origin = getStorageState(storagePath).origins.find(({ origin }) => origin === originUrl);
  const textContent = origin?.localStorage.find(({ name }) => name === 'JWT').value;
  let sessionInfo = { token: '' };
  try {
    sessionInfo = JSON.parse(textContent);
  } catch {
    // most likely not logged in - nothing to do here
  }
  return sessionInfo.token;
};

export const prepareCookies = async (context: BrowserContext, domain: string, userId: string) => {
  await context.addCookies([
    { name: `${userId}-onboarded`, value: 'true', path: '/', domain },
    { name: 'cookieconsent_status', value: 'allow', path: '/', domain }
  ]);
  return context;
};

export const prepareNewPage = async ({
  baseUrl,
  browser,
  context: passedContext,
  password,
  username
}: {
  baseUrl: string;
  browser?: Browser;
  context?: BrowserContext;
  password: string;
  username: string;
}) => {
  let context = passedContext;
  if (!context) {
    context = await browser.newContext();
  }
  if (context.browser()?.browserType().name() === 'chromium') {
    await context.grantPermissions(['clipboard-read'], { origin: baseUrl });
  }
  let logInResult = { userId: '', token: '' };
  if (username && password) {
    logInResult = await login(username, password, baseUrl);
  }
  const domain = baseUrlToDomain(baseUrl);
  context = await prepareCookies(context, domain, logInResult.userId);
  await context.addInitScript(token => {
    window.localStorage.setItem('JWT', JSON.stringify({ token }));
    window.localStorage.setItem(`onboardingComplete`, 'true');
  }, logInResult.token);
  const page = await context.newPage();
  await page.goto(`${baseUrl}ui/`);
  return page;
};

const updateConfigFileWithUrl = (fileName, serverUrl = 'https://docker.mender.io', token = '') => {
  const connectConfigFile = fs.readFileSync(`dockerClient/${fileName}.json`, 'utf8');
  const connectConfig = JSON.parse(connectConfigFile);
  connectConfig.ServerURL = serverUrl;
  if (token) {
    connectConfig.TenantToken = token;
  }
  fs.writeFileSync(`dockerClient/${fileName}-test.json`, JSON.stringify(connectConfig));
};

const deviceType = 'qemux86-64';
const artifactName = 'original';
const updateInterval = 5;
const attributes = {
  device_type: deviceType,
  client_version: 'mender-2.2.0',
  artifact_name: artifactName,
  kernel: 'test Linux version',
  mac_enp0: '12.34'
};
const clientArgs = [
  'run',
  ...Object.entries(attributes).map(([key, value]) => `--inventory-attribute="${key}:${value}"`),
  `--device-type=${deviceType}`,
  `--artifact-name=${artifactName}`,
  `--auth-interval=${updateInterval}`,
  `--inventory-interval=${updateInterval}`,
  `--update-interval=${updateInterval}`
];
export const startClient = async (baseUrl, token, count) => {
  const srippedBaseUrl = baseUrl.replace(/\/$/, '');
  const args = [...clientArgs, `--count=${count}`, `--server-url=${srippedBaseUrl}`];
  if (token) {
    args.push(`--tenant-token=${token}`);
  }
  console.log(`starting using: ./mender-stress-test-client ${args.join(' ')}`);
  const child = spawn('./mender-stress-test-client', args);
  child.on('error', err => console.error(`${err}`));
  child.on('message', err => console.error(`${err}`));
  child.on('spawn', err => console.error(`${err}`));
  // child.stdout.on('data', data => {
  //   console.log(`stdout mstc: ${data}`);
  // });
  child.stderr.on('data', data => {
    console.error(`stderr mstc: ${data}`);
  });
  child.on('close', code => {
    console.log(`child process exited with code ${code}`);
  });
};

export const startDockerClient = async (baseUrl, token) => {
  const projectRoot = process.cwd();
  const srippedBaseUrl = baseUrl.replace(/\/$/, '');
  updateConfigFileWithUrl('mender', srippedBaseUrl, token);
  updateConfigFileWithUrl('mender-connect', srippedBaseUrl, token);
  // NB! to run the tests against a running local Mender backend, uncomment & adjust the following
  // const localNetwork = ['--network', 'menderintegration_mender'];
  const localNetwork = baseUrl.includes('docker.mender.io') ? ['--network', 'gui-tests_mender'] : [];
  const args = [
    'run',
    '-d',
    '--name',
    'connect-client',
    ...localNetwork,
    '-v',
    `${projectRoot}/dockerClient/mender-test.json:/etc/mender/mender.conf`,
    '-v',
    `${projectRoot}/dockerClient/mender-connect-test.json:/etc/mender/mender-connect.conf`,
    'mendersoftware/mender-client-docker-addons:mender-3.6.x'
  ];
  console.log(`starting with token: ${token}`);
  console.log(`starting using: docker ${args.join(' ')}`);
  const child = spawn('docker', args);
  child.on('error', err => console.error(`${err}`));
  child.on('message', err => console.error(`${err}`));
  child.on('spawn', err => console.error(`${err}`));
  // child.stdout.on('data', data => {
  //   console.log(`stdout docker: ${data}`);
  // });
  child.stderr.on('data', data => {
    console.error(`stderr docker: ${data}`);
  });
  child.on('close', code => {
    console.log(`child process exited with code ${code}`);
  });
  return Promise.resolve();
};

export const stopDockerClient = async () => {
  console.log('stopping: docker');
  const child = spawn('docker stop connect-client && docker rm connect-client', {
    shell: true
  });
  child.on('error', err => console.error(`${err}`));
  child.on('message', err => console.error(`${err}`));
  child.on('spawn', err => console.error(`${err}`));
  child.stderr.on('data', data => console.error(`stderr docker: ${data}`));
  child.on('close', code => console.log(`child process exited with code ${code}`));
  return Promise.resolve();
};

export const login = async (username: string, password: string, baseUrl: string) => {
  const request = await axios({
    url: `${baseUrl}api/management/v1/useradm/auth/login`,
    method: 'POST',
    headers: {
      Authorization: `Basic ${Buffer.from(`${username}:${password}`).toString('base64')}`,
      'Content-Type': 'application/json'
    },
    httpsAgent: new https.Agent({
      rejectUnauthorized: false
    })
  });

  if (request.status !== 200) {
    throw 'oh no';
  }

  const token = request.data;
  const userId = jwtDecode(token).sub;
  return { token, userId };
};

export const isLoggedIn = async (page: Page, timeout: number = 0) => {
  const cookieConsentButton = await page.getByText(/decline/i);
  if (await cookieConsentButton?.isVisible()) {
    await cookieConsentButton.click();
    await page.keyboard.press('Escape');
  }
  return page.getByRole('link', { name: selectors.loggedInText }).waitFor({ timeout });
};

export const isEnterpriseOrStaging = environment => ['enterprise', 'staging'].includes(environment);

export const processLoginForm = async ({ username, password, environment, page, stayLoggedIn = false }) => {
  await page.click(selectors.email);
  await page.fill(selectors.email, username);

  if (isEnterpriseOrStaging(environment)) {
    // enterprise supports two-step login, and the first screen does not have password field until submit clicked
    await page.getByRole('button', { name: /log in/i }).click();
  }

  await page.click(selectors.password);
  await page.fill(selectors.password, password);

  if (stayLoggedIn) {
    const checkbox = await page.getByLabel(/stay logged in/i);
    await checkbox.check();
  }

  await page.getByRole('button', { name: /log in/i }).click();
};

export const tenantTokenRetrieval = async (baseUrl: string, page: Page) => {
  await page.goto(`${baseUrl}ui/settings/organization-and-billing`);
  await page.waitForSelector('.tenant-token-text');
  return page.$eval('.tenant-token-text', el => el.textContent);
};

let previousSecret;
export const generateOtp = async (otpSecret?) => {
  let filesecret;
  try {
    filesecret = fs.readFileSync('secret.txt', 'utf8');
    console.log(filesecret);
  } catch {
    console.log('no secret.txt found - moving on...');
  }
  previousSecret = otpSecret ?? previousSecret ?? filesecret;
  const secret = previousSecret;
  if (!secret) {
    throw new Error('No secret has been provided.');
  }
  fs.writeFileSync('secret.txt', secret);
  console.log(`2fa secret: ${secret}`);
  return authenticator.generate(secret);
};

const protocol = 'https://';
export const baseUrlToDomain = (baseUrl: string) => baseUrl.substring(baseUrl.indexOf(protocol) + protocol.length, baseUrl.length - 1);

export const compareImages = (expectedPath, actualPath, options = { threshold: 0.1, usePercentage: true }) => {
  const { threshold, usePercentage } = options;
  if (!fs.existsSync(expectedPath)) {
    fs.copyFileSync(actualPath, expectedPath);
  }
  const img1 = PNG.sync.read(fs.readFileSync(actualPath));
  const img2 = PNG.sync.read(fs.readFileSync(expectedPath));
  const { width, height } = img1;
  const diff = new PNG({ width, height });
  const numDiffPixels = pixelmatch(img1.data, img2.data, diff.data, width, height, options);
  const diffPath = path.join(__dirname, '..', 'test-results', 'diffs');
  if (!fs.existsSync(diffPath)) {
    fs.mkdirSync(diffPath);
  }
  fs.writeFileSync(path.join(diffPath, `diff-${Date.now()}.png`), PNG.sync.write(diff));
  const pass = usePercentage ? (numDiffPixels / (width * height)) * 100 < threshold : numDiffPixels < threshold;
  return { pass, numDiffPixels };
};

export const tagRelease = async (releaseName: string, tag: string, baseUrl: string, token: string) => {
  const request = await axios.put(`${baseUrl}api/management/v2/deployments/deployments/releases/${releaseName}/tags`, [tag], {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    httpsAgent: new https.Agent({ rejectUnauthorized: false })
  });

  if (request.status >= 300) {
    console.error(`failed to tag release ${releaseName} got status:`, request.status);
    throw 'oh no';
  }
  return Promise.resolve();
};

export const startIdpServer = ({ acsUrl = 'https://example.com/acs', metadataLocation = 'https://example.com/metadata', ...options }, callback) =>
  runServer(
    {
      ...options,
      cert: path.join(__dirname, '..', 'fixtures', 'idp-public-cert.pem'),
      key: path.join(__dirname, '..', 'fixtures', 'idp-private-key.pem'),
      acsUrl,
      audience: metadataLocation,
      signResponse: false
    },
    callback
  );
