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

func newPeer(peerID string) (fnet.Net, error) {
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
	// ws := fnet.NewTransportWebsocket("0.0.0.0", 0)

	f := fnet.New(ctx)

	relay := fnet.NewRelayProtocol(f)

	f.AddTransport(yamux, router)
	f.AddTransport(tcp, tls, yamux, router)
	// f.AddTransport(ws, []fnet.Protocol{tls, yamux, router})

	f.AddProtocols(
		router,
		tls,
		yamux,
		identity,
		ping,
		relay,
	)

	router.AddRoute(relay)
	router.AddRoute(ping)
	router.AddRoute(identity, ping)

	return f, nil
}
