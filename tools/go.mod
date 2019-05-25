module nimona.io/tools

go 1.12

require (
	github.com/cheekybits/genny v1.0.0
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/golangci/golangci-lint v1.16.1-0.20190421084833-39f46be46090
	github.com/goreleaser/goreleaser v0.106.0
	github.com/mattn/go-isatty v0.0.8 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/spf13/cobra v0.0.4 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5
	golang.org/x/net v0.0.0-20190522164419-f3200d17e092 // indirect
	golang.org/x/sys v0.0.0-20190525133720-dbbf3f1254d4 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20190525133720-3d17549cdc6b // indirect
	google.golang.org/appengine v1.6.0 // indirect
)

replace (
	nimona.io => ../
	nimona.io/cmd/nimona => ../cmd/nimona
	nimona.io/tools/community => ../tools/community
	nimona.io/tools/objectify => ../tools/objectify
	nimona.io/tools/vanity => ../tools/vanity
)

replace github.com/ugorji/go/codec => github.com/ugorji/go v1.1.2
