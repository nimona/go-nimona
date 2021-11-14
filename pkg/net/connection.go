package net

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

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

		conn    io.ReadWriteCloser
		closer  chan struct{}
		closed  bool
		reading chan struct{}

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

func (c *connection) Read(ctx context.Context) object.ReadCloser {
	// TODO: link ctx with errCh?
	errCh := make(chan error)
	sub := c.pubsub.Subscribe()
	r := object.NewReadCloser(
		ctx,
		sub.Channel(),
		errCh,
		c.closer,
	)
	// TODO: refactor, this seems kind of hacky
	select {
	case c.reading <- struct{}{}:
	default:
	}
	return r
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
		closer:     make(chan struct{}, 1),
		reading:    make(chan struct{}, 1),
	}

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				c.Close()
			}
		}()
		// wait until someone is reading
		select {
		case <-c.reading:
		case <-time.After(250 * time.Millisecond):
		}
		for {
			o := &object.Object{}
			err := c.decoder.Decode(o)
			if err != nil {
				c.Close()
				return
			}
			c.pubsub.Publish(o)
		}
	}()

	return c
}
