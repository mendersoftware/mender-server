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
import type { Page } from '@playwright/test';

import { timeouts } from './constants.ts';

export const selectReleaseByName = async (page: Page, name: string) => {
  await page.getByRole('button', { name: 'Select a release' }).click();
  await page.getByRole('textbox', { name: 'Search releases...' }).fill(name);
  await page.waitForTimeout(timeouts.default);
  return locateReleaseByName(page, name).click();
};

export const locateReleaseByName = (page: Page, name: string) => {
  const container = page.locator('#deployment-release-container');
  return container.getByText(name).first();
};

export const selectDeviceLimitInput = (page: Page, tier: string) => {
  const tierCheckbox =  page.getByLabel(`${tier} devices`);
  const deviceInputContainer = page.locator('div').filter({ has: tierCheckbox }).last();
  return deviceInputContainer.getByLabel('Device limit');
}
