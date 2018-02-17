PHONY: test-cover-html
PACKAGES = $(shell find ./ -type d -not -path '*/\.*' | egrep -v 'vendor|examples')

test-cover:
	@echo "mode: count" > coverage.txt
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage-temp.txt -covermode=count $(pkg);\
		tail -n +2 coverage-temp.txt | grep -v _mock >> coverage.txt;)

test-cover-html:
	@make test-cover
	go tool cover -html=coverage.txt