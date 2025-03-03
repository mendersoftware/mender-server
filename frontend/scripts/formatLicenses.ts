import { join } from 'jsr:@std/path';

import { rootDir } from './common.js';

type License = {
  author?: string;
  license: string;
  licenseText: string;
  name: string;
  repository: string;
  source: string;
  version: string;
};

const adjustRepoLocation = (repository: string) => {
  let accessibleRepo = repository;
  try {
    const url = new URL(repository);
    accessibleRepo = `https://${url.hostname || 'github.com/'}${url.pathname}`;
  } catch {
    const match = /(git\w+)\.com/.exec(repository);
    if (match) {
      const repo = repository.substring(repository.indexOf(match[0]) + match[0].length + 1);
      accessibleRepo = `https://${match[0]}/${repo}`;
    } else {
      accessibleRepo = `https://github.com/${repository}`; // assume github as the default and 🤞
    }
  }
  return accessibleRepo;
};

const formatLicenseEntry = (licenseRecord: License) => {
  const { name, version, repository, license, licenseText } = licenseRecord;
  const accessibleRepo = adjustRepoLocation(repository);

  return `## ${name}\n
* Name: ${name}
* Version: ${version}
* License: [${license}](${accessibleRepo})\n
\`\`\`\n${licenseText}\n\`\`\`\n`;
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
