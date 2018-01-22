package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	fabric "github.com/nimona/go-nimona-fabric"
	ping "github.com/nimona/go-nimona-fabric/examples/ping"
)

func handler(ctx context.Context, c fabric.Conn) (context.Context, fabric.Conn, error) {
	fmt.Println("Going through ping")

	// close connection when done
	defer c.Close()

	rp, ok := ctx.Value(fabric.ContextKeyRemoteIdentity).(string)
	if !ok {
		return nil, nil, errors.New("Could not find remote id")
	}

	// client pings
	fmt.Println("Ping: Reading ping from", rp)
	ping, err := fabric.ReadToken(c)
	if err != nil {
		fmt.Println("Could not read remote ping", err)
		return nil, nil, err
	}

	fmt.Println("Ping: Read ping:", string(ping))

	// we pong back
	fmt.Println("Ping: Writing pong...")
	if err := fabric.WriteToken(c, []byte("PONG")); err != nil {
		fmt.Println("Could not pong", err)
		return nil, nil, err
	}

	fmt.Println("Ping: Wrote pong")

	// return connection as it was
	return nil, nil, nil
}

func main() {
	crt, err := ping.GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return
	}

	//
	// tls/select/relay/identity/ping
	// tls/select/relay/identity/pong
	//

	yamux := &fabric.YamuxMiddleware{}
	router := &fabric.RouterMiddleware{}
	identity := &fabric.IdentityMiddleware{Local: "SERVER"}
	tls := &fabric.SecMiddleware{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}

	f := fabric.New(tls, router)
	f.AddTransport(fabric.NewTransportTCP("0.0.0.0:3000"))
	f.AddMiddleware(yamux)
	f.AddMiddleware(router)
	f.AddMiddleware(identity)
	f.AddMiddleware(tls)
	f.AddHandlerFunc("ping", handler)

	fmt.Println("Listening...")

	f.Listen()
}
