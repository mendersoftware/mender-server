COMPONENT := create-artifact-worker

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

# Dummy targets, there are no acceptance tests for this service.
.PHONY: test-acceptance-run
test-acceptance-run:

.PHONY: test-acceptance
test-acceptance: test-acceptance-run
