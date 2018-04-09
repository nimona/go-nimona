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

	peerC, err := newPeer("PeerC")
	if err != nil {
		log.Fatal("Could not create peer C", err)
	}

	log.Println("Peer A address:", peerA.GetAddresses())

	for _, addr := range peerA.GetAddresses() {
		endpoint := addr + "/tls/yamux/router/relay:keepalive"
		log.Println("-------- Dialing", endpoint)
		if _, _, err := peerB.DialContext(context.Background(), endpoint); err != nil {
			log.Println("Dial error", err)
		}

		// endpoint = addr + "/tls/yamux/router/ping"
		// time.Sleep(5 * time.Second)
		// log.Println("-------- SECOND Dial", endpoint)
		// if err := peerB.DialContext(context.Background(), endpoint); err != nil {
		// 	log.Println("Dial error", err)
		// }

		addrPeerB := peerB.GetAddresses()[0]
		endpoint = addrPeerB + "/tls/yamux/router/relay:" + addr + "/tls/yamux/router/ping"
		log.Println("-------- THIRD Dial", endpoint)
		if _, _, err := peerC.DialContext(context.Background(), endpoint); err != nil {
			log.Println("Dial error", err)
		}
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
	// ws := nnet.NewTransportWebsocket("0.0.0.0", 0)

	nn := nnet.New(ctx)
	relay := prot.NewRelayProtocol(nn, []string{})
	nn.AddTransport(yamux, router)
	nn.AddTransport(tcp, tls, yamux, router)
	// nn.AddTransport(ws, []nnet.Protocol{tls, yamux, router})
	nn.AddProtocols(router, tls, yamux, identity, ping, relay)
	router.AddRoute(relay)
	router.AddRoute(ping)
	router.AddRoute(identity, ping)
	return nn, nil
}
