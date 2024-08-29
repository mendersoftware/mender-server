include ../Makefile.common

.PHONY: docker
docker: $(platform_targets)
	$(MAKE) -C ../workflows $(platform_targets)
	docker build $(DOCKER_BUILD_ARGS) \
		--build-arg BIN_FILE=$(DOCKER_BIN) \
		--platform $(DOCKER_PLATFORM) \
		-f Dockerfile \
		-t $(DOCKER_TAG) \
		$(GIT_ROOT)

.PHONY: docker-acceptance
docker-acceptance: export DOCKER_BIN = $(subst $(GIT_ROOT),,$(binfile))
docker-acceptance: export BUILDFLAGS += -cover -installsuffix .test -tags acceptance
docker-acceptance: export bindir = $(GIT_ROOT)/backend/tests/bin
docker-acceptance:
	@$(MAKE) DOCKER_TAG=$(DOCKER_TAG)-test docker


# Dummy targets, there are no acceptance tests for this service.
.PHONY: test-acceptance-run
test-acceptance-run:

.PHONY: test-acceptance
test-acceptance: test-acceptance-run
