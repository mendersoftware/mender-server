MAKEDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
GIT_DIR := $(shell git rev-parse --show-toplevel)
COMBINED_SPEC := backend/docs/api/dist/openapi.yaml
OPENAPI_GENERATOR_VERSION ?= v7.20.0
MOCKERY_VERSION ?= v3.7.0

generate-openapi: generate-openapi-client generate-openapi-mocks
	@echo "Generated $(MAKEDIR)/client"

.PHONY: generate-openapi-client
generate-openapi-client:
	@echo "Regenerating OpenAPI client"
	@rm -rf client
	@docker run -u $(shell id -u):$(shell id -g) --rm \
		-v $(GIT_DIR)/backend:/backend \
		-w /backend/pkg/api \
		openapitools/openapi-generator-cli:$(OPENAPI_GENERATOR_VERSION) \
		generate \
		-i=/backend/docs/api/dist/openapi.yaml \
		-o=client \
		-g=go \
		-c=.openapi-generator.yaml

.PHONY: generate-openapi-client-mocks
generate-openapi-mocks:
	@echo "Regenerating mocks"
	@docker run -u $(shell id -u):$(shell id -g) --rm \
		-v $(GIT_DIR)/backend:/backend \
		-w /backend/pkg/api \
		vektra/mockery:$(MOCKERY_VERSION) \
			--config /backend/pkg/api/.mockery.yml
