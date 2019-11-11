package net

import (
	"bufio"
	"io"

	"nimona.io/pkg/crypto"
)

type Connection struct {
	LocalPeerKey  crypto.PublicKey
	RemotePeerKey crypto.PublicKey
	IsIncoming    bool

	remoteAddress string
	localAddress  string

	conn  io.ReadWriteCloser
	lines chan []byte
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
		lines:      make(chan []byte, 100),
	}

	go func() {
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return
				} else {
					// TODO log error
					return
				}
			}
			c.lines <- line
		}
	}()

	return c
}
