MENDER_IMAGE_PREFIX ?= localhost:5000/mender-server
MENDER_IMAGE_TAG ?= latest

DOCKER_TAG ?= $(MENDER_IMAGE_PREFIX)/gui:$(MENDER_IMAGE_TAG)
# DOCKER_PLATFORM default is defined with respect to the docker server.
# This fixes the default case on darwin where docker is running inside a VM (linux).
DOCKER_PLATFORM ?= $(shell docker system info -f ' \
	{{- .OSType}}/{{- if eq .Architecture "aarch64" -}} \
	  arm64 \
	{{- else -}} \
	  amd64 \
	{{- end -}} \
')

.PHONY: docker
docker: $(platform_targets)
	docker build $(DOCKER_BUILD_ARGS) \
		--build-arg BIN_FILE=$(DOCKER_BIN) \
		--platform $(DOCKER_PLATFORM) \
		-f Dockerfile \
		-t $(DOCKER_TAG) \
		.

build:
	npm ci && \
    npm run build
