export BIN ?= ${CURDIR}/bin

GO_BUILD = go build -tags vtable -trimpath -o $(BIN)/kqlite ${CURDIR}/cmd/kqlite

# Make sure BIN is on the PATH
export PATH := $(BIN):$(PATH)

# IMAGE_REGISTRY used to indicate the registery/group for kqlite
IMAGE_REGISTRY ?= epenchev
IMAGE_NAME ?= $(IMAGE_REGISTRY)/kqlite

GO := $(shell which go)

default: clean fmt build

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

#.PHONY: clean build clean-binary default fmt fmt-check mod-tidy test
.PHONY: clean fmt build help

##@ Development

build: kqlite

.PHONY: docker-build
docker-build: ## Build docker image.
	docker build -t kqlite -f Dockerfile .
	docker tag kqlite ${IMAGE_NAME}

.PHONY: docker-push
docker-push: ## Push kqlite image.
	docker push ${IMAGE_NAME}

.PHONY: kqlite
kqlite: ## Build kqlite binary.
	$(GO_BUILD)

.PHONY: example
example: ## Build example client program.
	go build --trimpath -o $(BIN)/example.bin ${CURDIR}/cmd/example

.PHONY: clean
clean: clean-binary

.PHONY: clean-binary
clean-binary:
	go clean
	rm -f $(BIN)/kqlite

.PHONY: fmt
fmt: ## Format source code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: vendor
vendor: ## Runs go mod vendor
	go mod vendor

.PHONY: tidy
tidy: ## Runs go mod tidy
	go mod tidy

.PHONY: fmt-check
fmt-check:
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

.PHONY: test
test: ## Run unit tests.
test: envtest fmt vet
	${GO} test ./... -cover -v -ginkgo.v -coverprofile=coverage.out

.PHONY: test-simple
test-simple: ## Run unit tests without verbose/debug output.
test-simple: envtest fmt vet
	${GO} test ./... -cover

.PHONY: test-package
test-package: ## Run unit tests for specific package.
test-package: envtest fmt vet
	${GO} test -v ./internal/$(package) -ginkgo.v

.PHONY: test-coverage
test-coverage: ## Display test coverage as html output in the browser.
test-coverage: test
	${GO} tool cover -html=coverage.out
	rm -f coverage.out
