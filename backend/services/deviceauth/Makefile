COMPONENT := deviceauth

override TESTFLAGS := $(TESTFLAGS) -p 1

DOCFILES := $(wildcard docs/*_api.yml)
MAKEDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

tests/%: docs/%.yml
	[ -e $@ ] && rm -r $@; \
	docker run --rm -t -v $(MAKEDIR):/work -w /work \
		--ulimit nofile=65536:65536 \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v7.17.0 generate \
		-g python -i $< \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$*

.PHONY: docs
docs: $(patsubst docs/%.yml,tests/%,$(DOCFILES))

test_acceptance_run_deps := docs

include ../Makefile.common
