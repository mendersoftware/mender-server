import { getFiles } from './common.js';

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

const encoder = new TextEncoder();
const decoder = new TextDecoder();

const sourceFilesRegex = new RegExp('[j|t]sx?$');
const resourceToFileInfo = async res => {
  if (!sourceFilesRegex.test(res)) {
    return;
  }
  const command = `git log --diff-filter=A --follow --format=%aI -- ${res} | tail -1`;
  const executor = Deno.run({ cmd: ['bash'], stdout: 'piped', stdin: 'piped' });
  await executor.stdin.write(encoder.encode(command));
  await executor.stdin.close();
  const output = await executor.output();
  const time = decoder.decode(output);
  await executor.close();
  const extension = res.substring(res.lastIndexOf('.'));
  const filename = res.substring(res.lastIndexOf('/') + 1, res.lastIndexOf(extension));
  return { birthyear: time.substring(0, time.indexOf('-')), extension, filename, path: res };
};

const processFiles = async root => {
  const files = await getFiles(root, { fileProcessor: resourceToFileInfo });
  return files.map(async ({ birthyear, extension, path }) => {
    const fileContent = await Deno.readTextFile(path);
    if (fileContent.startsWith('/*')) {
      return;
    }
    const licenseHeader = getLicenseHeader(birthyear, extension);
    const newContent = licenseHeader.concat(fileContent);
    await Deno.writeTextFile(path, newContent);
  });
};

await processFiles('src');
await processFiles('tests');
