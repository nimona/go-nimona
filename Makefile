# Go Options
MODULE       := nimona.io
LDFLAGS      := -w -s
BINDIR       := bin
GOBIN        := $(CURDIR)/$(BINDIR)
PATH         := $(GOBIN):$(PATH)
COVEROUT     := ./coverage.out
CLITOOL      := cli-tool
VERSION      := dev # TODO get VERSION from git
CI           := $(CI)

# Targets & Sources
MAINBIN := $(BINDIR)/nimona
SOURCES := $(shell find . -name "*.go" -or -name "go.mod" -or -name "go.sum")

# Tools
BIN_GOBIN = github.com/myitcv/gobin
TOOLS += github.com/cheekybits/genny
TOOLS += github.com/goreleaser/goreleaser@v0.118.2
TOOLS += github.com/golangci/golangci-lint/cmd/golangci-lint@9161de5
TOOLS += github.com/vektra/mockery/cmd/mockery

# Internal tools
TOOLS_INTERNAL += codegen
TOOLS_INTERNAL += objectify
TOOLS_INTERNAL += community
TOOLS_INTERNAL += vanity
TOOLS_INTERNAL += proxy

# Go env vars
export GO111MODULE=on
export CGO_ENABLED=0

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
			-installsuffix cgo \
			-ldflags '$(LDFLAGS)' \
			./nimona

build-proxy:
	cd tools && \
		go install $(V) \
			-installsuffix cgo \
			./proxy

# Clean up everything
.PHONY: clean
clean:
	rm -f *.cov
	rm -rf $(GOBIN)

# Tidy go modules
.PHONY: tidy
tidy:
	$(info Tidying go modules)
	@find . -type f -name "go.sum" -not -path "./vendor/*" -execdir rm {} \;
	@find . -type f -name "go.mod" -not -path "./vendor/*" -execdir go mod tidy \;

# Generate community docs
.PHONY: community-docs
community-docs: community
	$(GOBIN)/community

# Install deps
.PHONY: deps
deps:
	$(info Installing dependencies)
	-go mod download

# Run go generate
.PHONY: generate
generate: github.com/myitcv/gobin codegen
	-go generate $(V) ./...
	-$(GOBIN)/codegen -a .

# Run go test
.PHONY: test
test:
	$(eval TAGS += integration)
	LOG_LEVEL=debug \
	CGO_ENABLED=1 \
	BIND_LOCAL=true \
	go test $(V) \
		-tags="$(TAGS)" \
		-count=1 \
		--race \
		-covermode=atomic \
		-coverprofile=$(COVEROUT) \
		./...

# Install tools
.PHONY: tools
tools: github.com/myitcv/gobin $(TOOLS) $(TOOLS_INTERNAL)

# Check tools
.PHONY: $(TOOLS)
$(TOOLS): %:
	$(GOBIN)/gobin "$*"

# Check internal tools
.PHONY: $(TOOLS_INTERNAL)
$(TOOLS_INTERNAL): %:
ifndef CI
	cd tools/$*; go install $(V)
endif

# Check gobin
.PHONY: $(BIN_GOBIN)
$(BIN_GOBIN): %:
	GO111MODULE=off go get -u $(BIN_GOBIN)

# Lint code
.PHONY: lint
lint: github.com/myitcv/gobin github.com/golangci/golangci-lint/cmd/golangci-lint@9161de5
	$(info Running Go linters)
	$(GOBIN)/golangci-lint $(V) run

# Local bootstrap
.PHONY: local-bootstrap
local-bootstrap: build
	ENV=dev \
	BIND_LOCAL=true \
	NIMONA_CONFIG=.local/bootstrap/config.json \
	NIMONA_DAEMON_BOOTSTRAP_ADDRESSES= \
	NIMONA_DAEMON_OBJECT_PATH=.local/bootstrap/objects \
	NIMONA_DAEMON_TCP_PORT=10000 \
	NIMONA_DAEMON_HTTP_PORT=10080 \
	NIMONA_API_PORT=10800 \
	$(MAINBIN)

# Local test peer one
.PHONY: local-peer-one
local-peer-one: build
	ENV=dev \
	BIND_LOCAL=true \
	NIMONA_CONFIG=.local/peer-one/config.json \
	NIMONA_DAEMON_BOOTSTRAP_ADDRESSES=tcps:rajaniemi.bootstrap.nimona.io:21013,tcps:liu.bootstrap.nimona.io:21013,tcps:egan.bootstrap.nimona.io:21013 \
	NIMONA_DAEMON_OBJECT_PATH=.local/peer-one/objects \
	NIMONA_DAEMON_TCP_PORT=10001 \
	NIMONA_DAEMON_HTTP_PORT=10081 \
	NIMONA_API_PORT=10801 \
	$(MAINBIN)

# Local test peer two
.PHONY: local-peer-two
local-peer-two: build
	ENV=dev \
	BIND_LOCAL=true \
	NIMONA_CONFIG=.local/peer-two/config.json \
	NIMONA_DAEMON_BOOTSTRAP_ADDRESSES=tcps:rajaniemi.bootstrap.nimona.io:21013,tcps:liu.bootstrap.nimona.io:21013,tcps:egan.bootstrap.nimona.io:21013 \
	NIMONA_DAEMON_OBJECT_PATH=.local/peer-two/objects \
	NIMONA_DAEMON_TCP_PORT=10002 \
	NIMONA_DAEMON_HTTP_PORT=10082 \
	NIMONA_API_PORT=10802 \
	$(MAINBIN)
