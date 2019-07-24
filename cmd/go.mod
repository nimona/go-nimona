module nimona.io/cmd

go 1.12

replace nimona.io => ../

replace github.com/ugorji/go/codec => github.com/ugorji/go v1.1.7

require (
	github.com/caarlos0/env/v6 v6.0.0
	github.com/cayleygraph/cayley v0.7.5
	nimona.io v0.0.0-00010101000000-000000000000
)
