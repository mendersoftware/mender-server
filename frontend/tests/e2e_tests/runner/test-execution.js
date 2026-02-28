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
import chalk from 'chalk';
import { join } from 'path';

import { environments, testSuiteVariants } from './cli.js';
import { composeDown, composeExec, composeRun, composeUp, removeOldClient, runCommand, withSpinner } from './compose.js';
import { exportToProcessEnv } from './config.js';

const isDuplicateError = error => {
  const message = typeof error === 'string' ? error : JSON.stringify(error);
  return /duplicate|conflict|already exists/i.test(message);
};

const createTenant = async (credentials, config, addons = [], options = '') => {
  const { name, password, username } = credentials;
  try {
    const tenantResult = await composeExec(
      'tenantadm',
      `tenantadm create-org --name=${name} --username=${username} --password=${password} --device-limit=-1 ${addons.map(addon => `--addon=${addon}`).join(' ')} ${options}`,
      config
    );
    return tenantResult.out.trim();
  } catch (error) {
    if (isDuplicateError(error)) {
      console.log(chalk.yellow(`Tenant "${name}" already exists, retrieving existing tenant ID...`));
      const result = await composeExec('tenantadm', `tenantadm get-tenant --username ${username}`, config);
      return JSON.parse(result.out).id;
    }
    throw error;
  }
};

const setupTenantToken = async (tenantId, config) =>
  await withSpinner(
    '🔑 Retrieving tenant token...',
    async () => {
      const tenantTokenResult = await composeExec('tenantadm', `tenantadm get-tenant --id ${tenantId}`, config);
      config.tenantToken = JSON.parse(tenantTokenResult.out).tenant_token;
      console.log(chalk.gray(`👤 Configured tenant token for client to pick up: ${config.tenantToken}`));
    },
    'Tenant token retrieved',
    'Failed to setup tenant token'
  );

const createServiceProviderTenant = async (credentials, config) =>
  await withSpinner(
    '📝 Creating SP tenant...',
    async () => {
      const spTenantId = await createTenant({ ...credentials, name: 'secondary', username: credentials.spTenant }, config, [], '--device-limit 100');
      // updateOne with $set is already idempotent
      await composeExec('mongo', 'mongosh --eval "db.getSiblingDB("tenantadm").tenants.updateOne({},{$set:{max_child_tenants:100}})"', config);
      await createTenant({ ...credentials, name: 'secondary', username: credentials.username2 }, config);
      try {
        await composeExec('tenantadm', `tenantadm update-tenant --id ${spTenantId} --service-provider`, config);
      } catch (error) {
        if (isDuplicateError(error)) {
          console.log(chalk.yellow(`SP tenant "${spTenantId}" already configured as service provider, continuing...`));
        } else {
          throw error;
        }
      }
    },
    'SP tenant created & configured',
    'Failed to create SP tenant'
  );

const setupEnterprise = async config =>
  await withSpinner(
    '🏢 Setting up Enterprise Environment...',
    async () => {
      const { credentials } = config;

      await removeOldClient(config);
      const tenantId = await createTenant({ ...credentials, name: 'test' }, config, ['configure', 'monitor', 'troubleshoot']);
      await setupTenantToken(tenantId, config);
      await composeRun('client', [], config, { commandOptions: ['-d'] });
      await createServiceProviderTenant(credentials, config);
    },
    'Enterprise setup completed',
    'Enterprise setup failed'
  );

const setupOS = async config =>
  await withSpinner(
    '🚀 Setting up OS Environment...',
    async () => {
      const { credentials } = config;
      try {
        await composeExec('useradm', `useradm create-user --username ${credentials.username} --password ${credentials.password}`, config);
      } catch (error) {
        if (isDuplicateError(error)) {
          console.log(chalk.yellow(`User "${credentials.username}" already exists, continuing...`));
        } else {
          throw error;
        }
      }
    },
    'OS setup completed',
    'OS setup failed'
  );

const setupQemuClient = async (config, currentProcesses) => {
  await withSpinner(
    '🔎 getting QEMU client address...',
    async () => {
      const clientContainers = await runCommand('docker', ['ps', '-q', '--filter', 'label=com.docker.compose.service=client'], config, {}, currentProcesses);
      const clientContainerId = clientContainers.split('\n')[0].trim();
      if (!clientContainerId) {
        throw new Error('No client container found');
      }
      config.clientIp = await runCommand(
        'docker',
        ['inspect', '--format={{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}', clientContainerId],
        config,
        {},
        currentProcesses
      );
      console.log(`Client container found: ID: ${clientContainerId}, IP: ${config.clientIp}`);
    },
    'QEMU client container found',
    'QEMU client is in hiding, good luck!'
  );
};

export const runTests = async (config, currentProcesses) => {
  console.log(chalk.blue('🧪 Running Tests\n'));

  if (config.local || config.environment === environments.staging) {
    const testScript = config.visual ? 'test-visual-new' : 'test';
    const { username, password } = config.credentials;
    console.log(`   Active credentials: ${username} / ${password}`);
    await withSpinner(
      `🏃 Executing ${chalk.cyan(testScript)} with ${chalk.cyan(config.project)}...`,
      async () =>
        await runCommand(
          'npm',
          ['run', testScript, '--', `--project=${config.project}`],
          config,
          {
            cwd: join(config.guiRepository, 'tests/e2e_tests'),
            quiet: false
          },
          currentProcesses
        ),
      'Local tests completed',
      'Local tests failed'
    );
    return;
  }
  if (!config.skipCleanup) {
    await composeDown(config);
  }
  await composeUp(config);
  if (!config.skipSetup) {
    try {
      if (config.environment === environments.enterprise) {
        await setupEnterprise(config);
      } else {
        await setupOS(config);
      }
      if (config.variant === testSuiteVariants.qemu) {
        await setupQemuClient(config, currentProcesses);
      }
    } catch (error) {
      console.error(chalk.red(`💥 ${config.environment} setup failed:`, JSON.stringify(error)));
      throw error;
    }
  }
  exportToProcessEnv(config);
  let playwrightConfig = `--project=${config.project}`;
  if (config.variant === testSuiteVariants.qemu) {
    playwrightConfig = '--config=playwright-qemu.config.ts --project=qemu-tests';
  }
  console.log(chalk.yellow(`🐳 Running tests in docker using ${chalk.cyan(config.project)}/${chalk.blue(config.variant)}...`));

  await composeExec('gui-tests-runner', `npm install`, config);
  await composeExec('gui-tests-runner', `npx playwright install ${config.project}`, config);

  await composeExec('gui-tests-runner', `npm run test -- ${playwrightConfig} `, config, {
    callback: (data, source) => (source === 'stderr' ? console.error(chalk.red(data.toString())) : console.log(chalk.white(data.toString())))
  });
};
