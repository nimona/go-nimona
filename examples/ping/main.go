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
		if err := peerB.CallContext(context.Background(), endpoint); err != nil {
			log.Println("Dial error", err)
		}
		endpoint = addr + "/tls/yamux/router/ping"
		log.Println("-------- SECOND Dial", endpoint)
		if err := peerB.CallContext(context.Background(), endpoint); err != nil {
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
	// ws := fabric.NewTransportWebsocket("0.0.0.0", 0)

	f := fabric.New(ctx)

	f.AddTransport(yamux, []fabric.Protocol{router})
	f.AddTransport(tcp, []fabric.Protocol{tls, yamux, router})
	// f.AddTransport(ws, []fabric.Protocol{tls, yamux, router})

	f.AddProtocol(router)
	f.AddProtocol(tls)
	f.AddProtocol(yamux)
	f.AddProtocol(identity)
	f.AddProtocol(ping)

	router.AddRoute(ping)
	router.AddRoute(identity, ping)

	return f, nil
}
