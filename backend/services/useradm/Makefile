COMPONENT := useradm

DOCFILES := $(wildcard ../../../backend/docs/api/useradm/*.yaml)
MAKEDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

tests/%: ../../../backend/docs/api/useradm/%.yaml
	[ -e $@ ] && rm -r $@; \
	docker run --rm -t -v $(MAKEDIR):/work -v $(MAKEDIR)/../../../backend/docs/api:/api -w /work \
		--ulimit nofile=65536:65536 \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v7.19.0 generate \
		-g python -i /api/useradm/$(notdir $<) \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$*

.PHONY: docs
docs: $(patsubst ../../../backend/docs/api/useradm/%.yaml,tests/%,$(DOCFILES))

test_acceptance_run_deps := docs

include ../Makefile.common
