package net

import (
	"net"

	"nimona.io/pkg/net/peer"
)

type Connection struct {
	Conn       net.Conn
	RemotePeer *peer.PeerInfo
}
