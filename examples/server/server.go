package main

import fabric "github.com/nimona/go-nimona-fabric"

func main() {
	f := fabric.New()
	f.AddTransport(&fabric.TCP{})
	f.AddMiddleware(&fabric.IdentityMiddleware{
		Local: "SERVER",
	})
	f.AddMiddleware(&fabric.PingMiddleware{})
	f.Listen()
}
