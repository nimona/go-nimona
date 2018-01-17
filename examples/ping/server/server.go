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

	yamux := &fabric.YamuxMiddleware{}
	// nselect := &fabric.SelectMiddleware{
	// 	Handlers: map[string]fabric.HandlerFunc{
	// 		"ping": handler,
	// 	},
	// }
	identity := &fabric.IdentityMiddleware{Local: "SERVER"}
	security := &fabric.SecMiddleware{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}

	f := fabric.New()
	f.AddTransport("tcp", fabric.NewTransportTCP())
	f.AddHandler("ping", handler)
	f.AddHandler("tls", security.Handle)
	f.AddHandler("yamux", yamux.Handle)
	f.AddHandler("identity", identity.Handle)
	fmt.Println("Listening...")
	f.Listen()
}
