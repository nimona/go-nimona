package main

import (
	"context"
	"log"
	"os"

	"github.com/nimona/go-nimona/blx"
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

	bs := []string{}
	port := 0

	if peerID == "bootstrap" {
		port = 26801
	} else {
		bs = append(bs, "tcp:localhost:26801/router/wire")
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

	blx.NewBlockExchange(pbs)

	net.AddProtocols(wre)

	rtr.AddRoute(wre)

	net.AddTransport(tcp, rtr)
}
