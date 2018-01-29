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

	peerA, err := newPeer(3001, 3002, "PeerA")
	if err != nil {
		log.Fatal("Could not create peer A", err)
	}

	peerA.Listen(ctx)

	peerB, err := newPeer(4001, 4002, "PeerB")
	if err != nil {
		log.Fatal("Could not create peer B", err)
	}

	peerB.Listen(ctx)

	// ping through ws
	if _, _, err := peerB.DialContext(context.Background(), "ws:127.0.0.1:3002/tls/router/ping"); err != nil {
		fmt.Println("Dial error", err)
	}

	// ping through tcp with identity
	if _, _, err := peerB.DialContext(context.Background(), "tcp:127.0.0.1:3001/tls/router/identity/ping"); err != nil {
		fmt.Println("Dial error", err)
	}
}

func newPeer(tcpPort, wsPort int, peerID string) (*fabric.Fabric, error) {
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
	f.AddTransport(fabric.NewTransportTCP(fmt.Sprintf("0.0.0.0:%d", tcpPort)))
	f.AddTransport(fabric.NewTransportWebsocket(fmt.Sprintf("0.0.0.0:%d", wsPort)))

	f.AddMiddleware(yamux)
	f.AddMiddleware(identity)
	f.AddMiddleware(ping)

	f.AddHandlerFunc("ping", ping.Handle)
	f.AddHandlerFunc("identity/ping", ping.Handle)

	return f, nil
}
