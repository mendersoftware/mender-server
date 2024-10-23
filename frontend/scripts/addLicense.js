import { extname, join } from 'jsr:@std/path';

import { getFiles, rootDir } from './common.js';

const commentByExtension = {
  '.js': '//',
  '.jsx': '//',
  '.sh': '#',
  '.ts': '//',
  '.tsx': '//',
  '.yaml': '#',
  '.yml': '#'
};

const getLicenseHeader = (year, extension) => {
  const comment = commentByExtension[extension];
  return `${comment} Copyright ${year} Northern.tech AS
${comment}
${comment}    Licensed under the Apache License, Version 2.0 (the "License");
${comment}    you may not use this file except in compliance with the License.
${comment}    You may obtain a copy of the License at
${comment}
${comment}        http://www.apache.org/licenses/LICENSE-2.0
${comment}
${comment}    Unless required by applicable law or agreed to in writing, software
${comment}    distributed under the License is distributed on an "AS IS" BASIS,
${comment}    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
${comment}    See the License for the specific language governing permissions and
${comment}    limitations under the License.
`;
};

const decoder = new TextDecoder();

const sourceFilesRegex = new RegExp('[j|t]sx?$');
const existingHeaders = [`'use strict';\n\n// Copyright`, '/*', '// Copyright'];
const resourceToFileInfo = async res => {
  if (!sourceFilesRegex.test(res)) {
    return;
  }
  const fileContent = await Deno.readTextFile(res);
  if (existingHeaders.some(start => fileContent.startsWith(start))) {
    return;
  }
  const command = new Deno.Command('git', { args: ['log', '--diff-filter=A', '--follow', '--format=%aI', '--', res] });
  const { stdout } = await command.output();
  const output = decoder.decode(stdout);
  const time = output.split(`\n`).reduceRight((accu, time) => (accu ? accu : time), '');
  const extension = extname(res);
  const birthyear = time.substring(0, time.indexOf('-'));
  return { birthyear, extension, fileContent, path: res };
};

const processFiles = async root => {
  const files = await getFiles(root, { fileProcessor: resourceToFileInfo });
  return files.map(async ({ birthyear, extension, fileContent, path }) => {
    const licenseHeader = getLicenseHeader(birthyear, extension);
    const newContent = licenseHeader.concat(fileContent);
    await Deno.writeTextFile(path, newContent);
  });
};

await processFiles(join(rootDir, 'frontend', 'src'));
await processFiles(join(rootDir, 'frontend', 'tests'));
