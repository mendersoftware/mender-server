import { join } from 'jsr:@std/path';

import { rootDir } from './common.js';

type License = {
  name: string;
  version: string;
  author?: string;
  repository: string;
  source: string;
  license: string;
  licenseText: string;
};

const formatLicenseEntry = (licenseRecord: License) => {
  const { name, version, repository, license, licenseText } = licenseRecord;
  return `## ${name}\n
* Name: ${name}
* Version: ${version}
* License: [${license}](${repository})\n
\`\`\`\n${licenseText}\`\`\`\n`;
};

const processLicenseFile = async () => {
  const root = rootDir ?? '.';
  const { default: licenses } = await import(join(root, 'frontend', 'licenses.json'), {
    with: { type: 'json' }
  });
  const licenseEntries = licenses.map(formatLicenseEntry).join('\n');
  await Deno.writeTextFile(join(root, 'frontend', 'licenses.md'), `# Licenses\n\n\n${licenseEntries}`);
};

await processLicenseFile();
