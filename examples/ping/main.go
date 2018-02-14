package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	fabric "github.com/nimona/go-nimona-fabric"
)

func main() {
	peerA, err := newPeer("PeerA")
	if err != nil {
		log.Fatal("Could not create peer A", err)
	}

	peerB, err := newPeer("PeerB")
	if err != nil {
		log.Fatal("Could not create peer B", err)
	}

	log.Println("Peer A address:", peerA.GetAddresses())

	for _, addr := range peerA.GetAddresses() {
		endpoint := addr + "/tls/yamux/router/identity/ping"
		log.Println("-------- Dialing", endpoint)
		if _, _, err := peerB.DialContext(context.Background(), endpoint); err != nil {
			log.Println("Dial error", err)
		}
	}
}

func newPeer(peerID string) (*fabric.Fabric, error) {
	ctx := context.Background()
	crt, err := GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return nil, err
	}

	yamux := fabric.NewYamux()
	router := fabric.NewRouter()
	identity := &fabric.IdentityProtocol{Local: peerID}
	tls := &fabric.SecProtocol{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}
	ping := &Ping{}

	tcp := fabric.NewTransportTCP("0.0.0.0", 0)
	ws := fabric.NewTransportWebsocket("0.0.0.0", 0)

	f := fabric.New(tls, yamux, router)

	f.AddTransport(tcp, []fabric.Protocol{yamux, router})
	f.AddTransport(ws, []fabric.Protocol{router})

	f.AddProtocol(yamux)
	f.AddProtocol(identity)
	f.AddProtocol(ping)

	router.AddRoute(ping)
	router.AddRoute(identity, ping)

	if err := f.Listen(ctx); err != nil {
		log.Fatal("Could not listen for peer A", err)
	}

	return f, nil
}
