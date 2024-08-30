# Makefile interface

## Top-level Makefile
 * `build`: compiles and links all project components.
 * `docker`: builds the docker container images.

* Variables:
  - **MENDER_IMAGE_PREFIX**/**MENDER_IMAGE_TAG**(`docker`): The `docker` targets defines the tag as
    `${MENDER_IMAGE_PREFIX}/<component name>:${MENDER_IMAGE_TAG}`, where `component name` is
    the target component (micro service) being built.
  - **DOCKER_BUILD_ARGS**(`docker`): Additional arguments passed to docker build.
  - **DOCKER_PLATFORM** (`docker`): passed to `--platform` argument to `docker` targets.
    Defaults to the docker system os/architecture for the given context.
    Can be overwritten to a comma-separated list of targets for multiplatform support.

    > NOTE: Docker default context does not support building for multiple platforms.
    >       Use `docker builder create` to initialize a new builder and context for multiplatform builds.

    > NOTE: When building for multiple platforms, you need to set `DOCKER_BUILD_ARGS=--push/--load`.

## backend/Makefile
In addition to top-level target, also includes:
 * `test`: Run all test suites.
 * `test-unit`: Run unit (Go) tests for all packages.
 * `test-acceptance`: Run all service acceptance tests (Pytest).
 * `test-integration`: Run backend integration tests (Pytest).
 * `docker-acceptance`: A `docker` convenience target that builds the target 
   with coverage instrumentaion and appends a `-test` suffix to the build tag.

## backend/services/*/Makefile
Extends `backend/Makefile` with the following additional targets:
 * `generate`: Executes all `go:generate` directives in the source code.
 * `test-acceptance-run`: Run acceptance tests but keep resources running after execution.

Variables:
 * **BUILDFLAGS**(`build`,`test`): flags passed directly to `go build` and `go test`.
 * **TESTFLAGS**(`test`): flags passed directly to `go test`.
 * **bindir**(`build`): Directory to put built binaries.

## frontend/Makefile
No additional targets compared to top-level Makefile.
