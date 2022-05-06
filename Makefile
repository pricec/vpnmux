MKFILE_DIR := $(abspath $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
MOD_PATH   := github.com/pricec/vpnmux

ALPINE_VERSION  := 3.15
GO_VERSION      := 1.18.1
GO_IMAGE        := golang:${GO_VERSION}
ALPINE_IMAGE    := alpine:${ALPINE_VERSION}
GO_ALPINE_IMAGE := golang:$(GO_VERSION)-alpine$(ALPINE_VERSION)
TEST_IMAGE      := vpnmux-test

COMMON_FLAGS := -e CGO_ENABLED=0 -e GOCACHE=/v/.cache -v $(MKFILE_DIR):/v -w /v --rm -t
GO_FLAGS     := -u $$(id -u):$$(id -g) $(COMMON_FLAGS)
TEST_FLAGS   := --privileged -v /var/run/docker.sock:/var/run/docker.sock $(COMMON_FLAGS)

GO      := docker run $(GO_FLAGS) $(GO_IMAGE) go
GO_TEST := docker run $(TEST_FLAGS) $(TEST_IMAGE) go test

.PHONY: all
all: vpnmux

.PHONY: test-ctr
test-ctr:
	docker build --build-arg GO_ALPINE_IMAGE=$(GO_ALPINE_IMAGE) -t $(TEST_IMAGE) -f Dockerfile.test $(MKFILE_DIR)

.PHONY: test
test: test-ctr
	$(GO_TEST) -v -cover $(MOD_PATH)/...

.PHONY: %
%:
	$(GO) build -o bin/$@ $(MOD_PATH)/cmd/$@
