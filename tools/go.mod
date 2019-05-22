module nimona.io/tools

go 1.12

require (
	github.com/Masterminds/semver v1.4.2
	github.com/cheekybits/genny v1.0.0
	github.com/fatih/color v1.7.0
	github.com/golangci/golangci-lint v1.16.1-0.20190421084833-39f46be46090
	github.com/goreleaser/goreleaser v0.106.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/shurcooL/httpfs v0.0.0-20181222201310-74dc9339e414 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/spf13/cobra v0.0.4
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5
	golang.org/x/tools v0.0.0-20190522164419-521d6ed310dd
	gopkg.in/yaml.v2 v2.2.2
	nimona.io v0.0.0-00010101000000-000000000000
)

replace (
	github.com/ugorji/go/codec => github.com/ugorji/go v1.1.2
	nimona.io => ../
)
