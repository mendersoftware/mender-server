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
import { appendFileSync, mkdirSync, writeFileSync } from 'fs';
import { join } from 'path';
import process from 'process';

import { testSuiteVariants } from './cli.js';
import { composeDown, composeLogs, runCommand } from './compose.js';

export const killTestProcesses = currentProcesses => {
  console.log(chalk.yellow('🔪 Terminating remaining processes...'));
  currentProcesses.forEach(child => {
    if (child?.killed) {
      return;
    }
    try {
      child.kill('SIGTERM');
      setTimeout(() => {
        // clean up without compromise after grace period
        if (!child.killed) {
          child.kill('SIGKILL');
        }
      }, 3000);
    } catch (error) {
      console.log(chalk.yellow(`Warning: Failed to kill process ${child.pid}: ${error.message}`));
    }
  });
  currentProcesses.length = 0;
};

export const collectClientLogs = async (logDir, config, currentProcesses) => {
  // the client gets often started outside of the compose setup, so track it down by name
  console.log(chalk.yellow(`📋 Capturing client logs to ${chalk.cyan(join(logDir, 'client.*'))}`));
  const containerNames = await runCommand('docker', ['ps', '-a', `--format={{.Names}}`], config, {}, currentProcesses);
  const clientContainer = containerNames.split('\n').find(name => name.includes('client'));
  if (!clientContainer) {
    console.log(chalk.yellow('📋 Client container not found'));
    return;
  }
  const clientLogPath = join(logDir, 'client.log');
  const fullClientLogPath = join(logDir, 'fullClient.log');
  const debugClientFilesPath = join(logDir, 'debugClient.log');

  const clientLog = await runCommand('docker', ['logs', clientContainer], config, {}, currentProcesses);
  writeFileSync(clientLogPath, clientLog);
  if (config.variant !== testSuiteVariants.qemu) {
    console.log(chalk.yellow('🟢 Docker client logs written'));
    return;
  }

  const ip = await runCommand(
    'docker',
    ['inspect', `--format={{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}`, clientContainer],
    config,
    {},
    currentProcesses
  );
  const fullClientLog = await runCommand(
    'ssh',
    ['-p', '8822', '-o', 'StrictHostKeyChecking=no', `root@${ip}`, 'journalctl', '--no-pager', '--all'],
    config,
    {},
    currentProcesses
  );
  writeFileSync(fullClientLogPath, fullClientLog);

  const clientConf = await runCommand(
    'ssh',
    ['-p', '8822', '-o', 'StrictHostKeyChecking=no', `root@${ip}`, 'cat', '/etc/mender/mender.conf'],
    config,
    {},
    currentProcesses
  );
  writeFileSync(debugClientFilesPath, 'Mender configuration:');
  appendFileSync(debugClientFilesPath, clientConf);
  const deploymentsLogs = await runCommand(
    'ssh',
    ['-p', '8822', '-o', 'StrictHostKeyChecking=no', `root@${ip}`, 'cat', '/data/mender/deployment*.log'],
    config,
    { throwOnError: false, shell: true },
    currentProcesses
  );
  appendFileSync(debugClientFilesPath, 'Deployment logs:');
  appendFileSync(debugClientFilesPath, deploymentsLogs);
};

export const cleanup = async (config, currentProcesses) => {
  killTestProcesses(currentProcesses);
  if (config.local) {
    return;
  }
  const logDir = join(config.guiRepository, 'logs');
  const logPath = join(logDir, 'gui_e2e_tests.txt');

  try {
    mkdirSync(logDir, { recursive: true });
    await collectClientLogs(logDir, config, currentProcesses);
    console.log(chalk.yellow(`📋 Dumping logs to ${chalk.cyan(logPath)}`));
    const logs = await composeLogs(config);
    writeFileSync(logPath, logs);
  } catch (error) {
    console.error(chalk.red('💥 Failed to dump logs:'), error);
  }

  if (config.skipCleanup) {
    console.log(chalk.yellow('⚠️ Containers left running'));
    return;
  }
  await composeDown(config);
};

export const initiateShutDownSequence = (signal, getConfig, currentProcesses, shutdownState) => async () => {
  if (shutdownState.inProgress) {
    console.log(chalk.yellow(`\n⚡ Shutdown already in progress...`));
    return;
  }
  shutdownState.inProgress = true;
  console.log(chalk.yellow(`\n⚡ Received ${signal}, cleaning up...`));
  const config = getConfig();
  if (config) {
    await cleanup(config, currentProcesses);
  } else {
    killTestProcesses(currentProcesses);
  }
  process.exit(1);
};
