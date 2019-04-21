NMAKE = go run nimona.io/tools/nmake
DAEMN = go run nimona.io/cmd/nimona
CMMNT = go run nimona.io/tools/generators/community

export GO111MODULE=on

.PHONY: build
build:
	@$(NMAKE) build

.PHONY: cleanup
cleanup:
	@$(NMAKE) cleanup

.PHONY: community-docs
community-docs:
	@$(CMMNT)

.PHONY: deps
deps:
	@$(NMAKE) deps

.PHONY: generate
generate:
	@$(NMAKE) generate

.PHONY: install
install:
	@$(NMAKE) install

.PHONY: lint
lint:
	@$(NMAKE) lint

.PHONY: run
run:
	@$(DAEMN)

.PHONY: test
test:
	@$(NMAKE) test

.PHONY: tools
tools:
	@$(NMAKE) tools

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
