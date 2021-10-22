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

type (
	Connection interface {
		Close() error
		LocalAddr() string
		RemoteAddr() string
		LocalPeerKey() crypto.PublicKey
		RemotePeerKey() crypto.PublicKey
		Write(ctx context.Context, o *object.Object) error
		Read(ctx context.Context) object.ReadCloser
	}
	connection struct {
		ID string

		localPeerKey  crypto.PublicKey
		remotePeerKey crypto.PublicKey
		IsIncoming    bool

		remoteAddress string
		localAddress  string

		encoder *json.Encoder
		decoder *json.Decoder

		conn   io.ReadWriteCloser
		closer chan struct{}
		closed bool

		pubsub ObjectPubSub
		mutex  sync.RWMutex // R used to check if closed
	}
)

func (c *connection) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed {
		return nil
	}

	// TODO close all subs
	close(c.closer)
	c.closed = true
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *connection) IsClosed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.closed
}

func (c *connection) LocalAddr() string {
	return c.localAddress
}

func (c *connection) RemoteAddr() string {
	return c.remoteAddress
}

func (c *connection) LocalPeerKey() crypto.PublicKey {
	return c.localPeerKey
}

func (c *connection) RemotePeerKey() crypto.PublicKey {
	return c.remotePeerKey
}

func (c *connection) Write(ctx context.Context, o *object.Object) error {
	// TODO use context for timeout
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.encoder.Encode(o); err != nil {
		return fmt.Errorf("error marshaling object: %w", err)
	}
	return nil
}

func (c *connection) read() (o *object.Object, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("paniced while reading object: %w", r.(error))
		}
	}()

	o = &object.Object{}
	err = c.decoder.Decode(o)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (c *connection) Read(ctx context.Context) object.ReadCloser {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	errCh := make(chan error)
	sub := c.pubsub.Subscribe()
	return object.NewReadCloser(
		ctx,
		sub.Channel(),
		errCh,
		c.closer,
	)
}

func newConnection(conn io.ReadWriteCloser, incoming bool) *connection {
	c := &connection{
		ID:         rand.String(12),
		conn:       conn,
		IsIncoming: incoming,
		encoder:    json.NewEncoder(conn),
		decoder:    json.NewDecoder(conn),
		pubsub:     NewObjectPubSub(),
		mutex:      sync.RWMutex{},
		closer:     make(chan struct{}),
	}

	go func() {
		for {
			o, err := c.read()
			if err != nil {
				c.Close()
				return
			}
			c.pubsub.Publish(o)
		}
	}()

	return c
}
