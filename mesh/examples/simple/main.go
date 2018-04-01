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

	bs := []string{
		"tcp:localhost:26801/router/messaging",
	}
	port := 0

	ctx := context.Background()
	tcp := net.NewTransportTCPWithUPNP("0.0.0.0", port)

	nn := net.New(ctx)
	rt := protocol.NewRouter()

	ps, _ := mesh.NewPubSub()
	rg, _ := mesh.NewRegisty(peerID, ps)
	ms, _ := mesh.NewMesh(nn, ps, rg)
	mg, _ := mesh.NewMessenger(ms)

	dht.NewDHT(ps, peerID, bs...)

	nn.AddProtocols(mg)

	rt.AddRoute(mg)

	nn.AddTransport(tcp, rt)

	messages, _ := ps.Subscribe(".*")
	for message := range messages {
		fmt.Printf("> Got new message %#v\n", message)
	}
}
