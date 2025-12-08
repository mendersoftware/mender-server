COMPONENT := deployments

override BUILDTAGS_REQUIRED := nopkcs11

DOCFILES := $(wildcard ../../../backend/docs/api/deployments/*.yaml)
MAKEDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

tests/%: ../../../backend/docs/api/deployments/%.yaml
	[ -e $@ ] && rm -r $@; \
	docker run --rm -t -v $(MAKEDIR):/work -v $(MAKEDIR)/../../../backend/docs/api:/api -w /work \
		--ulimit nofile=65536:65536 \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v4.3.1 generate \
		-g python -i /api/deployments/$(notdir $<) \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$* && \
		sed -i.gitignore -e 's/"log":false/"log":False/g' tests/*/test/test_*.py && \
		sed -i.gitignore -e 's/response\.getheaders()/response\.headers/g' -e 's/urllib3_response\.getheaders()/urllib3_response\.headers/g' -e 's/urllib3_response\.getheader(/urllib3_response\.headers\.get(/g' -e 's/response\.getheader(/response\.headers\.get(/g' tests/*/*.py
		# see https://github.com/OpenAPITools/openapi-generator/issues/11421
		# there is a bug in the python generator which produces non-python syntax
		# see the urllib depcrecating and removing the getheader call
		# https://urllib3.readthedocs.io/en/stable/v2-migration-guide.html#what-are-the-important-changes
		# this is hacky and a bit dirty but it was the fastest track to unlock

.PHONY: docs
docs: $(patsubst ../../../backend/docs/api/deployments/%.yaml,tests/%,$(DOCFILES))

test_acceptance_run_deps := docs

include ../Makefile.common
