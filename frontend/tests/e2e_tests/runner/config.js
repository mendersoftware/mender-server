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
import { execSync } from 'child_process';
import { existsSync } from 'fs';
import { join } from 'path';
import process from 'process';
import { v4 as uuid } from 'uuid';

import { defaults, environments, projects, testSuiteVariants } from './cli.js';

const serverRoot = process.env.SERVER_ROOT || execSync('git rev-parse --show-toplevel', { encoding: 'utf-8' }).trim();
const guiRepository = process.env.GUI_REPOSITORY || join(serverRoot, 'frontend');

const defaultCredentials = {
  username: 'mender-demo@example.com',
  username2: 'demo-secondary@example.com',
  spTenant: 'tenant-demo@example.com',
  password: 'mysecretpassword!123'
};

const getCredentials = config => {
  const credentials = {
    ...defaultCredentials,
    password: config.password ?? defaultCredentials.password,
    username: config.username ?? defaultCredentials.username
  };
  if (config.environment === environments.staging) {
    credentials.username = config.username ?? `${uuid()}@example.com`;
    credentials.password = config.password ?? uuid();
  }
  return credentials;
};

export const createConfig = (options = {}) => {
  const visual = options.visual || options.executionMode === 'visual' || false;
  const hasBaseUrl = !!(options.baseUrl || process.env.BASE_URL);
  const local = options.local || visual || hasBaseUrl || options.executionMode === 'local' || false;

  const baseConfig = {
    baseUrl: options.baseUrl || process.env.BASE_URL,
    environment: options.environment || process.env.TEST_ENVIRONMENT || defaults.environment,
    guiRepository,
    interactive: options.interactive || false,
    local,
    visual,
    project: options.project || defaults.project,
    serverRoot,
    skipCleanup: options.skipCleanup || false,
    skipSetup: hasBaseUrl,
    username: options.username || process.env.STAGING_USER,
    password: options.password || process.env.STAGING_PASSWORD,
    variant: options.variant,
    tenantToken: null,
    clientIp: null
  };

  const composeFiles = [join(serverRoot, 'docker-compose.yml'), join(guiRepository, 'tests/e2e_tests/docker-compose.e2e-tests.yml')];

  if (options.file || options.customComposeFile) {
    const filesToAdd = options.file || (options.customComposeFile ? [options.customComposeFile] : []);
    composeFiles.push(...filesToAdd);
  }

  if (baseConfig.environment === environments.enterprise) {
    composeFiles.push(join(baseConfig.serverRoot, 'compose/docker-compose.enterprise.yml'));
    composeFiles.push(join(baseConfig.guiRepository, 'tests/e2e_tests/docker-compose.e2e-tests.enterprise.yml'));
    composeFiles.push(join(baseConfig.serverRoot, 'compose/docker-compose.smtp4dev.yml'));
  }
  if (baseConfig.variant === testSuiteVariants.qemu) {
    composeFiles.push(join(baseConfig.guiRepository, 'tests/e2e_tests/docker-compose.e2e-tests.rofs.yml'));
  }

  const configWithFiles = { ...baseConfig, composeFiles };
  configWithFiles.credentials = getCredentials(configWithFiles);
  return configWithFiles;
};

const configToEnvMap = {
  baseUrl: 'BASE_URL',
  clientIp: 'CLIENT_IP',
  environment: 'TEST_ENVIRONMENT',
  guiRepository: 'GUI_REPOSITORY',
  serverRoot: 'SERVER_ROOT',
  tenantToken: 'TENANT_TOKEN'
};

export const exportToProcessEnv = config => {
  Object.entries(configToEnvMap).forEach(([key, envVar]) => {
    if (config[key]) {
      process.env[envVar] = config[key];
    }
  });
  if (config.credentials?.username) {
    process.env.STAGING_USER = config.credentials.username;
  }
  if (config.credentials?.password) {
    process.env.STAGING_PASSWORD = config.credentials.password;
  }
};

export const validateConfiguration = config => {
  const errors = [];

  if (!projects[config.project]) {
    errors.push(`Invalid project: ${config.project}. Valid projects are: ${Object.values(projects).join(', ')}`);
  }
  if (!environments[config.environment]) {
    errors.push(`Invalid environment: ${config.environment}. Valid environments are: ${Object.values(environments).join(', ')}`);
  }
  if (config.variant === testSuiteVariants.qemu && config.environment !== environments.enterprise) {
    errors.push(`--variant qemu requires --environment enterprise`);
  }
  const missingFiles = config.composeFiles.filter(file => !existsSync(file));
  if (missingFiles.length > 0) {
    errors.push(`Missing compose files: ${missingFiles.join(', ')}`);
  }
  if (config.skipCleanup && config.local) {
    console.log(chalk.gray('   Note: --skip-cleanup has no effect in local mode'));
  }
  if (!errors.length) {
    return;
  }
  console.error(chalk.red('💥 Configuration Validation Failed:'));
  errors.forEach(error => console.error(chalk.red(`   • ${error}`)));
  process.exit(1);
};

export const showConfiguration = config => {
  console.log(chalk.green('\n📋 Configuration Summary:'));
  console.log(`   Environment: ${config.environment}`);
  console.log(`   Variant: ${config.variant}`);
  console.log(`   Execution: ${config.local ? (config.visual ? chalk.magenta('Local Visual') : chalk.cyan('Local')) : chalk.blue('Docker')}`);
  console.log(`   Browser: ${chalk.cyan(config.project)}`);
  console.log(`   Cleanup: ${config.skipCleanup ? chalk.red('Skip') : chalk.green('Auto')}`);
  console.log(`   Compose Files: ${config.composeFiles}`);
  if (config.username && config.password) {
    console.log(`   Credentials: ${config.username} / ${config.password}`);
  }
  if (config.baseUrl) {
    console.log(`   Running against: ${config.baseUrl}`);
  }
  console.log('');
};
