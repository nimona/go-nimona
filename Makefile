# Go Options
MODULE       := github.com/karhoo/svc-payments
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
	-$(MAINBIN) daemon init --data-dir=.local/bootstrap
	BIND_LOCAL=true $(MAINBIN) daemon start start \
		--data-dir=.local/bootstrap \
		--port=8010 \
		--api-port=8810 \
		--bootstraps=

# Local test peer one
.PHONY: local-peer-one
local-peer-one: build
	-$(MAINBIN) daemon init --data-dir=.local/peer-one
	ENV=dev BIND_LOCAL=true $(MAINBIN) daemon start start \
		--data-dir=.local/peer-one \
		--port=8001 \
		--api-port=8801 \
		--bootstraps=tcps:localhost:8010

# Local test peer two
.PHONY: local-peer-two
local-peer-two: build
	-$(MAINBIN) daemon init --data-dir=.local/peer-two
	BIND_LOCAL=true $(MAINBIN) daemon start start \
		--data-dir=.local/peer-two \
		--port=8002 \
		--api-port=8802 \
		--bootstraps=tcps:localhost:8010
