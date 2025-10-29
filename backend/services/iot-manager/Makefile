COMPONENT := iot-manager

test_acceptance_run_deps := docs

include ../Makefile.common

DOCFILES := $(wildcard ../../../backend/docs/api/iot-manager/*.yaml)
ROOTDIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

tests/%: ../../../backend/docs/api/iot-manager/%.yaml
	[ -e $@ ] && rm -r $@; \
	docker run --rm -t -v $(ROOTDIR):$(ROOTDIR) -v $(ROOTDIR)/../../../backend/docs/api:/api -w $(ROOTDIR) \
		--ulimit=nofile=65535 \
		-u $(shell id -u):$(shell id -g) \
		openapitools/openapi-generator-cli:v7.19.0 generate \
		-g python -i /api/iot-manager/$(notdir $<) \
		-c tests/.openapi-generator.yml \
		-o $(dir $@) \
		--additional-properties=packageName=$*

.PHONY: docs
docs: $(patsubst ../../../backend/docs/api/iot-manager/%.yaml,tests/%,$(DOCFILES))
