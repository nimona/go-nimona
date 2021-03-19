package net

import (
	"encoding/json"
	"io"

	"nimona.io/internal/rand"
	"nimona.io/pkg/crypto"
)

type Connection struct {
	ID string

	LocalPeerKey  *crypto.PublicKey
	RemotePeerKey *crypto.PublicKey
	IsIncoming    bool

	remoteAddress string
	localAddress  string

	encoder *json.Encoder
	decoder *json.Decoder

	conn io.ReadWriteCloser
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) LocalAddr() string {
	return c.localAddress
}

func (c *Connection) RemoteAddr() string {
	return c.remoteAddress
}

func newConnection(conn io.ReadWriteCloser, incoming bool) *Connection {
	c := &Connection{
		ID:         rand.String(12),
		conn:       conn,
		IsIncoming: incoming,
		encoder:    json.NewEncoder(conn),
		decoder:    json.NewDecoder(conn),
	}

	return c
}
