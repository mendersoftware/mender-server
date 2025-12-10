MENDER_IMAGE_REGISTRY ?= docker.io
MENDER_IMAGE_REPOSITORY ?= mendersoftware
MENDER_IMAGE_TAG ?= latest

MENDER_PUBLISH_REGISTRY ?= docker.io
MENDER_PUBLISH_REPOSTIORY ?= mendersoftware
MENDER_PUBLISH_TAG ?= $(MENDER_IMAGE_TAG)

SERVICES = $(subst services,,$(subst /,,$(sort $(dir $(wildcard services/*/)))))
BUILD_TARGETS := $(addsuffix -build, $(SERVICES))
TEST_TARGETS := $(addsuffix -test, $(SERVICES))
TEST_UNIT_TARGETS := $(addsuffix -test-unit, $(SERVICES))
TEST_ACCEPTANCE_TARGETS := $(addsuffix -test-acceptance, $(SERVICES))
DOCKER_TARGETS := $(addsuffix -docker, $(SERVICES))
DOCKER_PULL_TARGETS := $(addsuffix -docker-pull, $(SERVICES))
BUILD_ACCEPTANCE_TARGETS := $(addsuffix -build-acceptance, $(SERVICES))
DOCKER_ACCEPTANCE_TARGETS := $(addsuffix -docker-acceptance, $(SERVICES))
DOCKER_PUBLISH_TARGETS := $(addsuffix -docker-publish, $(SERVICES))
REDOCLY := npx @redocly/cli@2.2.2
COMBINED_SPEC := docs/api/dist/openapi.yaml
OAPI_CODEGEN_VERSION := v2.4.1
OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)

.PHONY: generate-openapi-spec
generate-openapi-spec:
	@echo "Regenerating combined OpenAPI spec..."
	@cd docs/api && $(REDOCLY) join --prefix-tags-with-info-prop title \
		root.openapi.yaml \
		common/*.yaml \
		deployments/*.yaml \
		deviceauth/*.yaml \
		deviceconfig/*.yaml \
		deviceconnect/*.yaml \
		inventory/*.yaml \
		iot-manager/*.yaml \
		useradm/*.yaml \
		workflows/*.yaml \
		-o dist/openapi.tmp.yaml
	@cd docs/api && $(REDOCLY) bundle dist/openapi.tmp.yaml --config config/redocly.yaml -o dist/openapi.yaml
	@rm -f docs/api/dist/openapi.tmp.yaml
	@echo "Generated $(COMBINED_SPEC)"

.PHONY: docker
docker: $(DOCKER_TARGETS)

.PHONY: $(DOCKER_TARGETS)
$(DOCKER_TARGETS):
	@$(MAKE) -C services/$(subst -docker,,$@) docker

.PHONY: docker-pull
docker-pull: $(DOCKER_PULL_TARGETS)

.PHONY: $(DOCKER_PULL_TARGETS)
$(DOCKER_PULL_TARGETS):
	@$(MAKE) -C services/$(subst -docker-pull,,$@) docker-pull

.PHONY: build
build: $(BUILD_TARGETS)

.PHONY: $(SERVICES)
$(BUILD_TARGETS):
	@$(MAKE) -C services/$(subst -build,,$@) build


.PHONY: test
test: test-unit test-integration

.PHONY: $(TEST_TARGETS)
$(TEST_TARGETS):
	@$(MAKE) -C services/$(subst -test,,$@) test

.PHONY: pkg-test
pkg-test:
	@$(MAKE) -C pkg/ test

.PHONY: test-unit
test-unit: pkg-test $(TEST_UNIT_TARGETS)

.PHONY: $(TEST_UNIT_TARGETS)
$(TEST_UNIT_TARGETS):
	@$(MAKE) -C services/$(subst -test-unit,,$@) test-unit

.PHONY: test-acceptance
test-acceptance: $(TEST_ACCEPTANCE_TARGETS)
		docker compose -p acceptance down --remove-orphans

.PHONY: $(TEST_ACCEPTANCE_TARGETS)
$(TEST_ACCEPTANCE_TARGETS):
	@$(MAKE) -C services/$(subst -test-acceptance,,$@) test-acceptance-run

.PHONY: test-integration
test-integration:
	./tests/integration/run $(args)

.PHONY: build-acceptance
build-acceptance: $(BUILD_ACCEPTANCE_TARGETS)

.PHONY: $(BUILD_ACCEPTANCE_TARGETS)
$(BUILD_ACCEPTANCE_TARGETS):
	@$(MAKE) -C services/$(subst -build-acceptance,,$@) build-acceptance

.PHONY: docker-acceptance
docker-acceptance: $(DOCKER_ACCEPTANCE_TARGETS)

.PHONY: $(DOCKER_ACCEPTANCE_TARGETS)
$(DOCKER_ACCEPTANCE_TARGETS):
	@$(MAKE) -C services/$(subst -docker-acceptance,,$@) docker-acceptance

.PHONY: docker-publish
docker-publish: $(DOCKER_PUBLISH_TARGETS)

.PHONY: $(DOCKER_PUBLISH_TARGETS)
$(DOCKER_PUBLISH_TARGETS):
	@$(MAKE) -C services/$(subst -docker-publish,,$@) docker-publish

.PHONY: generate-api-clients
generate-api-clients: generate-api-types generate-api-internal generate-api-management generate-api-device

.PHONY: generate-api-types
generate-api-types:
	@echo "Generating shared API types..."
	@if [ -f pkg/api/types/generate.yaml ]; then \
		$(OAPI_CODEGEN) --config=pkg/api/types/generate.yaml $(COMBINED_SPEC); \
	else \
		echo "  Skipping: pkg/api/types/generate.yaml not found"; \
	fi

.PHONY: generate-api-internal
generate-api-internal:
	@echo "Generating internal API clients..."
	@for config in $$(find pkg/api/internalapi -name 'generate.yaml' 2>/dev/null); do \
		echo "  $$(dirname $$config)..."; \
		$(OAPI_CODEGEN) --config="$$config" $(COMBINED_SPEC); \
	done

.PHONY: generate-api-management
generate-api-management:
	@echo "Generating management API clients..."
	@for config in $$(find pkg/api/management -name 'generate.yaml' 2>/dev/null); do \
		echo "  $$(dirname $$config)..."; \
		$(OAPI_CODEGEN) --config="$$config" $(COMBINED_SPEC); \
	done

.PHONY: generate-api-device
generate-api-device:
	@echo "Generating device API clients..."
	@for config in $$(find pkg/api/device -name 'generate.yaml' 2>/dev/null); do \
		echo "  $$(dirname $$config)..."; \
		$(OAPI_CODEGEN) --config="$$config" $(COMBINED_SPEC); \
	done
