COMPONENT := $(notdir $(shell go list))
GIT_ROOT := $(shell git rev-parse --show-toplevel)
VERSION := $(shell git describe --tag --dirty 2>/dev/null)

distdir ?= $(GIT_ROOT)/dist
bindir ?= $(distdir)/$(GOOS)/$(GOARCH)
binfile ?= $(bindir)/$(COMPONENT)

CGO_ENABLED ?= 0
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
GOMODDIR := $(shell go list -m -f '{{.Dir}}')

LDFLAGS ?= "-s -w"
BUILDFLAGS ?= -tags nopkcs11 -trimpath -ldflags $(LDFLAGS)
TESTFLAGS ?=
PYTEST_ARGS ?=

DOCKER_TAG ?= $(COMPONENT):latest
DOCKER_BIN ?= $(subst $(GIT_ROOT),,$(binfile)) # Repo-local file to binary

.PHONY: build
build:
	# TODO: Add -ldflags "-X ...Version"
	env CGO_ENABLED=$(CGO_ENABLED) \
		GOOS=$(GOOS) \
		GOARCH=$(GOARCH) \
		go build -o $(binfile) $(BUILDFLAGS)

.PHONY: build-test
build-test: BUILDFLAGS += -cover -installsuffix .test
build-test: binfile = $(GIT_ROOT)/backend/tests/bin/$(COMPONENT).test
build-test: build

.PHONY: generate
generate:
	go generate ./...

.PHONY: test-unit
test-unit:
	go test $(BUILDFLAGS) $(TESTFLAGS) ./...

.PHONY: test-acceptance-run
test-acceptance-run: docker-acceptance
	docker compose -f tests/docker-compose.yml run --rm --use-aliases acceptance-tester $(PYTEST_ARGS)

.PHONY: test-acceptance
test-acceptance: test-acceptance-run
	docker compose -f tests/docker-compose.yml down --remove-orphans -v

.PHONY: test
test: test-unit test-acceptance

.PHONY: docker
docker: export GOOS = linux
docker: build
	docker build $(DOCKER_ARGS) \
		--build-arg BIN_FILE=$(DOCKER_BIN) \
		--platform $(GOOS)/$(GOARCH) \
		-f Dockerfile \
		-t $(DOCKER_TAG) \
		$(GIT_ROOT)

.PHONY: docker-acceptance
docker-acceptance: export DOCKER_TAG = $(COMPONENT):test
docker-acceptance: export binfile = $(GIT_ROOT)/backend/tests/bin/$(COMPONENT).test
docker-acceptance: export DOCKER_BIN = $(subst $(GIT_ROOT),,$(binfile))
docker-acceptance: export GOOS = linux
docker-acceptance: build-test
	docker build $(DOCKER_ARGS) \
		--build-arg BIN_FILE=$(DOCKER_BIN) \
		--platform $(GOOS)/$(GOARCH) \
		-f Dockerfile \
		-t $(DOCKER_TAG) \
		$(GIT_ROOT)
