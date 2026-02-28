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
import { Command } from 'commander';

export const banner = chalk.cyan(`
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║                      🧪 Mender E2E Tests                      ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
`);

export const projects = {
  chromium: 'chromium',
  firefox: 'firefox',
  webkit: 'webkit'
};

export const environments = {
  enterprise: 'enterprise',
  os: 'os',
  staging: 'staging'
};

export const testSuiteVariants = {
  regular: 'regular',
  qemu: 'qemu'
};

export const defaults = {
  project: projects.chromium,
  environment: environments.os,
  variant: testSuiteVariants.regular
};

export const program = new Command();
program
  .name('mender-e2e-runner')
  .version('2.0.0')
  .option('-c, --skip-cleanup', 'Leave containers running after tests')
  .option('--local', 'Execute tests on your local machine')
  .option('--visual', 'Run with Playwright UI (implies --local)')
  .option('-f, --file <path>', 'Specify custom compose file', (value, previous) => (previous ? [...previous, value] : [value]))
  .option('-e --environment <env>', `Specify environment to use (${Object.values(environments).join(', ')})`, defaults.environment)
  .option('-p, --project <browser>', `Browser project to run (${Object.values(projects).join(', ')})`, defaults.project)
  .option('-i, --interactive', 'Run in interactive mode with prompts')
  .option('--username <email>', 'User email to use')
  .option('--password <password>', 'User password to use')
  .option('--variant <variant>', `Special test variant to be run (one of ${Object.keys(testSuiteVariants)})`, defaults.variant)
  .option('--base-url <url>', 'Location to run tests against')
  .option('--no-banner', 'Skip the banner display')
  .addHelpText(
    'after',
    `\n
${chalk.yellow('Examples:')}
  ${chalk.green('$ node run.js --interactive')}                        Run with interactive prompts
  ${chalk.green('$ node run.js --local --skip-cleanup')}               Run locally and keep containers
  ${chalk.green('$ node run.js --environment enterprise')}             Run with enterprise setup
  ${chalk.green('$ node run.js --environment staging')}                Run with staging environment
  ${chalk.green('$ node run.js --visual')}                               Run with Playwright UI (implies --local)
  ${chalk.green('$ node run.js --project firefox')}                    Run tests with Firefox browser
  ${chalk.green('$ node run.js --project webkit --local')}             Run WebKit tests locally
  ${chalk.green('$ node run.js --base-url https://foo.example.com')}   Run tests against https://foo.example.com

${chalk.yellow('Browser Projects:')}
  ${chalk.cyan(projects.chromium)}     Chrome (default)
  ${chalk.cyan(projects.firefox)}      Firefox
  ${chalk.cyan(projects.webkit)}       Safari/WebKit

${chalk.yellow('Environments:')}
  ${chalk.cyan(environments.os)}           Open Source (default)
  ${chalk.cyan(environments.enterprise)}   Enterprise with SP tenant tests
  ${chalk.cyan(environments.staging)}      Staging

${chalk.yellow('Environment Variables:')}
  ${chalk.cyan('SERVER_ROOT')}        Root directory of the server repository
  ${chalk.cyan('GUI_REPOSITORY')}     Path to the GUI repository
  ${chalk.cyan('TEST_ENVIRONMENT')}   Test environment (${Object.values(environments).join('|')})
  ${chalk.cyan('STAGING_USER')}       User email
  ${chalk.cyan('STAGING_PASSWORD')}   User password
  ${chalk.cyan('BASE_URL')}           Location to run tests against
`
  );
