# Running E2E Tests Locally

End-to-end tests for the Mender GUI, built with [Playwright](https://playwright.dev/).

## Prerequisites

- **Node.js >= 22** and **npm >= 10** (check with `node -v` and `npm -v`)
- **Docker** and **Docker Compose** (the test runner manages containers automatically)

## Host file setup

The test stack uses the hostname `docker.mender.io` with a self-signed TLS certificate. Add these entries to your hosts file (`/etc/hosts` on macOS/Linux):

```
127.0.0.1 docker.mender.io
127.0.0.1 s3.docker.mender.io
```

Without this, tests will fail trying to reach the Mender server.

## Install dependencies

From this directory (`frontend/tests/e2e_tests`):

```bash
npm install
npx playwright install --with-deps chromium
```

Replace `chromium` with `firefox` or `webkit` if you need a different browser.

## Quickstart: run tests in Docker (recommended)

This starts the full Mender stack in Docker, creates test users, runs the tests inside a container, and tears everything down afterward:

```bash
node run.js --environment os
```

For enterprise tests (requires access to `registry.mender.io`):

```bash
node run.js --environment enterprise
```

That's it — the script handles `docker compose up`, user/tenant setup, test execution, and cleanup.

## Run tests locally (against a running stack)

If you already have a Mender stack running (e.g. via `docker compose up` from the repo root), you can run Playwright directly on your machine instead of inside the test-runner container:

```bash
node run.js --local --environment os
```

This is faster for iteration because it skips container lifecycle management. The stack must be reachable at `https://docker.mender.io/` (or override with `--base-url`).

To open the Playwright UI for headed, interactive debugging:

```bash
node run.js --local-visual --environment os
```

## Run a specific test

With Playwright directly (requires `--local` mode or a running stack):

```bash
# By file path
npx playwright test integration/01-basic/02-login.spec.ts

# By test name pattern
npx playwright test -g "Logs in using UI"

# By project (browser + test level)
npx playwright test --project=basic-chromium
```

## Debugging

### Playwright UI mode

```bash
node run.js --local-visual
```

Opens a browser-based UI where you can step through tests, inspect the DOM, and see screenshots at each step.

### Playwright Inspector

```bash
PWDEBUG=1 npx playwright test integration/01-basic/02-login.spec.ts
```

Pauses at each Playwright action so you can step through the test interactively.

### Trace Viewer

Traces are always recorded (configured in `playwright.config.ts`). After a test run, view traces for failed tests:

```bash
npx playwright show-trace test-results/<test-folder>/trace.zip
```

Screenshots and videos are also captured on failure and saved to `test-results/`.

## Troubleshooting

### Tests fail immediately or containers won't start

Check that Docker has enough disk space — the full stack pulls many images and can exhaust storage quickly:

```bash
docker system df
```

Free up space if needed:

```bash
docker system prune --volumes
```

### Services crash or tests time out

Verify all containers are running and healthy:

```bash
docker compose ps
```

Look for containers in `Exit` or `Restarting` state. Check their logs for the root cause:

```bash
docker compose logs <service-name>
```

A common cause is MongoDB failing to start due to insufficient memory or disk.

### "docker.mender.io" connection refused

Make sure your `/etc/hosts` file has the required entries (see [Host file setup](#host-file-setup) above) and that traefik is up and listening on port 443:

```bash
docker compose ps traefik
```

## CLI reference

All flags for `node run.js`:

| Flag | Default | Description |
|------|---------|-------------|
| `-e, --environment <env>` | `os` | Target environment: `os`, `enterprise`, or `staging` |
| `-p, --project <browser>` | `chromium` | Browser: `chromium`, `firefox`, or `webkit` |
| `--local` | — | Run Playwright on your machine instead of in the Docker test-runner container |
| `--local-visual` | — | Open Playwright UI mode (implies `--local`) |
| `--variant <variant>` | `regular` | `regular` for standard tests, `qemu` for hardware-dependent device tests |
| `--base-url <url>` | — | Override the target server URL |
| `--user <email>` | — | Override the test user email |
| `--password <pass>` | — | Override the test user password |
| `-c, --skip-cleanup` | — | Leave Docker containers running after tests finish |
| `-f, --file <path>` | — | Additional docker-compose override file (can be repeated) |
| `-i, --interactive` | — | Launch interactive configuration prompts |

Default credentials (os/enterprise): `mender-demo@example.com` / `mysecretpassword!123`

## Project structure

```
e2e_tests/
├── integration/               Test specs, ordered by dependency
│   ├── session.setup.ts       Auth setup — runs before all tests
│   ├── 01-basic/              Window/document checks, login
│   ├── 02-baseline/           Releases, settings, SAML, tenants
│   ├── 03-advanced/           Devices, webhooks, deployments, RBAC
│   ├── 04-qemu-dependent/     Hardware-dependent tests (separate config)
│   └── 09-potentially-destructive/
├── fixtures/fixtures.ts       Custom Playwright fixtures (baseUrl, credentials, environment)
├── utils/
│   ├── commands.ts            API-level helpers (login, tenant tokens, docker client)
│   ├── constants.ts           Shared selectors, timeouts, storage paths
│   └── utils.ts               UI-level helpers (release selection, deployment creation)
├── dockerClient/              Mender client config mounted into the Docker client container
├── playwright.config.ts       Main config (60s timeout, 3 browsers × 4 test levels)
├── playwright-qemu.config.ts  QEMU variant config (180s timeout, chromium only)
├── run.js                     Test runner — manages Docker stack and Playwright execution
└── package.json
```

Tests execute in order: `session.setup` → `01-basic` → `02-baseline` → `03-advanced` → `09-potentially-destructive`. Each level depends on the previous completing successfully.
