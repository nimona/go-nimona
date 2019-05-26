module nimona.io/tools

go 1.12

require (
	github.com/cheekybits/genny v1.0.0
	github.com/golangci/golangci-lint v1.16.1-0.20190421084833-39f46be46090
	github.com/goreleaser/goreleaser v0.106.0
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/spf13/cobra v0.0.4 // indirect
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5
	golang.org/x/tools v0.0.0-20190525133720-3d17549cdc6b // indirect
	nimona.io/tools/community v0.0.0-00010101000000-000000000000 // indirect
	nimona.io/tools/objectify v0.0.0-00010101000000-000000000000 // indirect
	nimona.io/tools/vanity v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	nimona.io => ../
	nimona.io/cmd/nimona => ../cmd/nimona
	nimona.io/tools/community => ../tools/community
	nimona.io/tools/objectify => ../tools/objectify
	nimona.io/tools/vanity => ../tools/vanity
)

replace github.com/ugorji/go/codec => github.com/ugorji/go v1.1.2
