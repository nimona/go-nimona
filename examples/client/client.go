package main

import (
	"context"
	"fmt"

	fabric "github.com/nimona/go-nimona-fabric"
)

func main() {
	f := fabric.New()
	f.AddTransport(&fabric.TCP{})
	f.AddMiddleware(&fabric.IdentityMiddleware{
		Local: "CLIENT",
	})
	f.AddMiddleware(&fabric.PingMiddleware{})
	if _, err := f.DialContext(context.Background(), "tcp:127.0.0.1:3000/nimona:SERVER/ping"); err != nil {
		fmt.Println("Dial error", err)
	}
}
