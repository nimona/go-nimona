package net

import (
	"bufio"
	"io"

	"nimona.io/pkg/crypto"
)

type Connection struct {
	RemotePeerKey *crypto.PublicKey
	IsIncoming    bool

	conn  io.ReadWriteCloser
	lines chan []byte
}

func (c *Connection) Close() error {
	return c.conn.Close()
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
