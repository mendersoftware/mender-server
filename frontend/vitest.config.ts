import react from '@vitejs/plugin-react';
import { cpus } from 'os';
import path from 'path';
import type { Plugin } from 'vite';
import svgr from 'vite-plugin-svgr';
import type { UserWorkspaceConfig } from 'vitest/config';
import { defineConfig } from 'vitest/config';

// `@mui/icons-material` v9 has 10k+ strict-ESM files and per-icon package exports entries.
// A barrel import drags all of them into the module graph, adding ~75s per test file on the current suite.
// Rewrite `import { Icon as MyIcon } from '@mui/icons-material'` into deep imports
// (import MyIcon from '@mui/icons-material/MyIcon') so only the icons actually referenced get loaded.
const muiIconsDeepImports = (): Plugin => ({
  name: 'mui-icons-deep-imports',
  enforce: 'pre',
  transform(code) {
    if (!code.includes('@mui/icons-material')) return null;
    const barrel = /import\s*\{([^}]+)\}\s*from\s*['"]@mui\/icons-material['"]\s*;?/g;
    const transformed = code.replace(barrel, (_, names: string) =>
      names
        .split(',')
        .map((n: string) => n.trim())
        .filter(Boolean)
        .map((n: string) => {
          const [orig, alias] = n.split(/\s+as\s+/).map(s => s.trim());
          return `import ${alias || orig} from '@mui/icons-material/${orig}';`;
        })
        .join('\n')
    );
    return transformed === code ? null : { code: transformed, map: null };
  }
});

export default defineConfig(() => {
  const isCi = process.env.CI;
  const cpuCount = cpus().length;
  const threadCount = isCi ? cpuCount / 4 : undefined;

  return {
    plugins: [
      muiIconsDeepImports(),
      react(),
      svgr({
        svgrOptions: {
          ref: true,
          svgo: false,
          titleProp: true,
          jsxRuntime: 'classic'
        },
        include: '**/*.svg',
        oxcOptions: {
          jsx: { runtime: 'classic' }
        }
      })
    ],

    resolve: {
      tsconfigPaths: true,
      alias: [
        {
          find: '@northern.tech/common-ui',
          replacement: path.resolve(__dirname, 'src/js/common-ui')
        },
        {
          find: '@/testUtils',
          replacement: path.resolve(__dirname, 'tests', 'testUtils')
        }
      ]
    },
    server: {
      port: 80,
      middlewareMode: false
    },
    test: {
      coverage: {
        reporter: ['json', 'lcov'],
        reportsDirectory: 'coverage'
      },
      env: {
        BABEL_ENV: 'test',
        LANG: 'en_US.UTF-8',
        LC_ALL: 'en_US.UTF-8',
        NODE_ENV: 'test',
        PUBLIC_URL: '',
        TZ: 'UTC'
      },
      environment: 'jsdom',
      globals: true,
      locale: 'en-US',
      setupFiles: path.resolve(__dirname, 'tests', 'setupTests.ts'),
      fakeTimers: {
        toFake: ['setTimeout', 'clearTimeout', 'setInterval', 'clearInterval', 'Date', 'requestAnimationFrame', 'cancelAnimationFrame']
      }
    },
    pool: 'threads',
    poolOptions: {
      threads: {
        minThreads: threadCount,
        maxThreads: threadCount,
        useAtomics: true
      }
    },
    disableConsoleIntercept: true
  } as UserWorkspaceConfig;
});
