module nimona.io/tools

go 1.12

require (
	github.com/cheekybits/genny v1.0.0
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golangci/golangci-lint v1.17.1
	github.com/goreleaser/goreleaser v0.113.0
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5
	golang.org/x/net v0.0.0-20190726094755-ca1201d0de80 // indirect
	golang.org/x/sys v0.0.0-20190726094755-fc99dfbffb4e // indirect
	golang.org/x/tools v0.0.0-20190726094755-2e34cfcb95cb // indirect
	google.golang.org/appengine v1.6.1 // indirect
)

replace (
	nimona.io => ../
	nimona.io/cmd/nimona => ../cmd/nimona
	nimona.io/tools/community => ../tools/community
	nimona.io/tools/objectify => ../tools/objectify
	nimona.io/tools/proxy => ../tools/proxy
	nimona.io/tools/vanity => ../tools/vanity
)

replace github.com/ugorji/go/codec => github.com/ugorji/go v1.1.7
