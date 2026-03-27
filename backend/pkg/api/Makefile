MAKEDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
GIT_DIR := $(shell git rev-parse --show-toplevel)
COMBINED_SPEC := backend/docs/api/dist/openapi.yaml

.PHONY: generate-openapi-client
generate-openapi-client:
	@echo "Regenerating OpenAPI client"
	@rm -rf client
	@docker compose run -u $(shell id -u):$(shell id -g) --rm generate-client
	@echo "Generated $(MAKEDIR)/client"
