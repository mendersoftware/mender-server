import react from '@vitejs/plugin-react';
import path from 'path';
import svgr from 'vite-plugin-svgr';
import tsconfigPaths from 'vite-tsconfig-paths';
import { UserWorkspaceConfig, defineConfig } from 'vitest/config';

export default defineConfig(
  () =>
    ({
      plugins: [
        react(),
        svgr({
          svgrOptions: {
            ref: true,
            svgo: false,
            titleProp: true
          },
          include: '**/*.svg'
        }),
        tsconfigPaths({ root: path.resolve(__dirname) })
      ],
      resolve: {
        alias: [
          {
            find: '@northern.tech/store',
            replacement: path.resolve(__dirname, 'src/js/store')
          },
          {
            find: '@northern.tech/common-ui',
            replacement: path.resolve(__dirname, 'src/js/common-ui')
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
          NODE_ENV: 'test',
          PUBLIC_URL: '',
          TZ: 'UTC'
        },
        environment: 'jsdom',
        globals: true,
        setupFiles: './tests/setupTests.jsx',
        fakeTimers: {
          toFake: ['setTimeout', 'clearTimeout', 'Date']
        }
      }
    }) as UserWorkspaceConfig
);
