go 1.13

module nimona.io

require (
	github.com/cheekybits/genny v1.0.0
	github.com/emersion/go-upnp-igd v0.0.0-20170924120501-6fb51d2a2a53
	github.com/geoah/go-queue v2.0.0+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/google/go-cmp v0.4.0
	github.com/gorilla/websocket v1.4.1
	github.com/hashicorp/go-multierror v1.0.1-0.20191120192120-72917a1559e1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-sqlite3 v1.13.0
	github.com/mr-tron/base58 v1.1.3
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.5.1
	github.com/vburenin/nsync v0.0.0-20160822015540-9a75d1c80410
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/h2non/gock.v1 v1.0.15
	gopkg.in/src-d/go-git.v4 v4.13.1
)

replace (
	nimona.io/cmd => ./cmd
	nimona.io/tools/codegen => ./tools/codegen
	nimona.io/tools/community => ./tools/community
	nimona.io/tools/vanity => ./tools/vanity
)
