go 1.14

module nimona.io

require (
	github.com/geoah/genny v1.0.3
	github.com/geoah/go-queue v2.0.0+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/golang/mock v1.4.3
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-multierror v1.0.1-0.20191120192120-72917a1559e1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-sqlite3 v1.13.0
	github.com/mitchellh/mapstructure v1.2.0
	github.com/mr-tron/base58 v1.2.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.7.1
	github.com/skip2/go-qrcode v0.0.0-20200526175731-7ac0b40b2038
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/teserakt-io/golang-ed25519 v0.0.0-20200315192543-8255be791ce4
	gitlab.com/NebulousLabs/fastrand v0.0.0-20181126182046-603482d69e40 // indirect
	gitlab.com/NebulousLabs/go-upnp v0.0.0-20181011194642-3a71999ed0d3
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
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
