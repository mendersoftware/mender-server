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
import { spawn } from 'child_process';
import * as compose from 'docker-compose';
import ora from 'ora';
import process from 'process';

const getComposeOptions = config => ({
  cwd: config.serverRoot,
  config: config.composeFiles,
  env: {
    ...process.env,
    TEST_ENVIRONMENT: config.environment,
    ...(config.tenantToken && { TENANT_TOKEN: config.tenantToken }),
    ...(config.clientIp && { CLIENT_IP: config.clientIp })
  }
});

export const formatErrorMessage = error => (error instanceof Error ? error.message : JSON.stringify(error));

export const withSpinner = async (message, operation, successMessage, failMessage) => {
  const spinner = ora(message).start();
  try {
    const result = await operation();
    spinner.succeed(`🟢 ${chalk.green(successMessage)}`);
    return result;
  } catch (error) {
    spinner.fail(`💥 ${chalk.red(failMessage)}: ${formatErrorMessage(error)}`);
    throw error;
  }
};

export const composeDown = async config =>
  await withSpinner(
    '🛑 Stopping and removing containers...',
    async () => await compose.down({ ...getComposeOptions(config), commandOptions: ['-v', '--remove-orphans'] }),
    'Containers stopped and removed',
    'Failed to stop containers'
  );

export const composeUp = async config =>
  await withSpinner(
    '🚀 Starting containers...',
    async () => await compose.upAll({ ...getComposeOptions(config), commandOptions: ['--quiet-pull', '--wait', '--wait-timeout', '120'] }),
    'Containers started and healthy',
    'Failed to start containers'
  );

export const composeExec = async (service, command, config, options = {}) =>
  await withSpinner(
    `⚡ Executing command in ${chalk.cyan(service)}...`,
    async () => await compose.exec(service, command, { ...getComposeOptions(config), ...options }),
    `Command executed in ${chalk.cyan(service)}`,
    `Command failed in ${chalk.cyan(service)}`
  );

export const composeRun = async (service, command, config, options = {}) =>
  await withSpinner(
    `🏃 Running ${chalk.cyan(service)}...`,
    async () => await compose.run(service, command, { ...getComposeOptions(config), ...options }),
    `${chalk.cyan(service)} completed`,
    `${chalk.cyan(service)} failed`
  );

export const composeLogs = async config =>
  await withSpinner(
    '📋 Collecting container logs...',
    async () => {
      const result = await compose.logs([], getComposeOptions(config));
      return result.out;
    },
    'Logs collected',
    'Failed to collect logs'
  );

export const removeOldClient = async config =>
  await withSpinner(
    '🗑️ Removing old client...',
    async () => await compose.rm({ ...getComposeOptions(config), commandOptions: ['-fsv', 'client'] }),
    'Old client removed',
    'Failed to remove old client'
  );

export const runCommand = (command, args = [], config, options = {}, currentProcesses = []) =>
  new Promise((resolve, reject) => {
    const { quiet = true, throwOnError = true, shell = false, ...remainderOptions } = options;
    let output = '';

    const child = spawn(command, args, {
      stdio: ['inherit', 'pipe', 'pipe'],
      shell,
      ...remainderOptions
    });
    child.stdout.on('data', data => {
      const text = data.toString();
      output += text;
      if (!quiet) {
        process.stdout.write(text);
      }
    });
    child.stderr.on('data', data => {
      const text = data.toString();
      output += text;
      if (!quiet) {
        process.stdout.write(text);
      }
    });

    const cleanup = () => {
      const index = currentProcesses.indexOf(child);
      if (index > -1) {
        currentProcesses.splice(index, 1);
      }
    };

    child.on('close', code => {
      if (!throwOnError || code === 0) {
        resolve(output.trim());
      } else {
        cleanup();
        reject(new Error(`Command failed with exit code ${code}: ${output}`));
      }
    });

    child.on('error', error => {
      cleanup();
      if (throwOnError) {
        reject(error);
      } else {
        resolve(output.trim());
      }
    });

    currentProcesses.push(child);
  });
