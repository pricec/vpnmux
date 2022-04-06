MKFILE_DIR := $(abspath $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
MOD_PATH   := github.com/pricec/vpnmux

ALPINE_VERSION := 3.14
GO_VERSION     := 1.17.2
GO_IMAGE       := golang:${GO_VERSION}-alpine${ALPINE_VERSION}
ALPINE_IMAGE   := alpine:${ALPINE_VERSION}

GO := docker run -u $$(id -u):$$(id -g) -e GOCACHE=/v/.cache -e CGO_ENABLED=0 -v $(MKFILE_DIR):/v -w /v --rm -t $(GO_IMAGE) go

.PHONY: test
test:
	$(GO) test -v -cover $(MOD_PATH)/...

.PHONY: %
%:
	$(GO) build -o bin/$@ $(MOD_PATH)/cmd/$@
