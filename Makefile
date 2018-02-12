.PHONY: test-docker build-docker build-all build-all-latest release test-excoveralls

CC_REPORTER = cc-test-reporter
GO = go
GO_COV = gocov

test:
	$(GO) test -v $($(GO) list | grep -v vendor)

test-cover-report:
	$(GO) test -v -coverprofile=c-temp.out $($(GO) list | grep -v vendor)
	cat c-temp.out | grep -v _mock.go > c.out
	rm c-temp.out
	$(GO_COV) convert c.out | $(GO_COV) report
	rm c.out

test-cover-htlm:
	$(GO) test -v -coverprofile=c-temp.out $($(GO) list | grep -v vendor)
	cat c-temp.out | grep -v _mock.go > c.out
	go tool cover -html=c.out