go 1.14

module nimona.io

require (
	github.com/bmatcuk/doublestar v1.3.2
	github.com/elliotchance/orderedmap v1.3.0
	github.com/geoah/genny v1.0.3
	github.com/geoah/go-queue v2.0.0+incompatible
	github.com/gobwas/glob v0.2.3
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.2
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-sqlite3 v1.14.2
	github.com/mitchellh/mapstructure v1.3.3
	github.com/mr-tron/base58 v1.2.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/teserakt-io/golang-ed25519 v0.0.0-20200315192543-8255be791ce4
	github.com/twmb/murmur3 v1.1.5
	gitlab.com/NebulousLabs/fastrand v0.0.0-20181126182046-603482d69e40 // indirect
	gitlab.com/NebulousLabs/go-upnp v0.0.0-20181011194642-3a71999ed0d3
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/tools v0.0.0-20200318054722-11a475a590ac
	gopkg.in/h2non/gock.v1 v1.0.15
	gopkg.in/yaml.v2 v2.3.0
)

replace (
	nimona.io/cmd => ./cmd
	nimona.io/tools/codegen => ./tools/codegen
	nimona.io/tools/community => ./tools/community
	nimona.io/tools/vanity => ./tools/vanity
)
