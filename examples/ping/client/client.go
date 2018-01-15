package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	fabric "github.com/nimona/go-nimona-fabric"
	ping "github.com/nimona/go-nimona-fabric/examples/ping"
)

type Ping struct{}

func (p *Ping) Negotiate(ctx context.Context, conn fabric.Conn) (context.Context, fabric.Conn, error) {
	// close conection when done
	defer conn.Close()

	rp, ok := ctx.Value(fabric.ContextKeyRemoteIdentity).(string)
	if !ok {
		return ctx, nil, errors.New("Could not find remote id")
	}

	// send ping
	fmt.Println("Ping: Writing ping to", rp)
	if err := fabric.WriteToken(conn, []byte("PING")); err != nil {
		fmt.Println("Could not ping", err)
		return ctx, nil, err
	}

	fmt.Println("Ping: Wrote ping")

	// get pong
	fmt.Println("Ping: Reading pong...")
	pong, err := fabric.ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote pong", err)
		return ctx, nil, err
	}

	fmt.Println("Ping: Read pong:", string(pong))
	return ctx, nil, nil
}

func main() {
	crt, err := ping.GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return
	}

	yamux := &fabric.YamuxMiddleware{}
	ident := &fabric.IdentityMiddleware{Local: "CLIENT"}
	security := &fabric.SecMiddleware{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}

	p := &Ping{}

	f := fabric.New()
	f.AddTransport("tcp", fabric.NewTransportTCP())
	f.AddNegotiatorFunc("yamux", yamux.Negotiate)
	f.AddNegotiatorFunc("tls", security.Negotiate)
	f.AddNegotiatorFunc("ping", p.Negotiate)
	f.AddNegotiatorFunc("identity", ident.Negotiate)

	// make a new connection to the the server's ping handler
	if _, _, err := f.DialContext(context.Background(), "tcp:127.0.0.1:3000/tls/yamux/identity/ping"); err != nil {
		fmt.Println("Dial error", err)
		return
	}
}
