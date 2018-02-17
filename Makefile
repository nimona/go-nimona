PHONY: test-cover-html
PACKAGES = $(shell find ./ -type d -not -path '*/\.*' | egrep -v 'vendor|examples')

test-cover:
	@echo "mode: count" > coverage.txt
	@rm coverage.txt || true
	@rm coverage-temp.txt || true
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage-temp.txt -covermode=atomic -race $(pkg);\
		tail -n +2 coverage-temp.txt | grep -v _mock >> coverage.txt;)
	@rm coverage-temp.txt

test-cover-html:
	@make test-cover
	go tool cover -html=coverage.txt