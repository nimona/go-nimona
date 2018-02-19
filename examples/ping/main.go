package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	fabric "github.com/nimona/go-nimona-fabric"
	protocol "github.com/nimona/go-nimona-fabric/protocol"
	transport "github.com/nimona/go-nimona-fabric/transport"
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
		if err := peerB.CallContext(context.Background(), endpoint); err != nil {
			log.Println("Dial error", err)
		}

		// endpoint = addr + "/tls/yamux/router/ping"
		// time.Sleep(5 * time.Second)
		// log.Println("-------- SECOND Dial", endpoint)
		// if err := peerB.CallContext(context.Background(), endpoint); err != nil {
		// 	log.Println("Dial error", err)
		// }

		addrPeerB := peerB.GetAddresses()[0]
		endpoint = addrPeerB + "/tls/yamux/router/relay:" + addr + "/tls/yamux/router/ping"
		log.Println("-------- THIRD Dial", endpoint)
		peerC.CallContext(context.Background(), endpoint)

	}

}

func newPeer(peerID string) (*fabric.Fabric, error) {
	ctx := context.Background()
	crt, err := GenX509KeyPair()
	if err != nil {
		fmt.Println("Cert creation error", err)
		return nil, err
	}

	yamux := protocol.NewYamux()
	router := protocol.NewRouter()
	identity := &protocol.IdentityProtocol{Local: peerID}
	tls := &protocol.SecProtocol{
		Config: tls.Config{
			Certificates:       []tls.Certificate{crt},
			InsecureSkipVerify: true,
		},
	}
	ping := &Ping{}

	tcp := transport.NewTransportTCP("0.0.0.0", 0)
	// ws := fabric.NewTransportWebsocket("0.0.0.0", 0)

	f := fabric.New(ctx)

	relay := protocol.NewRelayProtocol(f)

	f.AddTransport(yamux, []protocol.Protocol{router})
	f.AddTransport(tcp, []protocol.Protocol{tls, yamux, router})
	// f.AddTransport(ws, []fabric.Protocol{tls, yamux, router})

	f.AddProtocol(router)
	f.AddProtocol(tls)
	f.AddProtocol(yamux)
	f.AddProtocol(identity)
	f.AddProtocol(ping)
	f.AddProtocol(relay)

	router.AddRoute(relay)
	router.AddRoute(ping)
	router.AddRoute(identity, ping)

	return f, nil
}
