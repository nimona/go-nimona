package dht

import (
	net "github.com/nimona/go-nimona-net"
)

type RoutingTable interface {
	Save(net.Peer) error
	Remove(net.Peer) error
	Get(string) (net.Peer, error)
	GetPeerIDs() ([]string, error)
}
