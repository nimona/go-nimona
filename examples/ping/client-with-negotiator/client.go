package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	fabric "github.com/nimona/go-nimona-fabric"
	ping "github.com/nimona/go-nimona-fabric/examples/ping"
)

// Ping is our example client, it simply sends a PING string and expects a PONG
type Ping struct{}

// Name of our negotiator
func (p *Ping) Name() string {
	return "ping"
}

// Negotiate will be called after all the other middleware have been processed
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

	p := &Ping{}

	yamux := &fabric.YamuxMiddleware{}
	router := &fabric.RouterMiddleware{}
	identity := &fabric.IdentityMiddleware{Local: "CLIENT"}
	tls := &fabric.SecMiddleware{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}

	f := fabric.New(tls, router)
	f.AddTransport(fabric.NewTransportTCP())
	f.AddMiddleware(yamux)
	f.AddMiddleware(router)
	f.AddMiddleware(identity)
	f.AddMiddleware(tls)
	f.AddNegotiator(p)

	// make a new connection to the the server's ping handler
	if _, _, err := f.DialContext(context.Background(), "tcp:127.0.0.1:3000/tls/router/identity/ping"); err != nil {
		fmt.Println("Dial error", err)
	}
}
