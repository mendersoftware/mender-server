# RFC: Should we make all services explicit to allow auto-complete?
SERVICES = $(subst services/,,$(wildcard services/*))

BUILD_TARGETS := $(addsuffix "-build", $(SERVICES))
TEST_TARGETS := $(addsuffix "-test", $(SERVICES))
DOCKER_TARGETS := $(addsuffix "-docker", $(SERVICES))
DOCKER_ACCEPTANCE_TARGETS := $(addsuffix "-docker-acceptance", $(SERVICES))

.PHONY: build
build: $(BUILD_TARGETS)

.PHONY: $(SERVICES)
$(BUILD_TARGETS):
	@$(MAKE) -C services/$(subst -build,,$@) build


.PHONY: test
test: $(TEST_TARGETS) pkg-test integration-test

.PHONY: $(TEST_TARGETS)
$(TEST_TARGETS):
	@echo $(subst -test,,$@)
	@$(MAKE) -C services/$(subst -test,,$@) test

.PHONY: pkg-test
pkg-test:
	@$(MAKE) -C pkg/

.PHONY: integration-test
integration-test:
	./tests/integration/run

.PHONY: docker
docker: $(DOCKER_TARGETS)

.PHONY: $(DOCKER_TARGETS)
$(DOCKER_TARGETS):
	@$(MAKE) -C services/$(subst -docker,,$@) docker

.PHONY: docker-acceptance
docker-acceptance: $(DOCKER_ACCEPTANCE_TARGETS)

.PHONY: $(DOCKER_ACCEPTANCE_TARGETS)
$(DOCKER_ACCEPTANCE_TARGETS):
	@$(MAKE) -C services/$(subst -docker-acceptance,,$@) docker-acceptance
