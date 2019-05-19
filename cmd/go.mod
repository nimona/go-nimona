module nimona.io/cmd

go 1.12

replace (
	github.com/ugorji/go/codec => github.com/ugorji/go v1.1.2
	nimona.io => ../
)

require (
	github.com/cayleygraph/cayley v0.7.5
	github.com/gin-gonic/gin v1.4.0
	github.com/gorilla/websocket v1.4.0
	github.com/pkg/errors v0.8.1
	github.com/pkg/profile v1.3.0
	github.com/spf13/cobra v0.0.4
	github.com/spf13/viper v1.3.2
	gopkg.in/resty.v1 v1.12.0
	nimona.io v0.0.0-00010101000000-000000000000
)
