# Go Options
MODULE       := github.com/karhoo/svc-payments
LDFLAGS      := -w -s
BINDIR       := $(CURDIR)/bin
GOBIN        := $(BINDIR)
PATH         := $(GOBIN):$(PATH)
COVEROUT     := ./coverage.out
CLITOOL      := cli-tool
VERSION      := dev # TODO get VERSION from git

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
.DEFAULT_GOAL := all

# Verbose output
ifdef VERBOSE
V = -v
else
.SILENT:
endif

# Git dependencies
HAS_GIT := $(shell command -v git;)
ifndef HAS_GIT
	$(error Please install git)
endif

# Git Status
GIT_SHA ?= $(shell git rev-parse --short HEAD)

# Build nimona binary
.PHONY: build
build: deps
build: LDFLAGS += -X main.Date=$(shell date +%s)
build: LDFLAGS += -X main.Version=$(VERSION)
build: LDFLAGS += -X main.Commit=$(GIT_SHA)
build:
	$(info building binary to $(GOBIN)/nimona)
	@cd cmd; CGO_ENABLED=0 go build -o $(GOBIN)/nimona -installsuffix cgo -ldflags '$(LDFLAGS)' ./nimona

# Clean up everything
.PHONY: clean
clean:
	@rm -f *.cov
	@rm -rf $(GOBIN)

# Generate community docs
.PHONY: community-docs
community-docs: tools-check
	@$(GOBIN)/community

# Install deps
.PHONY: deps
deps:
	$(info Installing dependencies)
	-go mod download

# Run go generate
.PHONY: generate
generate: tools
	-GOBIN=$(GOBIN) go generate $(V) ./...

# Run go test
.PHONY: test
test: TAGS += integration
test: tools
	-go test $(V) -tags="$(TAGS)" -count=1 --race -covermode=atomic -coverprofile=$(COVEROUT) ./...

# Install tooling
.PHONY: tools
tools: deps $(TOOLS)

# Check tools
.PHONY: $(TOOLS)
$(TOOLS): %:
	cd tools; GOBIN=$(GOBIN) go install -v $*

# Lint code
.PHONY: lint
lint: tools
	$(info Running Go linters)
	@$(GOBIN)/golangci-lint $(V) run

# Local bootstrap
.PHONY: local-bootstrap
local-bootstrap: build
	-$(GOBIN)/nimona daemon init --data-dir=.local/bootstrap
	-BIND_LOCAL=true $(GOBIN)/nimona daemon start start --data-dir=.local/bootstrap --port=8010 --api-port=8810 --bootstraps=

# Local test peer one
.PHONY: local-peer-one
local-peer-one: build
	-$(GOBIN)/nimona daemon init --data-dir=.local/peer-one
	-ENV=dev BIND_LOCAL=true $(GOBIN)/nimona daemon start start --data-dir=.local/peer-one --port=8001 --api-port=8801 --bootstraps=tcps:localhost:8010

# Local test peer two
.PHONY: local-peer-two
local-peer-two: build
	-$(GOBIN)/nimona daemon init --data-dir=.local/peer-two
	-BIND_LOCAL=true $(GOBIN)/nimona daemon start start --data-dir=.local/peer-two --port=8002 --api-port=8802 --bootstraps=tcps:localhost:8010
