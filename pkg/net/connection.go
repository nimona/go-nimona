package net

import (
	"net"

	"nimona.io/pkg/crypto"
)

type Connection struct {
	Conn          net.Conn
	RemotePeerKey *crypto.PublicKey
	IsIncoming    bool
}
