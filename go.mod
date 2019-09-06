go 1.13

module nimona.io

require (
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/cayleygraph/cayley v0.7.5
	github.com/cheekybits/genny v1.0.0
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/dop251/goja v0.0.0-20190814175915-bb8ee191fdd3 // indirect
	github.com/emersion/go-upnp-igd v0.0.0-20170924120501-6fb51d2a2a53
	github.com/go-sourcemap/sourcemap v2.1.2+incompatible // indirect
	github.com/go-test/deep v1.0.3 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/gorilla/websocket v1.4.0
	github.com/jinzhu/copier v0.0.0-20190625015134-976e0346caa8
	github.com/joeycumines/go-dotnotation v0.0.0-20180131115956-2d3612e36c5d
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mr-tron/base58 v1.1.2
	github.com/pkg/errors v0.8.1
	github.com/remyoudompheng/bigfft v0.0.0-20190728182440-6a916e37a237 // indirect
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.3.0
	github.com/tylertreat/BoomFilters v0.0.0-20181028192813-611b3dbe80e8 // indirect
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace (
	nimona.io/cmd/nimona => ./cmd/nimona
	nimona.io/tools/community => ./tools/community
	nimona.io/tools/objectify => ./tools/objectify
	nimona.io/tools/proxy => ./tools/proxy
	nimona.io/tools/vanity => ./tools/vanity
)
