COMPONENT := $(notdir $(shell go list))
GIT_ROOT := $(shell git rev-parse --show-toplevel)
VERSION := $(shell git describe --tag --dirty 2>/dev/null)
DOCFILES := $(wildcard docs/*_api.yml)
ROOTDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

distdir ?= $(GIT_ROOT)/dist
bindir ?= $(distdir)/$(GOOS)/$(GOARCH)
binfile ?= $(bindir)/$(COMPONENT)

CGO_ENABLED ?= 0
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
GOMODDIR := $(shell go list -m -f '{{.Dir}}')

LDFLAGS ?= "-s -w"
BUILDFLAGS ?= -trimpath -ldflags $(LDFLAGS)
TESTFLAGS ?=

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
build-test: BUILDFLAGS += -cover -installsuffix .test -tags acceptance
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
	docker compose -f tests/docker-compose.yml run --rm -it --use-aliases acceptance-tester $(PYTEST_ARGS)

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

tests/%: docs/%.yml
	rm -rf $@; \
	docker run --rm -t -v $(ROOTDIR):$(ROOTDIR) -w $(ROOTDIR) \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v5.3.1 generate \
		-g python -i $< \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$*

.PHONY: docs
docs: $(patsubst docs/%.yml,tests/%,$(DOCFILES))

.PHONY: docker-acceptance
docker-acceptance: export DOCKER_TAG = $(COMPONENT):test
docker-acceptance: export binfile = $(GIT_ROOT)/backend/tests/bin/$(COMPONENT).test
docker-acceptance: export DOCKER_BIN = $(subst $(GIT_ROOT),,$(binfile))
docker-acceptance: export GOOS = linux
docker-acceptance: build-test docs
	docker build $(DOCKER_ARGS) \
		--build-arg BIN_FILE=$(DOCKER_BIN) \
		--platform $(GOOS)/$(GOARCH) \
		-f Dockerfile \
		-t $(DOCKER_TAG) \
		$(GIT_ROOT)
