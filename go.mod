go 1.13

module nimona.io

require (
	github.com/cheekybits/genny v1.0.0
	github.com/emersion/go-upnp-igd v0.0.0-20170924120501-6fb51d2a2a53
	github.com/go-test/deep v1.0.3 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gorilla/websocket v1.4.1
	github.com/jinzhu/copier v0.0.0-20190625015134-976e0346caa8
	github.com/joeycumines/go-dotnotation v0.0.0-20180131115956-2d3612e36c5d
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mr-tron/base58 v1.1.2
	github.com/pkg/errors v0.8.1
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297
	nimona.io/tools/codegen v0.0.0-00010101000000-000000000000 // indirect
	nimona.io/tools/community v0.0.0-00010101000000-000000000000 // indirect
	nimona.io/tools/objectify v0.0.0-00010101000000-000000000000 // indirect
	nimona.io/tools/proxy v0.0.0-00010101000000-000000000000 // indirect
	nimona.io/tools/vanity v0.0.0-00010101000000-000000000000 // indirect
)

replace (
	nimona.io/tools/codegen => ./tools/codegen
	nimona.io/tools/community => ./tools/community
	nimona.io/tools/objectify => ./tools/objectify
	nimona.io/tools/proxy => ./tools/proxy
	nimona.io/tools/vanity => ./tools/vanity
)
