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
import * as compose from 'docker-compose';
import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { parse } from 'yaml';

export type CompatibleClient = {
  connectVersion?: string;
  hostname: string;
  image: string;
  version: string;
};

const composeFile = fileURLToPath(new URL('../fixtures/docker-compose.compat.yml', import.meta.url));
const localComposeFile = fileURLToPath(new URL('../fixtures/docker-compose.compat.local.yml', import.meta.url));

type CompatClientService = { hostname: string; image: string; labels: Record<string, string> };

const { services }: { services: Record<string, CompatClientService> } = parse(readFileSync(composeFile, 'utf-8'), { merge: true });

export const compatibleClients: CompatibleClient[] = Object.values(services).map(({ hostname, image, labels }) => ({
  connectVersion: labels['io.mender.e2e.connect-version'],
  hostname,
  image: image.substring(image.indexOf('}/') + 2), // drop the registry for more readable test names
  version: labels['io.mender.e2e.client-version']
}));

// a project of its own, so the clients can be started late in the test run & removed independently of the stack they connect to
const composeOptions = [['--project-name', 'e2e-compat-clients']];

export const startCompatClients = async (baseUrl: string, tenantToken = '') => {
  const isLocal = baseUrl.includes('docker.mender.io');
  return compose.upAll({
    config: isLocal ? [composeFile, localComposeFile] : [composeFile],
    composeOptions,
    commandOptions: ['--quiet-pull', '--force-recreate', '--remove-orphans'],
    env: { ...process.env, SERVER_URL: baseUrl.replace(/\/$/, ''), TENANT_TOKEN: tenantToken }
  });
};

export const stopCompatClients = async () => compose.down({ composeOptions, commandOptions: ['--volumes', '--remove-orphans'] });
