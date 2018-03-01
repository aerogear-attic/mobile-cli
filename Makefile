PKG     = github.com/aerogear/mobile-cli
TOP_SRC_DIRS   = pkg
TEST_DIRS     ?= $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
                   -exec dirname {} \\; | sort | uniq")
BIN_DIR := $(GOPATH)/bin
SHELL = /bin/bash

LDFLAGS=-ldflags "-w -s -X main.Version=${TAG}"

.PHONY: setup
setup:
	@go get github.com/kisielk/errcheck
	glide get github.com/goreleaser/goreleaser\#v0.58.0

.PHONY: build
build: setup check build_binary

.PHONY: coveralls_build
coveralls_build: setup coveralls_check build_binary

.PHONY: build_binary_linux
build_binary_linux:
	env GOOS=linux GOARCH=amd64 go build -o mobile ./cmd/mobile

.PHONY: build_binary
build_binary:
	go build -o mobile ./cmd/mobile

.PHONY: generate
generate:
	./scripts/generate.sh

.PHONY: test-unit
test-unit:
	@echo Running tests:
	go test -v -race -cover $(UNIT_TEST_FLAGS) \
	  $(addprefix $(PKG)/,$(TEST_DIRS))

.PHONY: coveralls_test-unit
coveralls_test-unit:
	@echo Running tests:
	go test -v -race -cover -covermode=atomic -coverprofile=coverage.out $(UNIT_TEST_FLAGS) \
	  $(addprefix $(PKG)/,$(TEST_DIRS))
	goveralls -coverprofile=coverage.out -service=jenkins-ci -repotoken $(COVERALLS_TOKEN)

.PHONY: integration
integration: build_binary
	go test -v ./integration -args -namespace=`oc project -q` -executable=`pwd`/mobile

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

.PHONY: coveralls_check
coveralls_check: errcheck vet fmt coveralls_test-unit

integration_binary: 
	go test -c -v ./integration

.PHONY: integration
integration: build_binary integration_binary
	./integration.test -test.v -namespace=`oc project -q` -executable=`pwd`/mobile

.PHONY: release
release: setup
	goreleaser --rm-dist
