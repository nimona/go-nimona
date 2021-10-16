package net

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

type Connection struct {
	ID string

	LocalPeerKey  crypto.PublicKey
	RemotePeerKey crypto.PublicKey
	IsIncoming    bool

	remoteAddress string
	localAddress  string

	encoder *json.Encoder
	decoder *json.Decoder

	conn   io.ReadWriteCloser
	closer chan struct{}

	pubsub ObjectPubSub
	mutex  sync.Mutex
}

func (c *Connection) Close() error {
	// TODO close all subs
	close(c.closer)
	return c.conn.Close()
}

func (c *Connection) LocalAddr() string {
	return c.localAddress
}

func (c *Connection) RemoteAddr() string {
	return c.remoteAddress
}

func (c *Connection) Write(o *object.Object) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.encoder.Encode(o); err != nil {
		return fmt.Errorf("error marshaling object: %w", err)
	}
	return nil
}

func (c *Connection) read() (o *object.Object, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("paniced while reading object: %w", r)
		}
	}()

	o = &object.Object{}
	err = c.decoder.Decode(o)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (c *Connection) Read(ctx context.Context) object.ReadCloser {
	errCh := make(chan error)
	sub := c.pubsub.Subscribe()
	return object.NewReadCloser(
		ctx,
		sub.Channel(),
		errCh,
		c.closer,
	)
}

func newConnection(conn io.ReadWriteCloser, incoming bool) *Connection {
	c := &Connection{
		ID:         rand.String(12),
		conn:       conn,
		IsIncoming: incoming,
		encoder:    json.NewEncoder(conn),
		decoder:    json.NewDecoder(conn),
		pubsub:     NewObjectPubSub(),
		mutex:      sync.Mutex{},
	}

	go func() {
		for {
			o, err := c.read()
			if err != nil {
				if err == io.EOF {
					// TODO make connection as closed
					return
				}
				// TODO consider closing on some errors
				continue
			}
			c.pubsub.Publish(o)
		}
	}()

	return c
}
