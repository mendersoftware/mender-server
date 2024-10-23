import { parseArgs } from 'jsr:@std/cli/parse-args';
import { join, resolve } from 'jsr:@std/path';

const { rootDir: passedRoot } = parseArgs(Deno.args, { string: ['rootDir'] });

export const rootDir = passedRoot;

export const getFiles = async (folder, { fileProcessor }) => {
  const files = [];
  for await (const singleDirEntry of Deno.readDir(folder)) {
    if (singleDirEntry.isDirectory) {
      files.push(...(await getFiles(join(folder, singleDirEntry.name), { fileProcessor })));
    } else if (singleDirEntry.isFile) {
      const res = resolve(folder, singleDirEntry.name);
      const fileInfo = await fileProcessor(res);
      if (fileInfo) {
        files.push(fileInfo);
      }
    }
  }
  return files;
};
