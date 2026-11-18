# Contributing to the Mender Server

Thank you for considering a contribution to the Mender server! We welcome
contributions from the community.

This guide applies to the **server backend** — the Go services that live in the
[`backend/`](./backend) directory. If you are contributing to the web UI in
[`frontend/`](./frontend), please read the dedicated
[frontend contribution guide](./frontend/CONTRIBUTING.md) instead; it covers the
JavaScript/TypeScript toolchain (Prettier, React Testing Library, npm) that the
backend guidelines below do not.

## Reporting security issues

If you discover a security issue, please bring it to our attention right away!
Please **DO NOT** file a public issue, instead send your report privately to
[security@mender.io](mailto:security@mender.io). Security reports are greatly
appreciated and we will publicly thank you for it.

## Proposed tasks to get started

If you are looking for something to work on, issues labelled with the
`helpwanted` tag are a good place to start. You can browse them in our
[issue tracker](https://northerntech.atlassian.net/jira/software/c/projects/MEN/list)
(the `MEN` project on Jira).

Before starting on a larger change, it is usually best to find an existing issue
or open a new one to discuss the change you want to make, so we can agree on the
approach before you invest significant effort. For broader design discussions,
the [Mender Hub forum](https://hub.mender.io) is a good place to reach the team
and the community.

## Providing pull requests

When you contribute a change, please keep the following in mind:

- For anything beyond a small fix, discuss the work in an issue or on
  [Mender Hub](https://hub.mender.io) first, so we can align on the approach.
- Every commit must be signed off (see [Sign your work](#sign-your-work) below).
- Structure your work as small, cohesive, logically self-contained commits
  (see [Structuring your commits](#structuring-your-commits) below).
- Make sure the test suite passes and your change is covered by tests.
- Provide a clear pull request title and description: explain *what* changed and
  *why*, and reference the relevant ticket (e.g. `Ticket: MEN-1234`).

## Programming style

The server backend is written in **Go** (the module
`github.com/mendersoftware/mender-server`, Go 1.25). New code should match the
conventions already present in the [`backend/`](./backend) tree.

### Formatting and linting

- Format all Go code with `gofmt` and organise imports with `goimports`. Imports
  are grouped with the local prefix `github.com/mendersoftware/mender-server`, so
  project-internal imports form their own block.
- Linting is enforced with [`golangci-lint`](https://golangci-lint.run/) using
  the configuration in [`backend/.golangci.yml`](./backend/.golangci.yml). The
  enabled linters include `bodyclose`, `gocyclo`, `gosec`, `govet` and `lll`.
- Keep lines within **100 columns** (`lll`), indent with tabs, and keep
  cyclomatic complexity under the configured limit (`gocyclo`, currently 20). If
  a function grows past that, prefer splitting it over silencing the linter.
- CI runs the formatter and linter; pull requests that do not pass will fail
  their checks. Run them locally before submitting.

### Service structure

The backend is a single Go module containing one microservice per directory
under [`backend/services/`](./backend/services) (for example `deployments`,
`deviceauth`, `inventory`, `useradm`, `workflows`). Code shared across services
lives under [`backend/pkg/`](./backend/pkg) (logging, config, MongoDB helpers,
request IDs, routing, identity, and the generated OpenAPI client).

Each service follows the same layered layout. When adding or extending a
service, mirror it:

- `main.go` / `server.go` — entrypoint and server wiring.
- `api/http/` — the HTTP transport layer (routing, request/response handling).
- `app/` — business logic, kept independent of transport and storage.
- `store/mongo/` — the persistence layer (MongoDB).
- `model/` — domain types shared between the layers above.
- `client/`, `config/`, `utils/` — service-local clients, configuration and
  helpers as needed.
- `Makefile` and `Dockerfile` — per-service build and image.

The dependency direction is `api → app → store`, with `model` shared across
them; keep business logic out of the transport and storage layers. Put anything
genuinely reusable across services in `backend/pkg/` rather than duplicating it.

### Building and testing

The backend is driven from [`backend/Makefile`](./backend/Makefile), which
discovers services automatically. The most useful targets:

- `make build` — build all services.
- `make test-unit` — run unit tests for `pkg/` and every service.
- `make test` — unit plus integration tests.
- `make <service>-build` / `make <service>-test-unit` — operate on a single
  service.

Write tests alongside the code they cover. Table-driven tests are the norm in
this codebase — follow the patterns already in the service you are touching, and
make sure new behaviour is covered before opening a pull request.

For OpenAPI changes, regenerate the spec and client with
`make generate-openapi-all`.

## Sign your work

The sign-off is a simple line at the end of the explanation for the patch, which
certifies that you wrote it or otherwise have the right to pass it on as an
open-source patch. The rules are pretty simple: if you can certify the below
(from [developercertificate.org](https://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then you just add a line to every git commit message:

```
Signed-off-by: Joe Smith <joe.smith@email.com>
```

Use your real name (sorry, no pseudonyms or anonymous contributions).

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`. The sign-off is required and enforced
by `commitlint` (see below).

## Structuring your commits

Please structure your work as a series of small, cohesive, logically
self-contained commits — each commit should do one thing and build and test
cleanly on its own. Avoid mixing unrelated changes in a single commit, and avoid
"fixup" noise in the final history.

Commit messages follow the
[Conventional Commits](https://www.conventionalcommits.org/) format (a slightly
modified variant). This is enforced by the repository's
[`commitlint.config.js`](./commitlint.config.js), which is run as a commit hook.
Please read it — it is the source of truth for the allowed types, scopes and
rules. In summary:

- Allowed scopes match the backend services and shared areas:
  `create-artifact-worker`, `deployments`, `deviceauth`, `deviceconfig`,
  `deviceconnect`, `inventory`, `iot-manager`, `pkg`, `useradm`, `workflows`,
  plus `gui` and `e2e`. `feat`, `fix` and `test` may carry a scope; `perf` and
  `ci` take none.
- The commit body wraps at 100 columns and must be preceded by a blank line.
- Every commit must carry a `Signed-off-by:` trailer (enforced; see
  [Sign your work](#sign-your-work)).

A bug-fix commit looks like:

```
fix: correct release pagination off-by-one

A longer description of the problem and the fix, wrapped at 100 columns.

Ticket: MEN-1234
Signed-off-by: Joe Smith <joe.smith@email.com>
```

A feature commit looks like:

```
feat(deployments): add per-device deployment status filter

A longer description of the change and the motivation, wrapped at 100 columns.

Ticket: MEN-1234
Signed-off-by: Joe Smith <joe.smith@email.com>
```

Set the `Ticket:` trailer to the relevant Jira ticket, or `Ticket: None` if
there is no associated ticket. You can run `npm run commit` for an interactive
prompt that builds a conforming message, and `git commit --no-verify` only when
you have a good reason to bypass the hook.

## Use of AI/LLMs in contributions

We welcome the use of AI assistants and large language models when contributing,
within the following bounds:

- A human contributor must review, understand, and be able to explain and
  correct every line they submit. You are responsible for your contribution
  regardless of how it was produced.
- AI-assisted contributions are held to exactly the same quality, testing and
  review standards as any other contribution.
- You are the author of the contribution; submitting it (including the
  `Signed-off-by` line) certifies the DCO above applies just as it would for
  hand-written code.
- Ensure your use complies with applicable law (including the EU AI Act) and
  with the terms of service of the AI provider you use.
- These guidelines may evolve as the technology and our practices do.

## Contributor Code of Conduct

This project adheres to a Contributor Code of Conduct. By participating, you are
expected to uphold it. Please report unacceptable behaviour as described in the
project's code of conduct.

## Let us work together with you

We are always happy to help contributors get their changes merged. If you are
unsure about anything — the approach, the structure, the tests — open a draft
pull request or reach out on [Mender Hub](https://hub.mender.io), and we will
work through it with you.
