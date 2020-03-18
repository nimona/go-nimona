go 1.14

module nimona.io

require (
	github.com/cheekybits/genny v1.0.0
	github.com/geoah/go-queue v2.0.0+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/go-multierror v1.0.1-0.20191120192120-72917a1559e1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-sqlite3 v1.13.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mr-tron/base58 v1.1.3
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/vburenin/nsync v0.0.0-20160822015540-9a75d1c80410
	gitlab.com/NebulousLabs/fastrand v0.0.0-20181126182046-603482d69e40 // indirect
	gitlab.com/NebulousLabs/go-upnp v0.0.0-20181011194642-3a71999ed0d3
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/h2non/gock.v1 v1.0.15
)

// https://github.com/spaolacci/murmur3/issues/29
// https://github.com/spaolacci/murmur3/pull/30
replace github.com/spaolacci/murmur3 => github.com/calmh/murmur3 v1.1.1-0.20200226160057-74e9af8f47ac

replace (
	nimona.io/cmd => ./cmd
	nimona.io/tools/codegen => ./tools/codegen
	nimona.io/tools/community => ./tools/community
	nimona.io/tools/vanity => ./tools/vanity
)
