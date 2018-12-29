NMAKE = go run nimona.io/tools/nmake
DAEMN = go run nimona.io/cmd/nimona

export GO111MODULE=on

.PHONY: build
build:
	@$(NMAKE) build

.PHONY: cleanup
cleanup:
	@$(NMAKE) cleanup

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
