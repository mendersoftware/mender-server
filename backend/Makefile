MENDER_IMAGE_PREFIX ?= localhost:5000/mender-server
MENDER_IMAGE_TAG ?= latest
MENDER_IMAGE_TAG_TEST ?= test

SERVICES = $(subst services,,$(subst /,,$(wildcard services/*/)))
BUILD_TARGETS := $(addsuffix -build, $(SERVICES))
TEST_TARGETS := $(addsuffix -test, $(SERVICES))
TEST_UNIT_TARGETS := $(addsuffix -test-unit, $(SERVICES))
TEST_ACCEPTANCE_TARGETS := $(addsuffix -test-acceptance, $(SERVICES))
DOCKER_TARGETS := $(addsuffix -docker, $(SERVICES))
DOCKER_PULL_TARGETS := $(addsuffix -docker-pull, $(SERVICES))
BUILD_ACCEPTANCE_TARGETS := $(addsuffix -build-acceptance, $(SERVICES))
DOCKER_ACCEPTANCE_TARGETS := $(addsuffix -docker-acceptance, $(SERVICES))
DOCKER_PUBLISH_TARGETS := $(addsuffix -docker-publish, $(SERVICES))

.PHONY: docker
docker: $(DOCKER_TARGETS)

.PHONY: $(DOCKER_TARGETS)
$(DOCKER_TARGETS):
	@$(MAKE) -C services/$(subst -docker,,$@) docker \
		MENDER_IMAGE_PREFIX=$(MENDER_IMAGE_PREFIX) \
		MENDER_IMAGE_TAG=$(MENDER_IMAGE_TAG)

.PHONY: docker-pull
docker-pull: $(DOCKER_PULL_TARGETS)

.PHONY: $(DOCKER_PULL_TARGETS)
$(DOCKER_PULL_TARGETS):
	@$(MAKE) -C services/$(subst -docker-pull,,$@) docker-pull \
		MENDER_IMAGE_PREFIX=$(MENDER_IMAGE_PREFIX) \
		MENDER_IMAGE_TAG=$(MENDER_IMAGE_TAG)

.PHONY: build
build: $(BUILD_TARGETS)

.PHONY: $(SERVICES)
$(BUILD_TARGETS):
	@$(MAKE) -C services/$(subst -build,,$@) build


.PHONY: test
test: $(TEST_TARGETS) pkg-test test-integration

.PHONY: $(TEST_TARGETS)
$(TEST_TARGETS):
	@$(MAKE) -C services/$(subst -test,,$@) test

.PHONY: pkg-test
pkg-test:
	@$(MAKE) -C pkg/

.PHONY: test-unit
test-unit: $(TEST_UNIT_TARGETS)

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
test-integration: MENDER_IMAGE_TAG=$(MENDER_IMAGE_TAG_TEST)
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
	@$(MAKE) -C services/$(subst -docker-acceptance,,$@) docker-acceptance \
		MENDER_IMAGE_PREFIX=$(MENDER_IMAGE_PREFIX) \
		MENDER_IMAGE_TAG_TEST=$(MENDER_IMAGE_TAG_TEST)

.PHONY: docker-publish
docker-publish: $(DOCKER_PUBLISH_TARGETS)

.PHONY: $(DOCKER_PUBLISH_TARGETS)
$(DOCKER_PUBLISH_TARGETS):
	@$(MAKE) -C services/$(subst -docker-publish,,$@) docker-publish \
		MENDER_IMAGE_PREFIX=$(MENDER_IMAGE_PREFIX) \
		MENDER_IMAGE_TAG=$(MENDER_IMAGE_TAG) \
		MENDER_PUBLISH_TAG=$(MENDER_PUBLISH_TAG)
