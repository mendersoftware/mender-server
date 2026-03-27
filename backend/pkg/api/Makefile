MAKEDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
GIT_DIR := $(shell git rev-parse --show-toplevel)
COMBINED_SPEC := backend/docs/api/dist/openapi.yaml

generate-openapi: generate-openapi-client generate-openapi-mocks
	@echo "Generated $(MAKEDIR)/client"

.PHONY: generate-openapi-client
generate-openapi-client:
	@echo "Regenerating OpenAPI client"
	@rm -rf client
	@docker compose run -u $(shell id -u):$(shell id -g) --rm generate-client

.PHONY: generate-openapi-client-mocks
generate-openapi-mocks:
	@echo "Regenerating mocks"
	@docker compose run -u $(shell id -u):$(shell id -g) --rm generate-mocks
