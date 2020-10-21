module nimona.io/internal/simulation

go 1.15

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/stretchr/testify v1.6.1
	nimona.io v0.0.0
)

replace nimona.io => ../../

replace github.com/mitchellh/mapstructure => github.com/geoah/mapstructure v1.3.4-rc4
