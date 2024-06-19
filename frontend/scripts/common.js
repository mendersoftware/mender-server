import { join, resolve } from 'https://deno.land/std/path/posix/mod.ts';

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
