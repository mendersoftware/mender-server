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
import chalk from 'chalk';
import { mkdirSync, writeFileSync } from 'fs';
import { join } from 'path';
import process from 'process';

import { banner, program } from './runner/cli.js';
import { formatErrorMessage, withSpinner } from './runner/compose.js';
import { createConfig, exportToProcessEnv, showConfiguration, validateConfiguration } from './runner/config.js';
import { promptForConfiguration } from './runner/interactive.js';
import { cleanup, initiateShutDownSequence } from './runner/lifecycle.js';
import { runTests } from './runner/test-execution.js';

let config;
let currentProcesses = [];
const shutdownState = { inProgress: false };

const main = async () => {
  program.parse();
  const options = program.opts();
  if (options.banner) {
    console.log(banner);
  }
  config = createConfig(options);
  if (config.interactive) {
    config = await promptForConfiguration();
  } else if (!process.argv.slice(2).length) {
    program.help();
    return;
  }
  validateConfiguration(config);
  showConfiguration(config);
  exportToProcessEnv(config);

  await runTests(config, currentProcesses);
  console.log(chalk.green('\n🎉 All tests completed successfully!\n'));
};

process.on(
  'SIGINT',
  initiateShutDownSequence('SIGINT', () => config, currentProcesses, shutdownState)
);

process.on(
  'SIGTERM',
  initiateShutDownSequence('SIGTERM', () => config, currentProcesses, shutdownState)
);

const errorLogCutoffLength = 200;
let exitCode = 0;
try {
  await main();
} catch (error) {
  const errorMessage = formatErrorMessage(error);
  if (config?.guiRepository) {
    const logDir = join(config.guiRepository, 'logs');
    mkdirSync(logDir, { recursive: true });
    const errorPath = join(logDir, 'error.txt');
    console.error(chalk.red('\n💥 Error:', errorMessage.length < errorLogCutoffLength ? errorMessage : `will be dumped to ${errorPath}`));
    if (errorMessage.length >= errorLogCutoffLength) {
      writeFileSync(errorPath, error instanceof Error ? error.stack : errorMessage);
    }
  } else {
    console.error(chalk.red('\n💥 Error:', errorMessage));
  }
  exitCode = 1;
} finally {
  if (config) {
    await withSpinner('🧹 Cleanup', () => cleanup(config, currentProcesses), 'so sparkling clean', 'clean up failed');
  }
  if (exitCode === 0) {
    console.log(chalk.green('🚀 Test runner completed successfully!'));
  } else {
    console.log(chalk.red('💥 Test runner failed!'));
  }
  if (process.env.ALLOWED_TO_FAIL) {
    process.exit(464); // would equate to `incompatible protocol` as in test version & deployed version have diverged slightly - so failures might be tolerable
  }
  process.exit(exitCode);
}
