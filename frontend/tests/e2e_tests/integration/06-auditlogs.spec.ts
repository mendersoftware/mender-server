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
import * as fs from 'fs';
import * as path from 'path';
import * as readline from 'readline';

import test, { expect } from '../fixtures/fixtures';
import { compareImages, isEnterpriseOrStaging } from '../utils/commands';
import { selectors, storagePath, timeouts } from '../utils/constants';

test.describe('Auditlogs', () => {
  test.use({ storageState: storagePath });

  const secret = 'super secret something text';
  test('will track remote terminal sessions', async ({ environment, loggedInPage: page }) => {
    test.skip(!isEnterpriseOrStaging(environment));
    await page.click(`.leftNav :text('Devices')`);
    await page.click(`${selectors.deviceListItem} div:last-child`);
    await page.click(`text=/troubleshooting/i`);
    // the deviceconnect connection might not be established right away
    await page.waitForSelector('text=/Session status/i', { timeout: timeouts.tenSeconds });
    const connectionButton = await page.getByRole('button', { name: /connect/i });
    await connectionButton.first().click();
    await page.waitForSelector(`text=Connection with the device established`, { timeout: timeouts.tenSeconds });
    expect(await page.isVisible('.terminal.xterm .xterm-screen')).toBeTruthy();

    // the terminal content might take a bit to get painted - thus the waiting
    await page.click(selectors.terminalElement, { timeout: timeouts.default });

    const terminalText = await page.locator(selectors.terminalText);
    await terminalText.fill('passwd');
    await page.keyboard.press('Enter');
    const expectedPath = path.join(__dirname, '..', 'test-results', 'diffs', 'terminalSecretContent.png');
    const elementHandle = await page.$(selectors.terminalElement);
    await elementHandle.screenshot({ path: expectedPath });
    await page.waitForTimeout(timeouts.oneSecond);
    await terminalText.pressSequentially(secret);

    const screenShotPath = path.join(__dirname, '..', 'test-results', 'diffs', 'terminalSecretContent-actual.png');
    await elementHandle.screenshot({ path: screenShotPath });
    const { pass } = compareImages(expectedPath, screenShotPath);
    expect(pass).toBeTruthy();
    await page.click('[aria-label="close"]'); // short-form
    await page.click(`.leftNav.navLink:has-text('Audit log')`);

    await page.click(`.auditlogs-list-item :text('CLOSE_TERMINAL')`);
    await page.click(`button:has-text('Play')`);
    expect(await page.isVisible(`.MuiDrawer-paper a:has-text('Download'), .MuiDrawer-paper button:has-text('Download')`)).toBeTruthy();
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      page.click(`.MuiDrawer-paper a:has-text('Download'), .MuiDrawer-paper button:has-text('Download')`)
    ]);
    const downloadTargetPath = await download.path();
    expect(downloadTargetPath).toBeTruthy();

    const checkDownloadedReplayForSecret = async (path, secret) => {
      const fileStream = fs.createReadStream(path);
      const lines = readline.createInterface({
        input: fileStream,
        crlfDelay: Infinity
      });
      for await (const line of lines) {
        const search = `const transfer = '[{"content":`;
        if (line.includes(search)) {
          const encodedText = line.substring(line.indexOf(search) + search.length, line.lastIndexOf(`}]';`));
          const content = JSON.parse(Buffer.from(encodedText, 'base64').toString());
          const decodedContent = String.fromCharCode(...content.data);
          expect(decodedContent).not.toContain(secret);
          fileStream.close();
          return;
        }
      }
    };
    await checkDownloadedReplayForSecret(downloadTargetPath, 'secret something');
  });
});
