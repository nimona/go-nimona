# Go Options
MODULE       := nimona.io
LDFLAGS      := -w -s
BINDIR       := bin
GOBIN        := $(CURDIR)/$(BINDIR)
PATH         := $(GOBIN):$(PATH)
CLITOOL      := cli-tool
VERSION      := dev # TODO get VERSION from git
CI           := $(CI)

# Targets & Sources
MAINBIN := $(BINDIR)/nimona
SOURCES := $(shell find . -name "*.go" -or -name "go.mod" -or -name "go.sum")

# Tools
BIN_GOBIN = github.com/myitcv/gobin
TOOLS += github.com/geoah/genny@v1.0.3
TOOLS += github.com/goreleaser/goreleaser@v0.129.0
TOOLS += github.com/golangci/golangci-lint/cmd/golangci-lint@v1.28.1
TOOLS += mvdan.cc/gofumpt/gofumports
TOOLS += github.com/golang/mock/mockgen@v1.4.3
TOOLS += github.com/frapposelli/wwhrd@v0.3.0
TOOLS += github.com/ory/go-acc@v0.2.3

# Internal tools
TOOLS_INTERNAL += codegen
TOOLS_INTERNAL += community
TOOLS_INTERNAL += vanity

# Go env vars
export GO111MODULE=on
export CGO_ENABLED=1

# Go bin for tools
export GOBIN=$(CURDIR)/$(BINDIR)

# Generators path
export GENERATORS=$(CURDIR)/internal/generator

# Default target
.DEFAULT_GOAL := build

# Verbose output
ifdef VERBOSE
V = -v
endif

# Git dependencies
HAS_GIT := $(shell command -v git;)
ifndef HAS_GIT
	$(error Please install git)
endif

# Git Status
GIT_SHA ?= $(shell git rev-parse --short HEAD)

.PHONY: all
all: deps lint test build

build: $(MAINBIN)

$(MAINBIN): $(SOURCES)
	$(eval LDFLAGS += -X $(MODULE)/internal/version.Date=$(shell date +%s))
	$(eval LDFLAGS += -X $(MODULE)/internal/version.Version=$(VERSION))
	$(eval LDFLAGS += -X $(MODULE)/internal/version.Commit=$(GIT_SHA))
	cd cmd && \
		go install $(V) \
			-ldflags '$(LDFLAGS)' \
			./nimona

# Clean up everything
.PHONY: clean
clean:
	@rm -f *.cov
	@rm -rf $(GOBIN)

# Tidy go modules
.PHONY: tidy
tidy:
	$(info Tidying go modules)
	@find . -type f -name "go.sum" -not -path "./vendor/*" -execdir rm {} \;
	@find . -type f -name "go.mod" -not -path "./vendor/*" -execdir go mod tidy \;

# Tidy dependecies and make sure go.mod has been committed
# Currently only checks the main go.mod
.PHONY: tidy
check-tidy:
	$(info Checking if go.mod is tidy)
	@go mod tidy
	@git diff --exit-code go.mod

# Generate community docs
.PHONY: community-docs
community-docs: community
	@$(GOBIN)/community

# Install deps
.PHONY: deps
deps:
	$(info Installing dependencies)
	@go mod download

# Run go generate
.PHONY: generate
generate: github.com/myitcv/gobin codegen
	@go generate $(V) ./...
	@$(GOBIN)/codegen -a .

# Run e2e tests
.PHONY: e2e
e2e: clean
	$(eval TAGS += e2e)
	docker build -t nimona:dev .
	E2E_DOCKER_IMAGE=nimona:dev \
	cd internal/simulation; \
	go test $(V) \
		-tags="$(TAGS)" \
		-count=1 \
		./...

# Run go test
.PHONY: test
test:
	$(eval TAGS += integration)
	@LOG_LEVEL=debug \
	CGO_ENABLED=1 \
	BIND_LOCAL=true \
	go test $(V) \
		-tags="$(TAGS)" \
		-count=1 \
		--race \
		./...

# Code coverage
cover:
	$(info Checking code coverage)
	$(eval TAGS += integration)
	@LOG_LEVEL=debug \
	CGO_ENABLED=1 \
	BIND_LOCAL=true \
	$(BINDIR)/go-acc ./... --output coverage.tmp.out
	@cat coverage.tmp.out | grep -Ev "_generated|_mock|.pb.go" > coverage.out
	@rm -f coverage.tmp.out
	@go tool cover -func=coverage.out

# Install tools
.PHONY: tools
tools: github.com/myitcv/gobin $(TOOLS) $(TOOLS_INTERNAL)

# Check tools
.PHONY: $(TOOLS)
$(TOOLS): %:
	@$(GOBIN)/gobin "$*"

# Check internal tools
.PHONY: $(TOOLS_INTERNAL)
$(TOOLS_INTERNAL): %:
ifndef CI
	$(info Installing tools/$*)
	@cd tools/$*; go install $(V)
endif

# Check gobin
.PHONY: $(BIN_GOBIN)
$(BIN_GOBIN): %:
	@GO111MODULE=off go get -u $(BIN_GOBIN)

# Lint code
.PHONY: lint
lint:
	$(info Running Go linters)
	@$(GOBIN)/golangci-lint $(V) run

# Check licenses
licenses:
	$(info Checking licenses)
	@$(GOCMD) mod vendor
	@$(GOBIN)/wwhrd check

# Local bootstrap
.PHONY: local-bootstrap
local-bootstrap: build
	@ENV=dev \
	BIND_LOCAL=true \
	NIMONA_CONFIG=.local/bootstrap/config.json \
	NIMONA_PEER_BOOTSTRAP_ADDRESSES= \
	NIMONA_PEER_OBJECT_PATH=.local/bootstrap/objects \
	NIMONA_PEER_TCP_PORT=10000 \
	NIMONA_API_PORT=10800 \
	$(MAINBIN)

# Local test peer one
.PHONY: local-peer-one
local-peer-one: build
	@ENV=dev \
	BIND_LOCAL=true \
	LOG_LEVEL=debug \
	DEBUG_BLOCKS=true \
	NIMONA_CONFIG=.local/peer-one/config.json \
	NIMONA_PEER_OBJECT_PATH=.local/peer-one/objects \
	NIMONA_PEER_TCP_PORT=10001 \
	NIMONA_API_PORT=10801 \
	NIMONA_PEER_METRICS=true \
	NIMONA_PEER_DEBUG=true \
	$(MAINBIN)

# Local test peer two
.PHONY: local-peer-two
local-peer-two: build
	@ENV=dev \
	BIND_LOCAL=true \
	LOG_LEVEL=debug \
	DEBUG_BLOCKS=true \
	NIMONA_CONFIG=.local/peer-two/config.json \
	NIMONA_PEER_OBJECT_PATH=.local/peer-two/objects \
	NIMONA_PEER_TCP_PORT=10002 \
	NIMONA_API_PORT=10802 \
	NIMONA_PEER_METRICS=true \
	NIMONA_PEER_DEBUG=true \
	$(MAINBIN)

# Local test peer three
.PHONY: local-peer-three
local-peer-three: build
	@ENV=dev \
	BIND_LOCAL=true \
	LOG_LEVEL=debug \
	DEBUG_BLOCKS=true \
	NIMONA_CONFIG=.local/peer-three/config.json \
	NIMONA_PEER_OBJECT_PATH=.local/peer-three/objects \
	NIMONA_PEER_TCP_PORT=10003 \
	NIMONA_API_PORT=10803 \
	NIMONA_PEER_METRICS=true \
	NIMONA_PEER_DEBUG=true \
	$(MAINBIN)
