import { basename, dirname } from 'https://deno.land/std/path/posix/mod.ts';
import { parse, stringify } from 'https://deno.land/std/yaml/mod.ts';
import { camelCase } from 'https://deno.land/x/case/mod.ts';
import { generate } from 'npm:openapi-typescript-codegen';
import converter from 'npm:swagger2openapi';

import { getFiles } from '../common.js';

const apiTypes = ['management'];

const resourceToFileInfo = async res => {
  const filename = res.substring(res.lastIndexOf('/') + 1, res.indexOf('.yml'));
  const apiType = filename.substring(0, filename.indexOf('_'));
  let versionSuffix = filename.substring(filename.lastIndexOf('_'));
  versionSuffix = versionSuffix != '_api' ? versionSuffix : '';
  const service = basename(dirname(res));
  if (apiTypes.includes(apiType)) {
    return { filename, apiType, path: res, service, versionSuffix };
  }
  return;
};

const getFileContents = async file => {
  const fileContent = await Deno.readTextFile(file.path);
  const fileData = parse(fileContent);
  if (!fileData.swagger) {
    return { ...file, fileData };
  }
  const converted = await converter.convertObj(fileData, {});
  return { ...file, fileData: converted.openapi };
};

const sanitizeSchemaRefs = (pathSpec, renamedSchemas) => {
  const text = JSON.stringify(pathSpec);
  const sanitizedText = renamedSchemas.reduce((accu, { key, sanitizedKey }) => accu.replaceAll(`/schemas/${key}"`, `/schemas/${sanitizedKey}"`), text);
  return JSON.parse(sanitizedText);
};

const repairPathSpec = (pathSpec, service, versionSuffix) => {
  const text = JSON.stringify(pathSpec);
  const sanitizedText = text.replaceAll(`"summary":`, `"description":`);
  const content = JSON.parse(sanitizedText);
  return Object.entries(content).reduce(
    (accu, [key, value]) => ({
      ...accu,
      [key]: {
        ...value,
        summary: value.summary ?? value.operationId,
        operationId: camelCase(`${service}${value.operationId || value.summary}${versionSuffix}`)
      }
    }),
    {}
  );
};

const sanitizeSchemas = (schemas, existingSchemas, service, counter = {}) =>
  Object.entries(schemas).reduce(
    (schemaAccu, [key, schema]) => {
      let sanitizedKey = key;
      if (existingSchemas[key]) {
        schemaAccu.counter[key] = (schemaAccu.counter[key] ?? 1) + 1;
        console.log(`duplicate schema, candidate for merge: ${key} (${service}) - total: ${schemaAccu.counter[key]}`);
        sanitizedKey = `${key}${service[0].toUpperCase()}${service.substring(1)}`;
        schemaAccu.renamedSchemas.push({ key, sanitizedKey });
      }
      schemaAccu.serviceSchemas[sanitizedKey] = schema;
      return schemaAccu;
    },
    { renamedSchemas: [], serviceSchemas: {}, counter }
  );

const defaultManagementUrl = 'https://hosted.mender.io/api/management';
const baseSpec = {
  openapi: '3.0.3',
  info: {
    contact: { name: 'support@mender.io' },
    description: 'Combined API specification for the management features of the different Mender backend services, suitable for code generation applications',
    title: 'Mender Management API',
    version: '1.0'
  },
  tags: [{ name: 'Management API', description: 'used for management APIs' }],
  servers: [{ url: defaultManagementUrl }],
  paths: {},
  components: { requestBodies: {}, securitySchemes: {}, schemas: {} }
};

const processFiles = async root => {
  const files = await getFiles(root, { fileProcessor: resourceToFileInfo });
  const fileContents = await Promise.all(files.map(getFileContents));
  return fileContents.reduce((accu, { fileData, service, versionSuffix }) => {
    const { renamedSchemas, serviceSchemas, counter } = sanitizeSchemas(fileData.components.schemas, accu.components.schemas, service, accu.counter);
    const basePath = fileData.servers[0].url.substring(defaultManagementUrl.length);
    const scopedPaths = Object.entries(fileData.paths).reduce((pathAccu, [path, content]) => {
      const sanitizedContent = sanitizeSchemaRefs(content, renamedSchemas);
      pathAccu[`${basePath}${path}`] = repairPathSpec(sanitizedContent, service, versionSuffix);
      return pathAccu;
    }, {});
    return {
      ...accu,
      components: {
        ...accu.components,
        ...fileData.components,
        requestBodies: { ...accu.components.requestBodies, ...fileData.components.requestBodies },
        responses: { ...accu.components.responses, ...fileData.components.responses },
        schemas: { ...accu.components.schemas, ...serviceSchemas },
        securitySchemes: { ...accu.components.securitySchemes, ...fileData.components.securitySchemes }
      },
      counter,
      paths: { ...accu.paths, ...scopedPaths }
    };
  }, baseSpec);
};

const generateTypeIndex = async () => {
  // modify the generated types index to work with the structure we want to use in the gui codebase
  const fileContent = await Deno.readTextFile('./generated/index.ts');
  const lines = fileContent.split('\n').reduce((accu, line) => {
    // skip comments, but otherwise flatten the access to the types (since we discard the generated core & service folders, we can omit the indirection)
    if (line.startsWith('/* ')) {
      accu.push(line);
    } else if (line.includes(`from './models/`)) {
      if (line.endsWith('/models/Settings";')) {
        accu.push('// @ts-ignore');
      }
      accu.push(line.replace('/models', ''));
    }
    return accu;
  }, []);
  await Deno.writeTextFile('./generated/models/MenderTypes.ts', lines.join('\n'));
};

const mergedContent = await processFiles('./specs');
await generate({ input: mergedContent, output: './generated' });
await Deno.writeTextFile('combined.yml', stringify(mergedContent));
await generateTypeIndex();
