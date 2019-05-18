NMAKE = go run nimona.io/tools/nmake
DAEMN = go run nimona.io/cmd/nimona
CMMNT = go run nimona.io/tools/generators/community

BIN_NMAKE = ./tools/bin/nmake
BIN_CMMNT = ./tools/bin/community

export GO111MODULE=on
export GOBIN=$(CURDIR)/tools/bin

.PHONY: build
build: tools-check
	@$(BIN_NMAKE) build

.PHONY: clean
clean: tools-check
	@$(BIN_NMAKE) cleanup

.PHONY: community-docs
community-docs: tools-check
	@$(BIN_CMMNT)

.PHONY: deps
deps: tools-check
	@$(BIN_NMAKE) deps

.PHONY: generate
generate: tools-check
	@$(BIN_NMAKE) generate

.PHONY: install
install: tools-check
	@$(BIN_NMAKE) install

.PHONY: lint
lint: tools-check
	$(BIN_NMAKE) lint

.PHONY: run
run: tools-check
	@$(DAEMN)

.PHONY: test
test: tools-check
	@$(BIN_NMAKE) test

.PHONY: tools-check
tools-check: tools/bin/nmake

tools/bin/nmake:
	@$(MAKE) tools

.PHONY: tools
tools:
	cd tools; $(NMAKE) tools

.PHONY: tools-and-lint
tools-and-lint: tools
	-@$(NMAKE) lint

.PHONY: local-bootstrap
local-bootstrap: deps
	-go run nimona.io/cmd/nimona daemon init --data-dir=.local/bootstrap
	-BIND_LOCAL=true go run nimona.io/cmd/nimona daemon start start --data-dir=.local/bootstrap --port=8010 --api-port=8810 --bootstraps=

.PHONY: local-peer-one
local-peer-one: deps
	-go run nimona.io/cmd/nimona daemon init --data-dir=.local/peer-one
	-ENV=dev BIND_LOCAL=true go run nimona.io/cmd/nimona daemon start start --data-dir=.local/peer-one --port=8001 --api-port=8801 --bootstraps=tcps:localhost:8010

.PHONY: local-peer-two
local-peer-two: deps
	-go run nimona.io/cmd/nimona daemon init --data-dir=.local/peer-two
	-BIND_LOCAL=true go run nimona.io/cmd/nimona daemon start start --data-dir=.local/peer-two --port=8002 --api-port=8802 --bootstraps=tcps:localhost:8010
