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
	port := 0

	if peerID == "bootstrap" {
		port = 26801
	} else {
		bsp = append(bsp, "tcp:localhost:26801/router/messaging")
	}

	ctx := context.Background()
	tcp := net.NewTransportTCP("0.0.0.0", port)

	net := net.New(ctx)
	rtr := protocol.NewRouter()

	pbs, _ := mesh.NewPubSub()
	reg, _ := mesh.NewRegisty(peerID, pbs)
	msh, _ := mesh.NewMesh(net, pbs, reg)
	msg, _ := mesh.NewMessenger(msh)
	dht.NewDHT(pbs, peerID, true, bsp...)

	net.AddProtocols(msg)

	rtr.AddRoute(msg)

	net.AddTransport(tcp, rtr)

	messages, _ := pbs.Subscribe(".*")
	for message := range messages {
		fmt.Printf("> Got new message %#v\n", message)
	}
}
