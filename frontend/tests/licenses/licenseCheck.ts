// Copyright 2023 Northern.tech AS
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
import { DiffTerm, diff } from 'https://deno.land/x/diff_kit@v2.0.4/mod.ts';
import { parseArgs } from 'jsr:@std/cli/parse-args';
import { dirname, fromFileUrl, join, resolve as resolvePath } from 'jsr:@std/path';
// import { dirname, fromFileUrl, resolve as resolvePath } from 'jsr:@std/path/mod.ts';
import { asCSV, asSummary, init } from 'npm:license-checker-rseidelsohn@4.2.10';

const licenseFile = 'directDependencies.csv';
const licenseFileLocation = resolvePath(dirname(fromFileUrl(Deno.mainModule)), licenseFile);

const knownOptions = [
  { key: 'update', description: 'to update the current license file (needs "--allow-write")' },
  { key: 'summary', description: 'give an overview of the license usage' },
  { key: 'help', description: 'the text you are reading right now' }
];
const usageMessage = [
  'This will check for license changes in a package automatically, for other options try:\n',
  ...knownOptions.map(({ key, description }) => `  --${key} - ${description}`),
  ''
].join('\n');

const collectUsedLicenses = (root: string) => new Promise((resolve, reject) => init({ direct: 0, start: root }, (err, packages) => (err ? reject(err) : resolve(packages))));

const createPackageData = packageData => ({ licenses: packageData.licenses ?? 'unknown', repository: packageData.repository || packageData.url || '' });

const simplifyPackageData = packages =>
  Object.entries(packages).reduce((accu, [id, packageData]) => {
    const versionIndex = id.lastIndexOf('@');
    const packageName = id.substring(0, versionIndex);
    const version = id.substring(versionIndex + 1);
    const versionedPackage = { version, ...createPackageData(packageData) };
    if (!accu[packageName]) {
      accu[packageName] = versionedPackage;
    } else if (accu[packageName] && accu[packageName].licenses !== packageData.licenses) {
      accu[id] = versionedPackage;
    }
    return accu;
  }, {});

const checkLicenses = (root: string = '.') =>
  collectUsedLicenses(root)
    .then(simplifyPackageData)
    .then(async packages => {
      const existing = await Deno.readTextFile(licenseFileLocation);
      const current = `${asCSV(packages)}\n`;
      console.log(diff(current, existing, new DiffTerm({ indentWidth: 2 })));
      Deno.exit(existing === current ? 0 : 1);
    });

const writeLicenseFile = (root: string = '.') =>
  collectUsedLicenses(root)
    .then(simplifyPackageData)
    .then(async packages => {
      const newPackageData = `${asCSV(packages)}\n`;
      return await Deno.writeTextFile(licenseFileLocation, newPackageData);
    });

const checkGeneratedLicenses = async (root: string = '.') => {
  const { default: licenses } = await import(join(root, 'licenses.json'), {
    with: { type: 'json' }
  });
  if (!licenses.length) {
    throw new Error('no licenses found, license generation broken');
  }
};

const { '_': _catcher, rootDir: passedRoot, ...flags } = parseArgs(Deno.args, { boolean: knownOptions.map(({ key }) => key), string: ['rootDir'] });

await checkGeneratedLicenses(passedRoot);

const hasUnknownFlag = Object.keys(flags).some(option => !knownOptions.find(({ key }) => option === key));
if (flags.update) {
  await writeLicenseFile(passedRoot);
  Deno.exit(0);
}
if (flags.help || hasUnknownFlag) {
  console.error(usageMessage);
  Deno.exit(0);
}
if (flags.summary) {
  const result = await collectUsedLicenses(passedRoot);
  console.log(asSummary(result));
  Deno.exit(0);
}
checkLicenses(passedRoot);
