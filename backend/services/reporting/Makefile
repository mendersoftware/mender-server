COMPONENT := reporting

DOCFILES := $(wildcard docs/*_api.yml)
ROOTDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

tests/%: docs/%.yml
	[ -e $@ ] && rm -r $@; \
	docker run --rm -t -v $(ROOTDIR):$(ROOTDIR) -w $(ROOTDIR) \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v5.3.1 generate \
		-g python -i $< \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$* && \
		sed -i.gitignore -e 's/response\.getheaders()/response\.headers/g' -e 's/urllib3_response\.getheaders()/urllib3_response\.headers/g' -e 's/urllib3_response\.getheader(/urllib3_response\.headers\.get(/g' -e 's/response\.getheader(/response\.headers\.get(/g' tests/*/*.py
		# see the urllib depcrecating and removing the getheader call
		# https://urllib3.readthedocs.io/en/stable/v2-migration-guide.html#what-are-the-important-changes
		# this is hacky and a bit dirty but it was the fastest track to unlock

.PHONY: docs
docs: $(patsubst docs/%.yml,tests/%,$(DOCFILES))

.PHONY: test-acceptance-run test-acceptance
_no_include_acceptance_tests := true

include ../Makefile.common
