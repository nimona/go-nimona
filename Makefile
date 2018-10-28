NAME = nimona
BINARY = bin/${NAME}
COMMIT = $(shell git show --format="%h" --no-patch)
VERSION ?= $(shell \
	git describe --exact-match --tags 2>/dev/null \
	|| echo "dev-${COMMIT}")
SOURCES = $(shell find . \
	-name '*.go' \
	-o -name 'Gopkg.lock' \
	-o -name 'Gopkg.toml' \
	-o -name 'Makefile')
PACKAGES = $(shell find . -type d -not -path '*/\.*' | egrep -v 'vendor|examples')

$(BINARY): $(SOURCES)
	go generate ./...
	cd cmd/${NAME} \
		&& go build -tags=release -a -o ../../${BINARY} -ldflags \ "\
			-s -w \
			-X main.version=${VERSION} \
			-X main.commit=${COMMIT} \
			-X main.date=$(shell date +%Y-%m-%dT%T%z)"

.PHONY: build
build: $(BINARY)

.PHONY: run
run: $(BINARY)
	$(BINARY)

.PHONY: clean
clean:
	$(eval BIN_DIR := $(shell dirname ${BINARY}))
	if [ -f ${BINARY} ]; then rm ${BINARY}; fi
	if [ -d ${BIN_DIR} ]; then rmdir ${BIN_DIR}; fi

.PHONY: test-cover
test-cover:
	@echo "mode: atomic" > coverage.txt
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage-temp.txt -covermode=atomic -race $(pkg);\
		tail -n +2 coverage-temp.txt | grep -v _mock >> coverage.txt;)
	@rm coverage-temp.txt

.PHONY: test-cover-html
test-cover-html:
	@make test-cover
	go tool cover -html=coverage.txt
