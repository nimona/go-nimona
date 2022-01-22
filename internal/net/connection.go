package net

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/Tv0ridobro/data-structure/list"

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

		// Buffer allows us to hold the last x number of objects received by
		// this connection and replay them when a new subscription is created.
		// This is done as subscriptions some times are not created fast enough
		// after a new connection is created and the subscribers are losing
		// objects. This is especially annoying in tests.
		// TODO revisit the way we deal with subscriptions in order to try and
		// avoid this buffer.
		buffer     *list.List[*object.Object]
		bufferLock sync.RWMutex
		bufferSize int
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

	c.bufferLock.RLock()
	previousObjects := c.buffer.GetAll()
	c.bufferLock.RUnlock()

	errCh := make(chan error)
	objCh := make(chan *object.Object)
	sub := c.pubsub.Subscribe()
	go func() {
		for _, o := range previousObjects {
			objCh <- o
		}
		// TODO: This is pretty wasteful, we need to first push the buffered
		// items and then continue with all published objects, but this go
		// routine is not really ideal.
		subCh := sub.Channel()
		for {
			o, ok := <-subCh
			if !ok {
				return
			}
			objCh <- o
		}
	}()
	return object.NewReadCloser(
		ctx,
		objCh,
		errCh,
		c.closer,
	)
}

func (c *connection) publish(o *object.Object) {
	c.bufferLock.Lock()
	c.buffer.PushBack(o)
	if c.buffer.Len() > c.bufferSize {
		c.buffer.PopFront()
	}
	c.bufferLock.Unlock()
	c.pubsub.Publish(o)
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
		buffer:     &list.List[*object.Object]{},
		bufferLock: sync.RWMutex{},
		bufferSize: 8,
	}

	go func() {
		for {
			o, err := c.read()
			if err != nil {
				c.Close()
				return
			}
			c.publish(o)
		}
	}()

	return c
}
