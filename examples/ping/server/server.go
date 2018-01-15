package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	fabric "github.com/nimona/go-nimona-fabric"
	ping "github.com/nimona/go-nimona-fabric/examples/ping"
)

func handler(ctx context.Context, conn fabric.Conn) error {
	fmt.Println("Going through ping")

	// close connection when done
	defer conn.Close()

	rp, ok := ctx.Value(fabric.ContextKeyRemoteIdentity).(string)
	if !ok {
		return errors.New("Could not find remote id")
	}

	// client pings
	fmt.Println("Ping: Reading ping from", rp)
	ping, err := fabric.ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote ping", err)
		return err
	}

	fmt.Println("Ping: Read ping:", string(ping))

	// we pong back
	fmt.Println("Ping: Writing pong...")
	if err := fabric.WriteToken(conn, []byte("PONG")); err != nil {
		fmt.Println("Could not pong", err)
		return err
	}

	fmt.Println("Ping: Wrote pong")

	// return connection as it was
	return nil
}

func main() {
	crt, err := ping.GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return
	}

	yamux := &fabric.YamuxMiddleware{}
	// nselect := &fabric.NimonaMiddleware{
	// 	Handlers: map[string]fabric.HandlerFunc{
	// 		"ping": handler,
	// 	},
	// }
	ident := &fabric.IdentityMiddleware{Local: "SERVER"}
	security := &fabric.SecMiddleware{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}

	f := fabric.New()
	f.AddTransport("tcp", fabric.NewTransportTCP())
	f.AddHandlerFunc("tls/ping", fabric.BuildChain(handler, security.HandlerWrapper))
	f.AddHandlerFunc("tls/yamux/ping", fabric.BuildChain(handler, security.HandlerWrapper, yamux.HandlerWrapper))
	f.AddHandlerFunc("tls/yamux/identity/ping", fabric.BuildChain(handler, security.HandlerWrapper, yamux.HandlerWrapper, ident.HandlerWrapper))
	fmt.Println("Listening...")
	f.Listen()
}
