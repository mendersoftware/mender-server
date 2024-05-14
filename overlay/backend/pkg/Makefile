LDFLAGS ?= "-s -w"
BUILDFLAGS ?= -trimpath -ldflags $(LDFLAGS)
TESTFLAGS ?=

.PHONY: generate
generate:
	go generate ./...

.PHONY: test
test:
	go test $(BUILDFLAGS) $(TESTFLAGS) ./...
