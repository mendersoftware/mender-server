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
import { execSync, spawn } from 'child_process';
import { Command } from 'commander';
import * as compose from 'docker-compose';
import { existsSync, mkdirSync } from 'fs';
import { writeFileSync } from 'fs';
import inquirer from 'inquirer';
import ora from 'ora';
import { join } from 'path';
import process from 'process';
import { v4 as uuid } from 'uuid';
import validator from 'validator';

const banner = chalk.cyan(`
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║                      🧪 Mender E2E Tests                      ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
`);

const serverRoot = process.env.SERVER_ROOT || execSync('git rev-parse --show-toplevel', { encoding: 'utf-8' }).trim();
const guiRepository = process.env.GUI_REPOSITORY || join(serverRoot, 'frontend');

const projects = {
  chromium: 'chromium',
  firefox: 'firefox',
  webkit: 'webkit'
};

const environments = {
  enterprise: 'enterprise',
  os: 'os',
  staging: 'staging'
};

const testSuiteVariants = {
  regular: 'regular',
  qemu: 'qemu'
};

const defaults = {
  project: projects.chromium,
  environment: environments.os,
  variant: testSuiteVariants.regular
};

const defaultCredentials = {
  user: 'mender-demo@example.com',
  user2: 'demo-secondary@example.com',
  spTenant: 'tenant-demo@example.com',
  password: 'mysecretpassword!123'
};

const getCredentials = config => {
  const credentials = {
    ...defaultCredentials,
    password: config.password ?? defaultCredentials.password,
    user: config.user ?? defaultCredentials.user
  };
  if (config.environment === environments.staging) {
    credentials.user = config.user ?? `${uuid()}@example.com`;
    credentials.password = config.password ?? uuid();
  }
  return credentials;
};

/**  */

const createConfig = (options = {}) => {
  const baseConfig = {
    baseUrl: process.env.BASE_URL || options.baseUrl,
    environment: process.env.TEST_ENVIRONMENT || options.environment || defaults.environment,
    guiRepository,
    interactive: options.interactive || false,
    local: options.local || options.localVisual || options.executionMode === 'local' || options.executionMode === 'visual' || false,
    visual: options.localVisual || options.executionMode === 'visual' || false,
    project: options.project || defaults.project,
    serverRoot,
    skipCleanup: options.skipCleanup || false,
    user: options.user,
    password: options.password,
    variant: options.variant
  };

  const composeFiles = [join(serverRoot, 'docker-compose.yml'), join(guiRepository, 'tests/e2e_tests/docker-compose.e2e-tests.yml')];

  if (options.file || options.customComposeFile) {
    const filesToAdd = options.file || (options.customComposeFile ? [options.customComposeFile] : []);
    composeFiles.push(...filesToAdd);
  }

  if (baseConfig.environment === environments.enterprise) {
    composeFiles.push(join(baseConfig.serverRoot, 'compose/docker-compose.enterprise.yml'));
    composeFiles.push(join(baseConfig.guiRepository, 'tests/e2e_tests/docker-compose.e2e-tests.enterprise.yml'));
  }
  if (baseConfig.variant === testSuiteVariants.qemu) {
    composeFiles.push(join(baseConfig.guiRepository, 'tests/e2e_tests/docker-compose.e2e-tests.rofs.yml'));
  }

  return { ...baseConfig, composeFiles };
};

const promptForConfiguration = async () => {
  console.log(chalk.blue('\n📋 Interactive Configuration\n'));

  const answers = await inquirer.prompt([
    {
      type: 'list',
      name: 'environment',
      message: 'Which environment would you like to test?',
      choices: [
        { name: '🏢 Enterprise', value: environments.enterprise },
        { name: '🚀 Open Source', value: environments.os },
        { name: '🌐 Staging', value: environments.staging }
      ],
      default: defaults.environment
    },
    {
      type: 'list',
      name: 'executionMode',
      message: 'How would you like to run the tests?',
      choices: [
        { name: '🐳 Docker (recommended)', value: 'docker' },
        { name: '💻 Local (against a running setup)', value: 'local' },
        { name: '👀 Local with Visual Mode', value: 'visual' }
      ],
      default: 'docker'
    },
    {
      type: 'list',
      name: 'project',
      message: 'Which browser should we test with?',
      choices: [
        { name: '🌐 Chromium (Google Chrome/Chromium)', value: projects.chromium },
        { name: '🦊 Firefox (Mozilla Firefox)', value: projects.firefox },
        { name: '🧭 WebKit (Safari/WebKit)', value: projects.webkit }
      ],
      default: defaults.project
    },
    {
      type: 'confirm',
      name: 'skipCleanup',
      message: 'Keep containers running after tests?',
      default: false
    },
    {
      type: 'input',
      name: 'customComposeFile',
      message: 'Custom docker-compose file path (optional):',
      validate: input => {
        if (!input) return true;
        return existsSync(input) || 'File does not exist';
      }
    },
    {
      type: 'input',
      name: 'baseUrl',
      message: 'What location should the tests run against?',
      validate: input => {
        if (!input) return true;
        return validator.isURL(input) || 'Not a valid URL';
      }
    },
    {
      type: 'variant',
      name: 'variant',
      message: 'Special variant to be run',
      choices: [
        { name: 'Regular test sets', value: testSuiteVariants.regular },
        { name: 'QEMU dependent tests', value: testSuiteVariants.qemu }
      ],
      default: testSuiteVariants.regular
    }
  ]);

  config = createConfig(answers);
};

const getComposeOptions = config => ({
  cwd: config.serverRoot,
  config: config.composeFiles,
  env: {
    ...process.env,
    TEST_ENVIRONMENT: config.environment,
    ...(process.env.TENANT_TOKEN && { TENANT_TOKEN: process.env.TENANT_TOKEN })
  }
});

const withSpinner = async (message, operation, successMessage, failMessage) => {
  const spinner = ora(message).start();
  try {
    const result = await operation();
    spinner.succeed(`🟢 ${chalk.green(successMessage)}`);
    return result;
  } catch (error) {
    spinner.fail(`💥 ${chalk.red(failMessage)}`, JSON.stringify(error));
    throw error;
  }
};

const composeDown = async config =>
  await withSpinner(
    '🛑 Stopping and removing containers...',
    async () => await compose.down({ ...getComposeOptions(config), commandOptions: ['-v', '--remove-orphans'] }),
    'Containers stopped and removed',
    'Failed to stop containers'
  );

const composeUp = async config => {
  await withSpinner(
    '🚀 Starting containers...',
    async () => await compose.upAll({ ...getComposeOptions(config), commandOptions: ['--quiet-pull'] }),
    'Containers started successfully',
    'Failed to start containers'
  );

  await withSpinner(
    '⏳ Waiting for services to be ready...',
    async () => await new Promise(resolve => setTimeout(resolve, 5000)),
    'Services are ready',
    'Services failed to become ready'
  );
};

const composeExec = async (service, command, config, options = {}) =>
  await withSpinner(
    `⚡ Executing command in ${chalk.cyan(service)}...`,
    async () => await compose.exec(service, command, { ...getComposeOptions(config), ...options }),
    `Command executed in ${chalk.cyan(service)}`,
    `Command failed in ${chalk.cyan(service)}`
  );

const composeRun = async (service, command, config, options = {}) =>
  await withSpinner(
    `🏃 Running ${chalk.cyan(service)}...`,
    async () => await compose.run(service, command, { ...getComposeOptions(config), ...options }),
    `${chalk.cyan(service)} completed`,
    `${chalk.cyan(service)} failed`
  );

const composeLogs = async config =>
  await withSpinner(
    '📋 Collecting container logs...',
    async () => {
      console.log('getting em logs');
      const result = await compose.logs(getComposeOptions(config));
      console.log('gotten logs', result);
      return result.out;
    },
    'Logs collected',
    'Failed to collect logs'
  );

const runCommand = (command, args = [], config, options = {}) =>
  new Promise((resolve, reject) => {
    let stdout = '';
    let stderr = '';
    const child = spawn(command, args, { stdio: ['inherit', 'pipe', 'pipe'], env: { ...process.env, TEST_ENVIRONMENT: config.environment }, ...options });
    child.stdout.on('data', data => (stdout += data.toString()));
    child.stderr.on('data', data => (stderr += data.toString()));

    const cleanup = () => {
      const index = currentProcesses.indexOf(child);
      if (index > -1) {
        currentProcesses.splice(index, 1);
      }
    };

    child.on('close', code => {
      if (code === 0) {
        resolve(stdout.trim());
      } else {
        cleanup();
        reject(new Error(`Command failed with exit code ${code}: ${stderr}`));
      }
    });

    child.on('error', error => {
      cleanup();
      reject(error);
    });

    currentProcesses.push(child);
  });

const removeOldClient = async config =>
  await withSpinner(
    '🗑️ Removing old client...',
    async () => await compose.rm({ ...getComposeOptions(config), commandOptions: ['-fsv', 'client'] }),
    'Old client removed',
    'Failed to remove old client'
  );

const createTenant = async (credentials, config, addons = [], options = '') => {
  const { name, password, user } = credentials;
  const tenantResult = await composeExec(
    'tenantadm',
    `tenantadm create-org --name=${name} --username=${user} --password=${password} ${addons.map(addon => `--addon=${addon}`).join(' ')} ${options}`,
    config
  );
  return tenantResult.out.trim();
};

const setupTenantToken = async (tenantId, config) =>
  await withSpinner(
    '🔑 Retrieving tenant token...',
    async () => {
      const tenantTokenResult = await composeExec('tenantadm', `tenantadm get-tenant --id ${tenantId}`, config);
      process.env.TENANT_TOKEN = JSON.parse(tenantTokenResult.out).tenant_token;
    },
    'Tenant token retrieved',
    'Failed to setup tenant token'
  );

const createServiceProviderTenant = async (credentials, config) =>
  await withSpinner(
    '📝 Creating SP tenant...',
    async () => {
      const spTenantId = await createTenant({ ...credentials, name: 'secondary', user: credentials.spTenant }, config, [], '--device-limit 100');
      await composeExec('mongo', 'mongosh --eval "db.getSiblingDB("tenantadm").tenants.updateOne({},{$set:{max_child_tenants:100}})"', config);
      await createTenant({ ...credentials, name: 'secondary', user: credentials.user2 }, config);
      await composeExec('tenantadm', `tenantadm update-tenant --id ${spTenantId} --service-provider`, config);
    },
    'SP tenant created & configured',
    'Failed to create SP tenant'
  );

const setupEnterprise = async config => {
  console.log(chalk.blue('\n🏢 Setting up Enterprise Environment\n'));
  console.log(chalk.yellow('👤 Creating tenants and users...'));

  const credentials = getCredentials(config);

  await removeOldClient(config);
  const tenantId = await createTenant({ ...credentials, name: 'test' }, config, ['configure', 'monitor', 'troubleshoot']);
  await setupTenantToken(tenantId, config);
  await composeRun('client', [], config, { commandOptions: ['-d'] });
  await createServiceProviderTenant(credentials, config);

  console.log(chalk.green('🎉 Enterprise setup completed successfully!\n'));
};

const setupOS = async config => {
  console.log(chalk.blue('\n🚀 Setting up OS Environment\n'));
  console.log(chalk.yellow('👤 Creating default user...'));
  const credentials = getCredentials(config);
  await composeExec('useradm', `useradm create-user --username ${credentials.user} --password ${credentials.password}`, config);
  console.log(chalk.green('🎉 OS setup completed successfully!\n'));
};

const setupQemuClient = async config => {
  let clientIp = '';
  await withSpinner(
    `🔎 getting QEMU client address...`,
    async () => {
      const clientContainers = await runCommand('docker', ['ps', '-q', '--filter', 'label=com.docker.compose.service=client'], config);
      const clientContainerId = clientContainers.split('\n')[0].trim();
      if (!clientContainerId) {
        throw new Error('No client container found');
      }
      clientIp = await runCommand('docker', ['inspect', '--format={{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}', clientContainerId], config);
      console.log(`Client container found: ID: ${clientContainerId}, IP: ${clientIp}`);
    },
    'QEMU client container found',
    'QEMU client is in hiding, good luck!'
  );
  process.env.CLIENT_IP = clientIp;
};

const runTests = async config => {
  console.log(chalk.blue('🧪 Running Tests\n'));

  if (config.local || config.environment === environments.staging) {
    const testScript = config.visual ? 'test-visual-new' : 'test';
    const { user, password } = getCredentials(config);
    process.env.STAGING_USER = process.env.STAGING_USER || user;
    process.env.STAGING_PASSWORD = process.env.STAGING_PASSWORD || password;
    console.log(`   Active credentials: ${process.env.STAGING_USER} / ${process.env.STAGING_PASSWORD}`);
    await withSpinner(
      `🏃 Executing ${chalk.cyan(testScript)} with ${chalk.cyan(config.project)}...`,
      async () =>
        await runCommand('npm', ['run', testScript, '--', `--project=${config.project}`], config, { cwd: join(config.guiRepository, 'tests/e2e_tests') }),
      'Local tests completed',
      'Local tests failed'
    );
    return;
  }
  if (!config.skipCleanup) {
    await composeDown(config);
  }
  await composeUp(config);
  try {
    if (config.environment === environments.enterprise) {
      await setupEnterprise(config);
    } else {
      await setupOS(config);
    }
    if (config.variant === testSuiteVariants.qemu) {
      await setupQemuClient(config);
    }
  } catch (error) {
    console.error(chalk.red(`💥 ${config.environment} setup failed:`, JSON.stringify(error)));
    throw error;
  }
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

const program = new Command();
program
  .name('mender-e2e-runner')
  .version('2.0.0')
  .option('-c, --skip-cleanup', 'Leave containers running after tests')
  .option('--local', 'Execute tests on your local machine')
  .option('--local-visual', 'Execute tests in visual mode (implies --local)')
  .option('-f, --file <path>', 'Specify custom compose file', (value, previous) => (previous ? [...previous, value] : [value]))
  .option('-e --environment <env>', `Specify environment to use (${Object.values(environments).join(', ')})`, defaults.environment)
  .option('-p, --project <browser>', `Browser project to run (${Object.values(projects).join(', ')})`, defaults.project)
  .option('-i, --interactive', 'Run in interactive mode with prompts')
  .option('--user <email>', 'User email to use')
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
  ${chalk.green('$ node run.js --local-visual')}                       Run with visual testing mode
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

const mappedEnvironmentOptions = {
  baseUrl: 'BASE_URL',
  environment: 'TEST_ENVIRONMENT',
  guiRepository: 'GUI_REPOSITORY',
  password: 'STAGING_PASSWORD',
  serverRoot: 'SERVER_ROOT',
  user: 'STAGING_USER'
};

const setEnvironmentValues = options =>
  Object.entries(options).forEach(([key, value]) => {
    const variableName = mappedEnvironmentOptions[key];
    if (value && variableName) {
      process.env[variableName] = process.env[variableName] || value;
    }
  });

const validateConfiguration = config => {
  const errors = [];

  if (!projects[config.project]) {
    errors.push(`Invalid project: ${config.project}. Valid projects are: ${Object.values(projects).join(', ')}`);
  }
  if (!environments[config.environment]) {
    errors.push(`Invalid environment: ${config.environment}. Valid environments are: ${Object.values(environments).join(', ')}`);
  }
  const missingFiles = config.composeFiles.filter(file => !existsSync(file));
  if (missingFiles.length > 0) {
    errors.push(`Missing compose files: ${missingFiles.join(', ')}`);
  }
  if (!errors.length) {
    return;
  }
  console.error(chalk.red('💥 Configuration Validation Failed:'));
  errors.forEach(error => console.error(chalk.red(`   • ${error}`)));
  process.exit(1);
};

const showConfiguration = config => {
  console.log(chalk.green('\n📋 Configuration Summary:'));
  console.log(`   Environment: ${config.environment}`);
  console.log(`   Variant: ${config.variant}`);
  console.log(`   Execution: ${config.local ? (config.visual ? chalk.magenta('Local Visual') : chalk.cyan('Local')) : chalk.blue('Docker')}`);
  console.log(`   Browser: ${chalk.cyan(config.project)}`);
  console.log(`   Cleanup: ${config.skipCleanup ? chalk.red('Skip') : chalk.green('Auto')}`);
  console.log(`   Compose Files: ${config.composeFiles}`);
  if (config.user && config.password) {
    console.log(`   Credentials: ${config.user} / ${config.password}`);
  }
  if (config.baseUrl) {
    console.log(`   Running against: ${config.baseUrl}`);
  }
  console.log('');
};

let config; // need to keep this global for SIGINT & -TERM
let currentProcesses = [];
let shutdownInProgress = false;

const main = async () => {
  program.parse();
  const options = program.opts();
  if (options.banner) {
    console.log(banner);
  }
  config = createConfig(options);
  if (config.interactive) {
    await promptForConfiguration();
  } else if (!process.argv.slice(2).length) {
    program.help();
    return;
  }
  validateConfiguration(config);
  showConfiguration(config);
  setEnvironmentValues(config);

  await runTests(config);
  console.log(chalk.green('\n🎉 All tests completed successfully!\n'));
};

const killTestProcesses = () => {
  console.log(chalk.yellow('🔪 Terminating remaining processes...'));
  currentProcesses.forEach(child => {
    if (child?.killed) {
      return;
    }
    try {
      child.kill('SIGTERM');
      setTimeout(() => {
        // clean up without compromise after grace period 🔪🔪🔪
        if (!child.killed) {
          child.kill('SIGKILL');
        }
      }, 3000);
    } catch (error) {
      console.log(chalk.yellow(`Warning: Failed to kill process ${child.pid}: ${error.message}`));
    }
  });
  currentProcesses = [];
};

const cleanup = async (exitCode = 0) => {
  killTestProcesses();
  const logDir = join(config.guiRepository, 'logs');
  const logPath = join(logDir, 'gui_e2e_tests.txt');
  const clientLogPath = join(logDir, 'client.log');

  if (exitCode !== 0) {
    try {
      mkdirSync(logDir, { recursive: true });
      // the client gets often started outside of the compose setup, so track it down by name
      console.log(chalk.yellow(`📋 Capturing client logs to ${chalk.cyan(clientLogPath)}`));
      const containerNames = await runCommand('docker', ['ps', '-a', `--format={{.Names}}`], config);
      const clientContainer = containerNames.split('\n').find(name => name.includes('client'));
      if (clientContainer) {
        const clientLog = await runCommand('docker', ['logs', clientContainer], config);
        writeFileSync(clientLogPath, clientLog);
      }
      console.log(chalk.yellow(`📋 Tests failed, dumping logs to ${chalk.cyan(logPath)}`));
      const logs = await composeLogs(config);
      writeFileSync(logPath, logs);
    } catch (error) {
      console.error(chalk.red('💥 Failed to dump logs:'), error);
    }
  }

  if (config.skipCleanup) {
    console.log(chalk.yellow('⚠️ Containers left running'));
    return;
  }
  await composeDown(config);
};

const initiateShutDownSequence = signal => async () => {
  if (shutdownInProgress) {
    console.log(chalk.yellow(`\n⚡ Shutdown already in progress...`));
    return;
  }
  shutdownInProgress = true;
  console.log(chalk.yellow(`\n⚡ Received ${signal}, cleaning up...`));
  await cleanup(1);
  killTestProcesses();
  process.exit(1);
};

process.on('SIGINT', initiateShutDownSequence('SIGINT'));

process.on('SIGTERM', initiateShutDownSequence('SIGTERM'));

const errorLogCutoffLength = 200;
let exitCode = 0;
try {
  await main();
} catch (error) {
  const errorMessage = JSON.stringify(error);
  const logDir = join(config.guiRepository, 'logs');
  mkdirSync(logDir, { recursive: true });
  const errorPath = join(logDir, 'error.txt');
  console.error(chalk.red('\n💥 Error:', errorMessage.length < errorLogCutoffLength ? errorMessage : `will be dumped to ${errorPath}`));
  if (errorMessage.length >= errorLogCutoffLength) {
    writeFileSync(errorPath, errorMessage);
  }
  exitCode = 1;
} finally {
  await withSpinner('🧹 Cleanup', async () => await cleanup(exitCode), 'so sparkling clean', 'clean up failed');
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
