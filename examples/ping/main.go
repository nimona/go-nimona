package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	fabric "github.com/nimona/go-nimona-fabric"
)

func main() {
	ctx := context.Background()

	peerA, err := newPeer("0.0.0.0:3000", "PeerA")
	if err != nil {
		log.Fatal("Could not create peer A", err)
	}

	peerA.Listen(ctx)

	peerB, err := newPeer("0.0.0.0:3001", "PeerB")
	if err != nil {
		log.Fatal("Could not create peer B", err)
	}

	peerB.Listen(ctx)

	// make a new connection to the the server's ping handler
	if _, _, err := peerB.DialContext(context.Background(), "tcp:127.0.0.1:3000/tls/router/ping"); err != nil {
		fmt.Println("Dial error", err)
	}
}

func newPeer(host, peerID string) (*fabric.Fabric, error) {
	crt, err := GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return nil, err
	}

	yamux := &fabric.YamuxMiddleware{}
	router := &fabric.RouterMiddleware{}
	identity := &fabric.IdentityMiddleware{Local: peerID}
	tls := &fabric.SecMiddleware{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}
	ping := &Ping{}

	f := fabric.New(tls, router)
	f.AddTransport(fabric.NewTransportTCP(host))

	f.AddMiddleware(yamux)
	f.AddMiddleware(identity)
	f.AddMiddleware(ping)

	f.AddHandlerFunc("ping", ping.Handle)
	f.AddHandlerFunc("identity/ping", ping.Handle)

	return f, nil
}
