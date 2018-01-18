PKG     = github.com/aerogear/mobile-cli
TOP_SRC_DIRS   = pkg
TEST_DIRS     ?= $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
                   -exec dirname {} \\; | sort | uniq")
BIN_DIR := $(GOPATH)/bin
SHELL = /bin/bash

LDFLAGS=-ldflags "-w -s -X main.Version=${TAG}"


setup:
	@go get github.com/kisielk/errcheck
	@go get github.com/goreleaser/goreleaser

build: setup check build_binary

build_binary_linux:
	env GOOS=linux GOARCH=amd64 go build -o mobile ./cmd/mobile

build_binary:
	go build -o mobile ./cmd/mobile

generate:
	./scripts/generate.sh

test-unit:
	@echo Running tests:
	go test -v -race -cover $(UNIT_TEST_FLAGS) \
	  $(addprefix $(PKG)/,$(TEST_DIRS))

.PHONY: errcheck
errcheck:
	@echo errcheck
	@errcheck -ignoretests $$(go list ./... | grep -v mobile-cli/pkg/client)

.PHONY: vet
vet:
	@echo go vet
	@go vet $$(go list ./... | grep -v mobile-cli/pkg/client)
.PHONY: fmt
fmt:
	@echo go fmt
	diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)

.PHONY: check
check: errcheck vet fmt test-unit

.PHONY: integration
integration: build
	go test -v ./integration -args -namespace=`oc project -q` -executable=`pwd`/mobile

.PHONY: release
release: setup
	goreleaser --rm-dist
