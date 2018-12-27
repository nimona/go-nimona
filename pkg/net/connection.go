package net

import (
	"net"
)

type Connection struct {
	Conn     net.Conn
	RemoteID string
}
