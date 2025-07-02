LDFLAGS ?= "-s -w"
BUILDFLAGS ?= -trimpath -ldflags $(LDFLAGS)
TESTFLAGS :=
COMPONENT := pkg

# FIXME: The hardcoded -p 1 is required because of resource conflict with
#        mongodb. The dbtest package `Wipe` drops all databases, affecting
#        test binaries that are running in parallel.

.PHONY: test
test:
	go test $(BUILDFLAGS) $(TESTFLAGS) -p 1 ./...

.PHONY: generate
generate:
	go generate ./...

