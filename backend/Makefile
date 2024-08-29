MENDER_IMAGE_PREFIX ?= localhost:5000/mender-server
MENDER_IMAGE_TAG ?= latest

SERVICES = $(subst services,,$(subst /,,$(wildcard services/*/)))
BUILD_TARGETS := $(addsuffix "-build", $(SERVICES))
TEST_TARGETS := $(addsuffix "-test", $(SERVICES))
TEST_UNIT_TARGETS := $(addsuffix "-test-unit", $(SERVICES))
TEST_ACCEPTANCE_TARGETS := $(addsuffix "-test-acceptance", $(SERVICES))
DOCKER_TARGETS := $(addsuffix "-docker", $(SERVICES))
DOCKER_ACCEPTANCE_TARGETS := $(addsuffix "-docker-acceptance", $(SERVICES))

.PHONY: build
build: $(BUILD_TARGETS)

.PHONY: $(SERVICES)
$(BUILD_TARGETS):
	@$(MAKE) -C services/$(subst -build,,$@) build


.PHONY: test
test: $(TEST_TARGETS) pkg-test test-integration

.PHONY: $(TEST_TARGETS)
$(TEST_TARGETS):
	@echo $(subst -test,,$@)
	@$(MAKE) -C services/$(subst -test,,$@) test

.PHONY: pkg-test
pkg-test:
	@$(MAKE) -C pkg/

.PHONY: test-unit
test-unit: $(TEST_UNIT_TARGETS)

.PHONY: $(TEST_UNIT_TARGETS)
$(TEST_UNIT_TARGETS):
	@echo $(subst -test-unit,,$@)
	@$(MAKE) -C services/$(subst -test-unit,,$@) test-unit

.PHONY: test-acceptance
test-acceptance: $(TEST_ACCEPTANCE_TARGETS)

.PHONY: $(TEST_ACCEPTANCE_TARGETS)
$(TEST_ACCEPTANCE_TARGETS):
	@echo $(subst -test-acceptance,,$@)
	@$(MAKE) -C services/$(subst -test-acceptance,,$@) test-acceptance

.PHONY: test-integration
test-integration:
	./tests/integration/run $(args)

.PHONY: docker
docker: $(DOCKER_TARGETS)

.PHONY: $(DOCKER_TARGETS)
$(DOCKER_TARGETS):
	@$(MAKE) -C services/$(subst -docker,,$@) docker \
		MENDER_IMAGE_PREFIX=$(MENDER_IMAGE_PREFIX) \
		MENDER_IMAGE_TAG=$(MENDER_IMAGE_TAG)

.PHONY: docker-acceptance
docker-acceptance: $(DOCKER_ACCEPTANCE_TARGETS)

.PHONY: $(DOCKER_ACCEPTANCE_TARGETS)
$(DOCKER_ACCEPTANCE_TARGETS):
	@$(MAKE) -C services/$(subst -docker-acceptance,,$@) docker-acceptance
