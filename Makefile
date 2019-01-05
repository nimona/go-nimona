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

.PHONY: community
community:
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
