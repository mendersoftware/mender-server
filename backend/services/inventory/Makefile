COMPONENT := inventory

test_acceptance_run_deps := docs

include ../Makefile.common

DOCFILES := $(wildcard ../../docs/api/inventory*.yaml)
DOCTARGETS := $(patsubst ../../docs/api/%.yaml,tests/openapi_client/%,$(DOCFILES))
MAKEDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

tests/.openapi-generator/FILES: ../../docs/api/inventory.openapi.yaml
	docker run --rm -t -v $(GIT_ROOT):$(GIT_ROOT) -w $(MAKEDIR) \
		--ulimit nofile=65536:65536 \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v7.13.0 generate \
		-g python -i $< \
		-c tests/openapi_client/.openapi-generator.yaml \
		-o tests \
		--additional-properties=packageName=openapi_client

.PHONY: docs
docs: tests/.openapi-generator/FILES
