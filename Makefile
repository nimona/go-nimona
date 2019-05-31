# Go Options
MODULE       := nimona.io
LDFLAGS      := -w -s
BINDIR       := bin
GOBIN        := $(CURDIR)/$(BINDIR)
PATH         := $(GOBIN):$(PATH)
COVEROUT     := ./coverage.out
CLITOOL      := cli-tool
VERSION      := dev # TODO get VERSION from git

# Targets & Sources
MAINBIN := $(BINDIR)/nimona
SOURCES := $(shell find . -name "*.go" -or -name "go.mod" -or -name "go.sum")

# Tools
TOOLS += github.com/cheekybits/genny
TOOLS += github.com/goreleaser/goreleaser
TOOLS += github.com/golangci/golangci-lint/cmd/golangci-lint
TOOLS += github.com/vektra/mockery/cmd/mockery

# Internal tools
TOOLS += nimona.io/tools/objectify
TOOLS += nimona.io/tools/community
TOOLS += nimona.io/tools/vanity

# Enable Go modules
export GO111MODULE=on

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

build: $(MAINBIN)

$(MAINBIN): $(SOURCES)
	$(eval LDFLAGS += -X main.Date=$(shell date +%s))
	$(eval LDFLAGS += -X main.Version=$(VERSION))
	$(eval LDFLAGS += -X main.Commit=$(GIT_SHA))
	cd cmd && \
		CGO_ENABLED=0 go build $(V) \
			-o "../$@" \
			-installsuffix cgo \
			-ldflags '$(LDFLAGS)' \
			./nimona

# Clean up everything
.PHONY: clean
clean:
	rm -f *.cov
	rm -f $(MAINBIN)
	rm -rf $(GOBIN)

# Generate community docs
.PHONY: community-docs
community-docs: nimona.io/tools/community
	$(GOBIN)/community

# Install deps
.PHONY: deps
deps:
	$(info Installing dependencies)
	-go mod $(V) download

# Run go generate
.PHONY: generate
generate: tools
	-GOBIN=$(GOBIN) go generate $(V) ./...

# Run go test
.PHONY: test
test:
	$(eval TAGS += integration)
	BIND_LOCAL=true go test $(V) \
		-tags="$(TAGS)" \
		-count=1 \
		--race \
		-covermode=atomic \
		-coverprofile=$(COVEROUT) \
		./...

# Install tooling
.PHONY: tools
tools: deps $(TOOLS)

# Check tools
.PHONY: $(TOOLS)
$(TOOLS): %:
	cd tools; GOBIN=$(GOBIN) go install $(V) "$*"

# Lint code
.PHONY: lint
lint: github.com/golangci/golangci-lint/cmd/golangci-lint
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
	NIMONA_DAEMON_BOOTSTRAP_ADDRESSES=tcps:localhost:10000 \
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
	NIMONA_DAEMON_BOOTSTRAP_ADDRESSES=tcps:localhost:10000 \
	NIMONA_DAEMON_OBJECT_PATH=.local/peer-two/objects \
	NIMONA_DAEMON_TCP_PORT=10002 \
	NIMONA_DAEMON_HTTP_PORT=10082 \
	NIMONA_API_PORT=10802 \
	$(MAINBIN)
