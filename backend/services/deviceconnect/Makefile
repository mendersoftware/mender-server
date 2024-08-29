include ../Makefile.common

tests/%: docs/%.yml
	[ -e $@ ] && rm -r $@; \
	docker run --rm -t -v $(MAKEDIR):/work -w /work \
		--ulimit nofile=65536:65536 \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v4.3.1 generate \
		-g python -i $< \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$*

.PHONY: docs
docs: $(patsubst docs/%.yml,tests/%,$(DOCFILES))

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
test-acceptance-run: docker-acceptance docs
	docker compose -f tests/docker-compose.yml run --rm --use-aliases acceptance-tester $(PYTEST_ARGS)

.PHONY: test-acceptance
test-acceptance: test-acceptance-run
	docker compose -f tests/docker-compose.yml down --remove-orphans -v
