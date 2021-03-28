ROOT := $(CURDIR)
SOURCES := $(shell find . -name "*.go" -or -name "go.mod" -or -name "go.sum" \
	-or -name "Makefile")

# Verbose output
ifdef VERBOSE
V = -v
endif

#
# Environment
#

BINDIR := bin
TOOLDIR := $(BINDIR)/tools

# Global environment variables for all targets
SHELL ?= /bin/bash
SHELL := env \
	GO111MODULE=on \
	GOBIN=$(CURDIR)/$(TOOLDIR) \
	CGO_ENABLED=1 \
	GENERATORS=$(CURDIR)/internal/generator \
	PATH='$(CURDIR)/$(BINDIR):$(CURDIR)/$(TOOLDIR):$(PATH)' \
	$(SHELL)

#
# Defaults
#

# Default target
.DEFAULT_GOAL := build

.PHONY: all
all: lint test build

#
# Tools
#

TOOLS += $(TOOLDIR)/gobin
gobin: $(TOOLDIR)/gobin
$(TOOLDIR)/gobin:
	GO111MODULE=off go get -u github.com/myitcv/gobin

# external tool
define tool # 1: binary-name, 2: go-import-path
TOOLS += $(TOOLDIR)/$(1)

.PHONY: $(1)
$(1): $(TOOLDIR)/$(1)

$(TOOLDIR)/$(1): $(TOOLDIR)/gobin Makefile
	gobin $(V) "$(2)"
endef

# internal tool
define inttool # 1: name
TOOLS += $(TOOLDIR)/$(1)

.PHONY: $(1)
$(1): $(TOOLDIR)/$(1)

$(TOOLDIR)/$(1): $(SOURCES)
	cd "tools/$(1)" && go build $(V) -o "$(ROOT)/$(TOOLDIR)/$(1)"
endef

$(eval $(call tool,genny,github.com/geoah/genny@v1.0.3))
$(eval $(call tool,gofumports,mvdan.cc/gofumpt/gofumports))
$(eval $(call tool,golangci-lint,github.com/golangci/golangci-lint/cmd/golangci-lint@v1.38.0))
$(eval $(call tool,golds,go101.org/golds@v0.2.0))
$(eval $(call tool,mockgen,github.com/golang/mock/mockgen@v1.5.0))
$(eval $(call tool,wwhrd,github.com/frapposelli/wwhrd@v0.4.0))
$(eval $(call tool,golines,github.com/segmentio/golines@v0.1.0))

$(eval $(call inttool,codegen))
$(eval $(call inttool,community))

.PHONY: tools
tools: $(TOOLS)

#
# Build
#

MODULE := nimona.io
LDFLAGS := -w -s

VERSION ?= dev
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_SHA ?= $(shell git rev-parse --short HEAD)

CMDDIR := cmd
BINS := $(shell cd "$(CMDDIR)" && \
	find * -maxdepth 0 -type d -exec echo $(BINDIR)/{} \;)

.PHONY: build
build: $(BINS)

$(BINS): $(BINDIR)/%: $(SOURCES)
	mkdir -p "$(BINDIR)"
	cd "$(CMDDIR)/$*" && go build -a $(V) \
		-o "$(ROOT)/$(BINDIR)/$*" \
		-ldflags "$(LDFLAGS) \
			-X $(MODULE)/pkg/version.Date=$(DATE) \
			-X $(MODULE)/pkg/version.Version=$(VERSION) \
			-X $(MODULE)/pkg/version.Commit=$(GIT_SHA)"

#
# Examples
#

EXAMPLEDIR := $(CURDIR)/examples
EXAMPLES := $(shell cd "$(EXAMPLEDIR)" && \
	find * -maxdepth 0 -type d -exec echo $(BINDIR)/examples/{} \;)

.PHONY: fmt
fmt: gofumports golines
	golines -t 4 -m 78 --no-reformat-tags --base-formatter=gofumports -w .

.PHONY: build-examples
build-examples: $(EXAMPLES)

$(EXAMPLES): $(BINDIR)/examples/%: $(SOURCES)
	mkdir -p "$(BINDIR)/examples"
	cd "examples/$*" && go build $(V) -i \
		-o "$(ROOT)/$(BINDIR)/examples/$*" \
		-ldflags '$(LDFLAGS)'

#
# Development
#

# Clean up everything
.PHONY: clean
clean:
	rm -f coverage.out coverage.tmp-*.out
	rm -f $(BINS) $(TOOLS) $(EXAMPLES)
	rm -f ./go.mod.tidy-check ./go.sum.tidy-check
	rm -f $(OUTPUT_DIR)

# Tidy go modules
.PHONY: tidy
tidy:
	$(info Tidying go modules)
	@find . -type f -name "go.sum" -not -path "./vendor/*" -execdir rm {} \;
	@find . -type f -name "go.mod" -not -path "./vendor/*" -execdir go mod tidy \;

# Tidy dependecies and make sure go.mod has been committed
# Currently only checks the main go.mod
.PHONY: check-tidy
check-tidy:
	$(info Checking if go.mod is tidy)
	cp go.mod go.mod.tidy-check
	cp go.sum go.sum.tidy-check
	go mod tidy
	( \
		diff go.mod go.mod.tidy-check && \
		diff go.sum go.sum.tidy-check && \
		rm -f go.mod go.sum && \
		mv go.mod.tidy-check go.mod && \
		mv go.sum.tidy-check go.sum \
	) || ( \
		rm -f go.mod go.sum && \
		mv go.mod.tidy-check go.mod && \
		mv go.sum.tidy-check go.sum; \
		exit 1 \
	)

# Install deps
.PHONY: deps
deps:
	$(info Installing dependencies)
	@go mod download

# Run go generate
.PHONY: generate
generate: codegen genny mockgen
	@go generate $(V) ./...
	@codegen -a .

# Run go test
.PHONY: test
test:
	@LOG_LEVEL=debug NIMONA_UPNP_DISABLE=true \
		go test $(V) -tags="integration" -count=1 --race ./...

# Run go test -bench
.PHONY: benchmark
benchmark:
	@go test -run=^$$ -bench=. ./...

# Run e2e tests
.PHONY: e2e
e2e: clean
	docker build -t nimona:dev .
	cd internal/simulation && E2E_DOCKER_IMAGE=nimona:dev \
		go test $(V) -tags="e2e" -count=1 ./...

# Lint code
.PHONY: lint
lint: golangci-lint
	$(info Running Go linters)
	@GOGC=off golangci-lint $(V) run

# Check licenses
.PHONY: check-licenses
check-licenses: wwhrd
	$(info Checking licenses)
	@go mod vendor
	@wwhrd check

#
# Coverage
#

.PHONY: cover
cover: coverage.out

.PHONY: cover-html
cover-html: coverage.out
	go tool cover -html=coverage.out

.PHONY: cover-func
cover-func: coverage.out
	go tool cover -func=coverage.out

coverage.out: $(SOURCES)
	-@NIMONA_UPNP_DISABLE=true \
		go test $(V) -covermode=count -coverprofile=coverage.tmp-raw.out ./...
	-@cat coverage.tmp-raw.out | \
		grep -Ev '_generated\.go|_mock\.go|.pb.go|/cmd/|/examples/|/playground/' \
			> coverage.tmp-clean.out
	-@(head -n 1 coverage.tmp-clean.out && tail -n +2 coverage.tmp-clean.out | sort) > coverage.out
	cat coverage.out
	-@rm -f coverage.tmp-raw.out coverage.tmp-clean.out

#
# Documentation
#

# Generate community docs
.PHONY: community-docs
community-docs: community
	@community

# Serve docs
.PHONY: docs
docs: golds
	$(info Serving go docs)
	@golds -emphasize-wdpkgs ./...

# Serve site
.PHONY: site
site:
	$(info Serving vuepress)
	@yarn docs:dev


#
# Bindings
#

.PHONY: cross-build
cross-build:
	docker run -t --rm -v "${CURDIR}":/app -w /app \
		-e CGO_ENABLED=1 ${ARGS} \
		docker.elastic.co/beats-dev/golang-crossbuild:1.15.10-main \
		--build-cmd "${CMD}" -p "${GOOS}/${GOARCH}"

BUILD_MODE?=c-shared
OUTPUT_DIR?=output
BINDING_NAME?=libnimona
BINDING_FILE?=$(BINDING_NAME).so
BINDING_ARGS?=
BINDING_OUTPUT?=$(OUTPUT_DIR)/binding

.PHONY: bindings
bindings: bindings-ios bindings-darwin bindings-linux bindings-windows

.PHONY: _bindings
_bindings:
	mkdir -p $(BINDING_OUTPUT)
	go build \
		-ldflags "$(LDFLAGS) \
			-X $(MODULE)/pkg/version.Date=$(DATE) \
			-X $(MODULE)/pkg/version.Version=$(VERSION) \
			-X $(MODULE)/pkg/version.Commit=$(GIT_SHA)" \
	 	-o $(BINDING_OUTPUT)/$(BINDING_FILE) \
		-buildmode=$(BUILD_MODE) \
		$(BINDING_ARGS) \
		binding/*.go

IOS_OUTPUT?=ios
IOS_BINDING_OUTPUT?=$(BINDING_OUTPUT)/$(IOS_OUTPUT)

.PHONY: bindings-ios
bindings-ios: bindings-ios-arm64 bindings-ios-x86-64
	lipo $(IOS_BINDING_OUTPUT)/x86_64.a $(IOS_BINDING_OUTPUT)/arm64.a -create -output $(IOS_BINDING_OUTPUT)/$(BINDING_NAME).a
	cp $(IOS_BINDING_OUTPUT)/arm64.h $(IOS_BINDING_OUTPUT)/$(BINDING_NAME).h
	rm $(IOS_BINDING_OUTPUT)/arm64.h $(IOS_BINDING_OUTPUT)/arm64.a $(IOS_BINDING_OUTPUT)/x86_64.h $(IOS_BINDING_OUTPUT)/x86_64.a

.PHONY: bindings-ios-arm64
bindings-ios-arm64:
	BINDING_FILE=$(IOS_OUTPUT)/arm64.a BUILD_MODE="c-archive" \
	SDK=iphoneos CC=$(PWD)/clangwrap.sh CGO_CFLAGS="-fembed-bitcode" \
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 BINDING_ARGS="-tags ios" \
	make _bindings

.PHONY: bindings-ios-x86-64
bindings-ios-x86-64:
	BINDING_FILE=$(IOS_OUTPUT)/x86_64.a BUILD_MODE="c-archive" \
	SDK=iphonesimulator CC=$(PWD)/clangwrap.sh \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 BINDING_ARGS="-tags ios" \
	make _bindings

DARWIN_OUTPUT?=darwin
DARWIN_BINDING_OUTPUT?=$(BINDING_OUTPUT)/$(DARWIN_OUTPUT)
DARWIN_TARGET?=10.11

.PHONY: bindings-darwin
bindings-darwin: bindings-darwin-x86-64 bindings-darwin-arm64
	lipo \
		$(DARWIN_BINDING_OUTPUT)/x86_64.dylib \
		$(DARWIN_BINDING_OUTPUT)/arm64.dylib \
		-create -output $(DARWIN_BINDING_OUTPUT)/$(BINDING_NAME).dylib
	rm \
		$(DARWIN_BINDING_OUTPUT)/x86_64.dylib \
		$(DARWIN_BINDING_OUTPUT)/arm64.dylib \
		$(DARWIN_BINDING_OUTPUT)/*.h

.PHONY: bindings-darwin-x86-64
bindings-darwin-x86-64:
	BINDING_FILE=$(DARWIN_OUTPUT)/x86_64.dylib \
	BUILD_MODE="c-shared" \
	CGO_CFLAGS=-mmacosx-version-min=$(DARWIN_TARGET) \
	MACOSX_DEPLOYMENT_TARGET=$(DARWIN_TARGET) \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
	make _bindings

.PHONY: bindings-darwin-arm64
bindings-darwin-arm64:
	BINDING_FILE=$(DARWIN_OUTPUT)/arm64.dylib \
	BUILD_MODE="c-shared" \
	CGO_CFLAGS=-mmacosx-version-min=$(DARWIN_TARGET) \
	MACOSX_DEPLOYMENT_TARGET=$(DARWIN_TARGET) \
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
	make _bindings

.PHONY: bindings-darwin-archive-x86-64
bindings-darwin-archive-x86-64:
	BINDING_FILE=$(DARWIN_OUTPUT)/x86_64.a \
	BUILD_MODE="c-archive" \
	CGO_CFLAGS=-mmacosx-version-min=$(DARWIN_TARGET) \
	MACOSX_DEPLOYMENT_TARGET=$(DARWIN_TARGET) \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
	make _bindings

.PHONY: bindings-darwin-archive-arm64
bindings-darwin-archive-arm64:
	BINDING_FILE=$(DARWIN_OUTPUT)/arm64.a \
	BUILD_MODE="c-archive" \
	CGO_CFLAGS=-mmacosx-version-min=$(DARWIN_TARGET) \
	MACOSX_DEPLOYMENT_TARGET=$(DARWIN_TARGET) \
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
	make _bindings

LINUX_OUTPUT?=linux
LINUX_BINDING_NAME?=$(BINDING_NAME).so

.PHONY: bindings-linux
bindings-linux: bindings-linux-386 bindings-linux-amd64 bindings-linux-arm64 bindings-linux-armv7

.PHONY: bindings-linux-386
bindings-linux-386:
	GOOS=linux GOARCH=386 TAG=main \
	ARGS="-e BINDING_FILE=$(LINUX_OUTPUT)/386/$(LINUX_BINDING_NAME)" \
	CMD="make _bindings" \
	make cross-build

.PHONY: bindings-linux-amd64
bindings-linux-amd64:
	GOOS=linux GOARCH=amd64 TAG=main \
	ARGS="-e BINDING_FILE=$(LINUX_OUTPUT)/amd64/$(LINUX_BINDING_NAME)" \
	CMD="make _bindings" \
	make cross-build

.PHONY: bindings-linux-arm64
bindings-linux-arm64:
	GOOS=linux GOARCH=arm64 TAG=arm \
	ARGS="-e BINDING_FILE=$(LINUX_OUTPUT)/arm64/$(LINUX_BINDING_NAME)" \
	CMD="make _bindings" \
	make cross-build

.PHONY: bindings-linux-armv7
bindings-linux-armv7:
	GOOS=linux GOARCH=armv7 TAG=arm \
	ARGS="-e BINDING_FILE=$(LINUX_OUTPUT)/armv7/$(LINUX_BINDING_NAME)" \
	CMD="make _bindings" \
	make cross-build

WINDOWS_OUTPUT?=windows
WINDOWS_BINDING_NAME?=$(BINDING_NAME).dll

.PHONY: bindings-windows
bindings-windows: bindings-windows-386 bindings-windows-amd64

.PHONY: bindings-windows-386
bindings-windows-386:
	GOOS=windows GOARCH=386 \
	ARGS="-e BINDING_FILE=$(WINDOWS_OUTPUT)/386/$(WINDOWS_BINDING_NAME)" \
	TAG=main CMD="make _bindings" \
	make cross-build

.PHONY: bindings-windows-amd64
bindings-windows-amd64:
	GOOS=windows GOARCH=amd64 TAG=main \
	ARGS="-e BINDING_FILE=$(WINDOWS_OUTPUT)/amd64/$(WINDOWS_BINDING_NAME)" \
	CMD="make _bindings" \
	make cross-build