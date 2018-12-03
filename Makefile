.DEFAULT_GOAL := any

.PHONY: any
any:
	@go run mage.go -version >/dev/null || dep ensure -vendor-only
	@go run mage.go $@
