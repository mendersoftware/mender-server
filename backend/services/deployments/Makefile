LDFLAGS ?= "-s -w"
BUILDFLAGS ?= -trimpath -ldflags $(LDFLAGS) -tags nopkcs11

include ../Makefile.common

.PHONY: docker
docker: $(platform_targets)
	docker build $(DOCKER_BUILD_ARGS) \
		--build-arg BIN_FILE=$(DOCKER_BIN) \
		--platform $(DOCKER_PLATFORM) \
		-f Dockerfile \
		-t $(DOCKER_TAG) \
		$(GIT_ROOT)

.PHONY: docker-acceptance
docker-acceptance: export DOCKER_BIN = $(subst $(GIT_ROOT),,$(binfile))
docker-acceptance: export BUILDFLAGS += -cover -installsuffix .test
docker-acceptance: export bindir = $(GIT_ROOT)/backend/tests/bin
docker-acceptance:
	@$(MAKE) DOCKER_TAG=$(DOCKER_TAG)-test docker

.PHONY: test-acceptance-run
test-acceptance-run: docker-acceptance
	docker compose -f tests/docker-compose.yml run --rm --use-aliases acceptance-tester $(PYTEST_ARGS)

.PHONY: test-acceptance
test-acceptance: test-acceptance-run
	docker compose -f tests/docker-compose.yml down --remove-orphans -v
