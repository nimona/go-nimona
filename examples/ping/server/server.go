package main

import (
	"context"
	"fmt"

	fabric "github.com/nimona/go-nimona-fabric"
)

func handler(ctx context.Context, conn fabric.Conn) (fabric.Conn, error) {
	// close connection when done
	defer conn.Close()

	// client pings
	fmt.Println("Ping: Reading ping...")
	ping, err := fabric.ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote ping", err)
		return nil, err
	}

	fmt.Println("Ping: Read ping:", string(ping))

	// we pong back
	fmt.Println("Ping: Writing pong...")
	if err := fabric.WriteToken(conn, []byte("pong")); err != nil {
		fmt.Println("Could not pong", err)
		return nil, err
	}

	fmt.Println("Ping: Wrote pong")

	// return connection as it was
	return nil, nil
}

func main() {
	f := fabric.New()
	f.AddTransport(&fabric.TCP{})
	f.AddMiddleware(&fabric.IdentityMiddleware{Local: "SERVER"})
	f.AddHandlerFunc("ping", handler)
	f.Listen()
}
