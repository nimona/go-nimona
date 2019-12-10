go 1.13

module nimona.io

require (
	github.com/caarlos0/env/v6 v6.1.0
	github.com/cheekybits/genny v1.0.0
	github.com/emersion/go-upnp-igd v0.0.0-20170924120501-6fb51d2a2a53
	github.com/geoah/go-queue v2.0.0+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/google/go-cmp v0.3.1
	github.com/gorilla/websocket v1.4.1
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-sqlite3 v1.13.0
	github.com/mr-tron/base58 v1.1.3
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tsdtsdtsd/identicon v0.0.0-20190130180410-ca6dab10d534
	github.com/tyler-smith/go-bip39 v1.0.2
	golang.org/x/crypto v0.0.0-20191219195013-becbf705a915 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/h2non/gock.v1 v1.0.15
)

replace (
	nimona.io/cmd => ./cmd
	nimona.io/tools/codegen => ./tools/codegen
	nimona.io/tools/community => ./tools/community
	nimona.io/tools/vanity => ./tools/vanity
)
