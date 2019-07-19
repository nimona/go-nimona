module nimona.io/tools

go 1.12

require (
	github.com/cheekybits/genny v1.0.0
	github.com/golangci/golangci-lint v1.17.1
	github.com/goreleaser/goreleaser v0.112.2
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sirupsen/logrus v1.2.0 // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5
	nimona.io/tools/community v0.0.0-00010101000000-000000000000 // indirect
	nimona.io/tools/objectify v0.0.0-00010101000000-000000000000 // indirect
	nimona.io/tools/proxy v0.0.0-00010101000000-000000000000 // indirect
	nimona.io/tools/vanity v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	nimona.io => ../
	nimona.io/cmd/nimona => ../cmd/nimona
	nimona.io/tools/community => ../tools/community
	nimona.io/tools/objectify => ../tools/objectify
	nimona.io/tools/proxy => ../tools/proxy
	nimona.io/tools/vanity => ../tools/vanity
)

replace github.com/ugorji/go/codec => github.com/ugorji/go v1.1.2
