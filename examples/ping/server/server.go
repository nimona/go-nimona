package main

import (
	"context"
	"crypto/tls"
	"fmt"

	fabric "github.com/nimona/go-nimona-fabric"
	ping "github.com/nimona/go-nimona-fabric/examples/ping"
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
	if err := fabric.WriteToken(conn, []byte("PONG")); err != nil {
		fmt.Println("Could not pong", err)
		return nil, err
	}

	fmt.Println("Ping: Wrote pong")

	// return connection as it was
	return nil, nil
}

func main() {
	crt, err := ping.GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return
	}

	f := fabric.New()
	f.AddTransport(fabric.NewTransportTCP())
	f.AddMiddleware(&fabric.IdentityMiddleware{Local: "SERVER"})
	f.AddMiddleware(&fabric.SecMiddleware{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	})
	f.AddHandlerFunc("ping", handler)
	fmt.Println("Listening...")
	f.Listen()
}
