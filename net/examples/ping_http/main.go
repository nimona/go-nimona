package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	fnet "github.com/nimona/go-nimona/net"
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

	ping := &Ping{}

	for _, addr := range peerA.GetAddresses() {
		endpoint := addr + "/tls/yamux/router/ping"
		_, c, err := peerB.DialAndProcessWithContext(context.Background(), endpoint)
		if err != nil {
			log.Fatal("Dial error", err)
		}
		ping.Ping(c)
	}

}

func newPeer(peerID string) (*fnet.Net, error) {
	ctx := context.Background()
	crt, err := GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return nil, err
	}

	yamux := fnet.NewYamux()
	router := fnet.NewRouter()
	identity := &fnet.IdentityProtocol{Local: peerID}
	tls := &fnet.SecProtocol{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}
	ping := &Ping{}

	tcp := fnet.NewTransportTCP("0.0.0.0", 0)

	f := fnet.New(ctx)

	f.AddTransport(yamux, []fnet.Protocol{router})
	f.AddTransport(tcp, []fnet.Protocol{tls, yamux, router})

	f.AddProtocol(router)
	f.AddProtocol(tls)
	f.AddProtocol(yamux)
	f.AddProtocol(identity)
	f.AddProtocol(ping)

	router.AddRoute(ping)

	return f, nil
}
