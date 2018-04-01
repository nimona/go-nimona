package main

import (
	"context"
	"crypto/tls"
	"log"

	nnet "github.com/nimona/go-nimona/net"
	prot "github.com/nimona/go-nimona/net/protocol"
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
		_, c, err := peerB.DialContext(context.Background(), endpoint)
		if err != nil {
			log.Fatal("Dial error", err)
		}
		ping.Ping(c)
	}

}

func newPeer(peerID string) (nnet.Net, error) {
	ctx := context.Background()
	crt, err := GenX509KeyPair()
	if err != nil {
		return nil, err
	}

	yamux := prot.NewYamux()
	router := prot.NewRouter()
	identity := &prot.IdentityProtocol{Local: peerID}
	tls := &prot.SecProtocol{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}
	ping := &Ping{}

	tcp := nnet.NewTransportTCP("0.0.0.0", 0)

	nn := nnet.New(ctx)
	nn.AddTransport(yamux, router)
	nn.AddTransport(tcp, tls, yamux, router)
	nn.AddProtocols(router, tls, yamux, identity, ping)
	router.AddRoute(ping)
	return nn, nil
}
