LDFLAGS ?= "-s -w"
BUILDFLAGS ?= -trimpath -ldflags $(LDFLAGS)
TESTFLAGS ?=

.PHONY: test
test:
	go test $(BUILDFLAGS) $(TESTFLAGS) ./...

.PHONY: generate
generate:
	go generate ./...

