package net

import (
	"encoding/json"
	"fmt"
	"io"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

type Connection struct {
	LocalPeerKey  crypto.PublicKey
	RemotePeerKey crypto.PublicKey
	IsIncoming    bool

	remoteAddress string
	localAddress  string

	encoder *json.Encoder
	decoder *json.Decoder

	conn  io.ReadWriteCloser
	lines chan *object.Object
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
		conn:       conn,
		IsIncoming: incoming,
		lines:      make(chan *object.Object, 100),
		encoder:    json.NewEncoder(conn),
		decoder:    json.NewDecoder(conn),
	}

	go func() {
		defer close(c.lines)
		defer conn.Close()
		for {
			o := &object.Object{}
			err := c.decoder.Decode(o)
			if err != nil {
				fmt.Println(">>>>> READ BYTES DONE", err)
				return
			}
			c.lines <- o
		}
	}()

	return c
}
