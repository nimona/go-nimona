package net // import "nimona.io/go/net"

import (
	"net"
)

type Connection struct {
	Conn     net.Conn
	RemoteID string
}
