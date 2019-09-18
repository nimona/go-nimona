go 1.12

module nimona.io

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/cheekybits/genny v1.0.0
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/emersion/go-upnp-igd v0.0.0-20170924120501-6fb51d2a2a53
	github.com/gobwas/glob v0.2.3
	github.com/gorilla/websocket v1.4.1
	github.com/mr-tron/base58 v1.1.2
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297
)

replace (
	nimona.io/cmd => ./cmd
	nimona.io/tools/codegen => ./tools/codegen
	nimona.io/tools/community => ./tools/community
	nimona.io/tools/proxy => ./tools/proxy
	nimona.io/tools/vanity => ./tools/vanity
)
