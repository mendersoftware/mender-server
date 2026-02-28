# Mender GUI E2E Tests

End-to-end tests for the Mender GUI, built with [Playwright](https://playwright.dev/).

## Prerequisites

- Node.js (with npm)
- Docker and Docker Compose v2+ (for docker execution mode)
- Playwright browsers: `npx playwright install`

Install dependencies:

```bash
npm install
```

## Quick start

The test runner (`run.js`) handles environment setup, test execution, and cleanup.

```bash
# Run tests locally against a running Mender instance
node run.js --local

# Run against a specific URL
node run.js --base-url https://my-mender.example.com

# Run with Playwright UI for debugging
node run.js --visual

# Run in docker (spins up the full Mender stack)
node run.js --environment os
node run.js --environment enterprise
```

Or use the npm script shorthand:

```bash
npm run script -- --local
npm run script -- --environment enterprise
```

## Execution modes

### Local mode (`--local`, `--visual`, `--base-url`)

Runs Playwright directly on your machine against an already-running Mender instance.
No docker-compose orchestration is performed. Use this to reproduce test failures
or develop new tests.

- `--local` runs tests headlessly
- `--visual` opens the Playwright UI (implies `--local`)
- `--base-url <url>` runs against the given URL (implies `--local`, skips infrastructure setup)

### Docker mode (default)

Spins up the full Mender backend via docker-compose, creates users/tenants, runs
Playwright inside a container, collects logs, and tears everything down.

## Environments

| Environment    | Flag                        | Description                                                   |
|----------------|-----------------------------|---------------------------------------------------------------|
| `os`           | `--environment os` (default)| Open Source Mender, creates a single user                     |
| `enterprise`   | `--environment enterprise`  | Enterprise Mender with tenants, addons, and SP tenant setup   |
| `staging`      | `--environment staging`     | Hosted Mender staging (local execution only)                  |

## CLI reference

```
node run.js [options]
```

| Option                     | Description                                          |
|----------------------------|------------------------------------------------------|
| `--local`                  | Execute tests on your local machine                  |
| `--visual`                 | Open Playwright UI (implies `--local`)               |
| `--base-url <url>`         | Target URL (implies `--local`, skips setup)           |
| `-e, --environment <env>`  | `os`, `enterprise`, or `staging` (default: `os`)     |
| `-p, --project <browser>`  | `chromium`, `firefox`, or `webkit` (default: `chromium`) |
| `--variant <variant>`      | `regular` or `qemu` (default: `regular`)             |
| `--user <email>`           | User email for authentication                        |
| `--password <password>`    | User password for authentication                     |
| `-f, --file <path>`        | Additional docker-compose file(s)                    |
| `-c, --skip-cleanup`       | Leave containers running after tests                 |
| `-i, --interactive`        | Run with interactive prompts                         |
| `--no-banner`              | Skip the banner display                              |

### Environment variables

These can be set instead of (or in addition to) CLI flags:

| Variable            | Description                        |
|---------------------|------------------------------------|
| `SERVER_ROOT`       | Root of the mender-server repo     |
| `GUI_REPOSITORY`    | Path to the frontend directory     |
| `TEST_ENVIRONMENT`  | `os`, `enterprise`, or `staging`   |
| `STAGING_USER`      | User email                         |
| `STAGING_PASSWORD`  | User password                      |
| `BASE_URL`          | Target URL                         |

CLI flags take priority over environment variables.

## Test structure

```
e2e_tests/
  run.js                          # Entry point
  runner/                         # Test runner modules
    cli.js                        #   CLI definition and constants
    config.js                     #   Configuration, credentials, validation
    compose.js                    #   Docker compose operations, runCommand, withSpinner
    test-execution.js             #   Environment setup + test orchestration
    lifecycle.js                  #   Process cleanup, log collection, shutdown
    interactive.js                #   Interactive prompt mode
  integration/                    # Playwright test suites (ordered by phase)
    session.setup.ts              #   Login & session bootstrap
    session.teardown.ts           #   Cleanup (docker client, storage, temp files)
    01-basic/                     #   Basic functionality tests
    02-baseline/                  #   Baseline tests (depend on 01)
    03-advanced/                  #   Advanced tests (depend on 02)
    04-qemu-dependent/            #   QEMU device tests
    09-potentially-destructive/   #   Destructive tests (run last)
  fixtures/                       # Playwright fixtures and test data
  utils/                          # Shared test utilities and constants
  docker-compose.e2e-tests*.yml   # Compose overrides for test environments
  playwright.config.ts            # Main Playwright configuration
  playwright-qemu.config.ts       # QEMU variant Playwright configuration
```

### Test execution order

Playwright projects are chained via `dependencies` to enforce ordering:

```
setup-{browser}
  -> basic-{browser}
    -> baseline-{browser}
      -> advanced-{browser}
        -> {browser} (potentially-destructive)
          -> teardown
```

Each phase runs only after its dependency completes. The `teardown` project
runs after the final test phase and cleans up docker clients, storage state,
and temporary files.

## Examples

```bash
# Enterprise tests with Firefox
node run.js --environment enterprise --project firefox

# QEMU device tests (enterprise only)
node run.js --environment enterprise --variant qemu

# Run against a preview deployment
node run.js --base-url https://preview.example.com \
  --environment enterprise \
  --user user@example.com \
  --password mypassword

# Keep containers running for debugging after failure
node run.js --environment enterprise --skip-cleanup

# Run Playwright directly (if setup is already done)
npx playwright test --project=chromium
```
