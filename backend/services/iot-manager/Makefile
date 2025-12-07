COMPONENT := iot-manager

test_acceptance_run_deps := docs

include ../Makefile.common

DOCFILES := $(wildcard docs/*_api.yml)
ROOTDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

tests/%: docs/%.yml
	[ -e $@ ] && rm -r $@; \
	docker run --rm -t -v $(ROOTDIR):$(ROOTDIR) -w $(ROOTDIR) \
		--ulimit=nofile=65535 \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v7.12.0 generate \
		-g python -i $< \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$*

.PHONY: docs
docs: $(patsubst docs/%.yml,tests/%,$(DOCFILES))
