package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nimona/go-nimona/dht"
	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/net"
	"github.com/nimona/go-nimona/net/protocol"
)

func main() {
	peerID := os.Getenv("PEER_ID")
	if peerID == "" {
		log.Fatal("Missing PEER_ID")
	}

	bsp := []string{}
	rls := []string{}
	port := 0

	if peerID == "bootstrap" {
		port = 26801
	} else {
		rls = append(rls, "tcp:localhost:26801/yamux/router/relay")
		bsp = append(bsp, "tcp:localhost:26801/yamux/router/messaging")
	}

	ctx := context.Background()
	tcp := net.NewTransportTCP("0.0.0.0", port)

	net := net.New(ctx)
	rly := protocol.NewRelayProtocol(net, rls)
	mux := protocol.NewYamux()

	pbs, _ := mesh.NewPubSub()
	reg, _ := mesh.NewRegisty(peerID, pbs)
	msh, _ := mesh.NewMesh(net, pbs, reg)
	msg, _ := mesh.NewMessenger(msh)
	dht.NewDHT(pbs, peerID, true, bsp...)

	net.AddProtocols(msg)
	net.AddProtocols(rly)
	net.AddProtocols(mux)

	rtr := protocol.NewRouter()
	rtr.AddRoute(msg)
	rtr.AddRoute(rly)

	net.AddTransport(mux, rtr)
	net.AddTransport(tcp, mux, rtr)

	for protocol, protocolAddresses := range net.GetProtocols() {
		fmt.Printf("%s:\n", protocol)
		for _, protocolAddress := range protocolAddresses {
			fmt.Printf("  - %s\n", protocolAddress)
		}
	}

	messages, _ := pbs.Subscribe(".*")
	for message := range messages {
		fmt.Printf("> Got new message %#v\n", message)
	}
}
