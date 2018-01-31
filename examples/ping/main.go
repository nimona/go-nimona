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

	if err := peerA.Listen(ctx); err != nil {
		log.Fatal("Could not listen for peer A", err)
	}

	peerB, err := newPeer(4001, 4002, "PeerB")
	if err != nil {
		log.Fatal("Could not create peer B", err)
	}

	if err := peerB.Listen(ctx); err != nil {
		log.Fatal("Could not listen for peer A", err)
	}

	addrsA := peerA.GetAddresses()
	addrsB := peerB.GetAddresses()

	log.Println("Peer A addresses:", addrsA)
	log.Println("Peer B addresses:", addrsB)

	for _, addr := range addrsA {
		endpoint := addr + "/tls/router/ping"
		log.Println("-------- Dialing", endpoint)
		if _, _, err := peerB.DialContext(context.Background(), endpoint); err != nil {
			log.Fatal("Dial error", err)
		}
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
	f.AddTransport(fabric.NewTransportTCP("0.0.0.0", tcpPort))
	f.AddTransport(fabric.NewTransportWebsocket(fmt.Sprintf("0.0.0.0:%d", wsPort)))

	f.AddMiddleware(yamux)
	f.AddMiddleware(identity)
	f.AddMiddleware(ping)

	f.AddHandlerFunc("ping", ping.Handle)
	f.AddHandlerFunc("identity/ping", ping.Handle)

	return f, nil
}
