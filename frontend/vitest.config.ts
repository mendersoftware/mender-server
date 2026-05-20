import react from '@vitejs/plugin-react';
import path from 'path';
import type { Plugin } from 'vite';
import svgr from 'vite-plugin-svgr';
import type { UserWorkspaceConfig } from 'vitest/config';
import { defineConfig } from 'vitest/config';

// Mostly test spying on redux action execution need a fresh environment.
// Also tests rendering xterm don't have reliable internal counter.
// Add a file here when you need its module state isolated from the rest of the suite.
const isolatedTestFiles = [
  'App.test.tsx',
  'auditlogs/eventdetails/PortForward.test.tsx',
  'auditlogs/eventdetails/TerminalSession.test.tsx',
  'dashboard/Dashboard.test.tsx',
  'dashboard/Deployments.test.tsx',
  'dashboard/SoftwareDistribution.test.tsx',
  'deployments/Deployments.test.tsx',
  'deployments/DeploymentsList.test.tsx',
  'deployments/InProgressDeployments.test.tsx',
  'devices/AuthorizedDevices.test.tsx',
  'devices/device-details/Connection.test.tsx',
  'devices/device-details/Deployments.test.tsx',
  'devices/dialogs/PreauthDialog.test.tsx',
  'devices/troubleshoot/Terminal.test.tsx',
  'devices/troubleshoot/TerminalWrapper.test.tsx',
  'login/Login.test.tsx',
  'login/Password.test.tsx',
  'login/PasswordReset.test.tsx',
  'releases/Releases.test.tsx',
  'releases/dialogs/AddArtifact.test.tsx',
  'releases/manifests/ManifestQuickActions.test.tsx',
  'settings/AccessTokenManagement.test.tsx',
  'settings/ArtifactGeneration.test.tsx',
  'settings/organization/Billing.test.tsx',
  'settings/role-management/Roles.test.tsx',
  'settings/user-management/UserManagement.test.tsx',
  'subscription/SubscriptionConfirmation.test.tsx',
  'subscription/SubscriptionDrawer.test.tsx',
  'subscription/SubscriptionPage.test.tsx',
  'subscription/SubscriptionSummary.test.tsx',
  'tenants/ExpandedTenant.test.tsx',
  'tenants/TenantsForm.test.tsx'
].map(test => `src/js/components/${test}`);

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

export default defineConfig(
  () =>
    ({
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
        server: {
          deps: {
            inline: [/@northern\.tech\/store/]
          }
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
        },
        projects: [
          {
            extends: true,
            test: {
              name: 'fast',
              isolate: false,
              include: ['src/js/**/*.test.{ts,tsx}'],
              exclude: [...isolatedTestFiles]
            }
          },
          {
            extends: true,
            test: {
              name: 'isolated',
              isolate: true,
              include: [...isolatedTestFiles]
            }
          }
        ]
      }
    }) as UserWorkspaceConfig
);
