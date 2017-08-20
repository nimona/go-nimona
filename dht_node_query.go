package dht

import (
	net "github.com/nimona/go-nimona-net"
)

type query struct {
	id              string
	nonce           string
	closestPeer     net.Peer
	shortlistPeers  []net.Peer
	routingTable    *RoutingTable
	responseChannel chan net.Peer
}
