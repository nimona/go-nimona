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
	"github.com/nimona/go-nimona/wire"
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

	reg, _ := mesh.NewRegisty(peerID)
	msh, _ := mesh.NewMesh(net, reg)
	wre, _ := wire.NewWire(msh, reg)
	dht.NewDHT(wre, reg, peerID, true, bsp...)

	net.AddProtocols(wre)
	net.AddProtocols(rly)
	net.AddProtocols(mux)

	rtr := protocol.NewRouter()
	rtr.AddRoute(wre)
	rtr.AddRoute(rly)

	net.AddTransport(mux, rtr)
	net.AddTransport(tcp, mux, rtr)

	for protocol, protocolAddresses := range net.GetProtocols() {
		fmt.Printf("%s:\n", protocol)
		for _, protocolAddress := range protocolAddresses {
			fmt.Printf("  - %s\n", protocolAddress)
		}
	}
}
