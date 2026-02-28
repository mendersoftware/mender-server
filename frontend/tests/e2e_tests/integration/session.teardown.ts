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
import { test as base } from '@playwright/test';
import * as fs from 'fs';

import { stopDockerClient } from '../utils/commands';
import { storageFolder } from '../utils/constants';

const test = base;

test('teardown', async () => {
  try {
    await stopDockerClient();
  } catch {
    // best-effort: client may not be running
  }

  try {
    if (fs.existsSync(storageFolder)) {
      fs.rmSync(storageFolder, { recursive: true, force: true });
    }
  } catch {
    // best-effort
  }

  for (const file of ['loginInfo.json', 'secret.txt']) {
    try {
      if (fs.existsSync(file)) {
        fs.unlinkSync(file);
      }
    } catch {
      // best-effort
    }
  }
});
