PKG     = github.com/aerogear/mobile-cli
TOP_SRC_DIRS   = pkg
TEST_DIRS     ?= $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go \
                   -exec dirname {} \\; | sort | uniq")
BIN_DIR := $(GOPATH)/bin
SHELL = /bin/bash
LDFLAGS=-ldflags "-w -s -X main.Version=${TAG}"


build:
	go build -o mobile ./cmd/mobile

build_linux:
	env GOOS=linux GOARCH=amd64 go build -o mobile ./cmd/mobile

generate:
	vendor/k8s.io/code-generator/generate-internal-groups.sh client github.com/aerogear/mobile-cli/pkg/client/mobile github.com/aerogear/mobile-cli/pkg/apis github.com/aerogear/mobile-cli/pkg/apis  "mobile:v1alpha1"
	vendor/k8s.io/code-generator/generate-internal-groups.sh client github.com/aerogear/mobile-cli/pkg/client/servicecatalog github.com/aerogear/mobile-cli/pkg/apis github.com/aerogear/mobile-cli/pkg/apis  "servicecatalog:v1beta1"

test-unit:
	@echo Running tests:
	go test -v -race -cover $(UNIT_TEST_FLAGS) \
	  $(addprefix $(PKG)/,$(TEST_DIRS))