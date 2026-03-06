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
import { existsSync } from 'fs';
import inquirer from 'inquirer';
import validator from 'validator';

import { defaults, environments, projects, testSuiteVariants } from './cli.js';
import { createConfig } from './config.js';

export const promptForConfiguration = async () => {
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
      type: 'list',
      name: 'variant',
      message: 'Special variant to be run',
      choices: [
        { name: 'Regular test sets', value: testSuiteVariants.regular },
        { name: 'QEMU dependent tests', value: testSuiteVariants.qemu }
      ],
      default: testSuiteVariants.regular
    }
  ]);

  return createConfig(answers);
};
