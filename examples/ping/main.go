package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	fabric "github.com/nimona/go-nimona-fabric"
)

func main() {
	peerA, err := newPeer(0, 0, "PeerA")
	if err != nil {
		log.Fatal("Could not create peer A", err)
	}

	peerB, err := newPeer(0, 0, "PeerB")
	if err != nil {
		log.Fatal("Could not create peer B", err)
	}

	log.Println("Peer A address:", peerA.GetAddresses())

	for _, addr := range peerA.GetAddresses() {
		endpoint := addr + "/tls/router/ping"
		log.Println("-------- Dialing", endpoint)
		if _, _, err := peerB.DialContext(context.Background(), endpoint); err != nil {
			log.Fatal("Dial error", err)
		}
	}
}

func newPeer(tcpPort, wsPort int, peerID string) (*fabric.Fabric, error) {
	ctx := context.Background()
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
	f.AddTransport(fabric.NewTransportTCP("0.0.0.0", tcpPort))
	f.AddTransport(fabric.NewTransportWebsocket(fmt.Sprintf("0.0.0.0:%d", wsPort)))

	f.AddMiddleware(yamux)
	f.AddMiddleware(identity)
	f.AddMiddleware(ping)

	f.AddHandlerFunc("ping", ping.Handle)
	f.AddHandlerFunc("identity/ping", ping.Handle)

	if err := f.Listen(ctx); err != nil {
		log.Fatal("Could not listen for peer A", err)
	}

	return f, nil
}
