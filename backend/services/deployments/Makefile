COMPONENT := deployments

override BUILDTAGS_REQUIRED := nopkcs11

DOCFILES := $(wildcard ../../../backend/docs/api/*deployments.yaml)
MAKEDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

tests/%: ../../../backend/docs/api/%_deployments.yaml
	[ -e $@ ] && rm -r $@; \
	docker run --rm -t -v $(MAKEDIR):/work -v $(MAKEDIR)/../../../backend/docs/api:/api -w /work \
		--ulimit nofile=65536:65536 \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v4.3.1 generate \
		-g python -i /api/$(notdir $<) \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$* && \
		sed -i.gitignore -e 's/"log":false/"log":False/g' tests/*/test/test_*.py
		# see https://github.com/OpenAPITools/openapi-generator/issues/11421
		# there is a bug in the python generator which produces non-python syntax

.PHONY: docs
docs: $(patsubst ../../../backend/docs/api/%_deployments.yaml,tests/%,$(DOCFILES))

test_acceptance_run_deps := docs

include ../Makefile.common
