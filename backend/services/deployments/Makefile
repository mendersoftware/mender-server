include ../Makefile.common

BUILDFLAGS += -tags nopkcs11

.PHONY: docker
docker: $(platform_targets)
	docker build $(DOCKER_BUILD_ARGS) \
		--build-arg BIN_FILE=$(DOCKER_BIN) \
		--platform $(DOCKER_PLATFORM) \
		-f Dockerfile \
		-t $(DOCKER_TAG) \
		$(GIT_ROOT)

.PHONY: test-acceptance-run
test-acceptance-run: docker-acceptance
	MENDER_IMAGE_PREFIX=$(MENDER_IMAGE_PREFIX) \
		MENDER_IMAGE_TAG=$(MENDER_IMAGE_TAG) \
		docker compose -f tests/docker-compose.yml run --rm \
		--use-aliases --remove-orphans \
		acceptance-tester $(PYTEST_ARGS)

.PHONY: test-acceptance
test-acceptance: test-acceptance-run
	docker compose -f tests/docker-compose.yml down --remove-orphans -v
