COMPONENT := reporting

DOCFILES := $(wildcard docs/*_api.yml)
ROOTDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

tests/%: docs/%.yml
	[ -e $@ ] && rm -r $@; \
	docker run --rm -t -v $(ROOTDIR):$(ROOTDIR) -w $(ROOTDIR) \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v7.19.0 generate \
		-g python -i $< \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$*

.PHONY: docs
docs: $(patsubst docs/%.yml,tests/%,$(DOCFILES))

.PHONY: test-acceptance-run test-acceptance
_no_include_acceptance_tests := true

include ../Makefile.common
