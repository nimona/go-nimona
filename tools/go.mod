module nimona.io/tools

go 1.12

require (
	github.com/cheekybits/genny v1.0.0
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golangci/golangci-lint v1.17.1
	github.com/goreleaser/goreleaser v0.117.1
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	golang.org/x/sys v0.0.0-20190904154756-749cb33beabd // indirect
	golang.org/x/tools v0.0.0-20190905235650-93dcc2f048f5 // indirect
	google.golang.org/appengine v1.6.2 // indirect
)

replace (
	nimona.io => ../
	nimona.io/cmd/nimona => ../cmd/nimona
	nimona.io/tools/community => ../tools/community
	nimona.io/tools/objectify => ../tools/objectify
	nimona.io/tools/proxy => ../tools/proxy
	nimona.io/tools/vanity => ../tools/vanity
)
