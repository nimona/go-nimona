package main

import (
	"context"
	"crypto/tls"
	"fmt"

	fabric "github.com/nimona/go-nimona-fabric"
	ping "github.com/nimona/go-nimona-fabric/examples/ping"
)

type Ping struct{}

func (p *Ping) Negotiate(ctx context.Context, conn fabric.Conn) (err error) {
	// close conection when done
	defer conn.Close()

	// send ping
	fmt.Println("Ping: Writing ping...")
	if err := fabric.WriteToken(conn, []byte("PING")); err != nil {
		fmt.Println("Could not ping", err)
		return err
	}

	fmt.Println("Ping: Wrote ping")

	// get pong
	fmt.Println("Ping: Reading pong...")
	pong, err := fabric.ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote pong", err)
		return err
	}

	fmt.Println("Ping: Read pong:", string(pong))
	return nil
}

func main() {
	crt, err := ping.GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return
	}

	yamux := &fabric.YamuxMiddleware{}
	// ident := &fabric.IdentityMiddleware{Local: "CLIENT"}
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
	// f.AddNegotiatorFunc("identity", ident.Negotiate)

	// make a new connection to the the server's ping handler
	if _, err := f.DialContext(context.Background(), "tcp:127.0.0.1:3000/tls/ping"); err != nil {
		fmt.Println("Dial error", err)
		return
	}
}
