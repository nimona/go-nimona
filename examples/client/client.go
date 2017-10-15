package main

import (
	"context"

	fabric "github.com/nimona/go-nimona-fabric"
)

func main() {
	f := fabric.New("CLIENT")
	f.DialContext(context.Background(), "tcp:127.0.0.1:3000/nimona:SERVER/ping")
}
