import react from '@vitejs/plugin-react';
import { cpus } from 'os';
import path from 'path';
import svgr from 'vite-plugin-svgr';
import type { UserWorkspaceConfig} from 'vitest/config';
import { defineConfig } from 'vitest/config';

export default defineConfig(() => {
  const isCi = process.env.CI;
  const cpuCount = cpus().length;
  const threadCount = isCi ? cpuCount / 4 : undefined;

  return {
    plugins: [
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
          // @ts-expect-error -- vite-plugin-svgr's oxcOptions type is not correct
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
    }
  } as UserWorkspaceConfig;
});
