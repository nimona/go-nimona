package nimona

import (
	"context"
	"fmt"

	"github.com/fxamacker/cbor"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

// ConnectionManager manages the dialing and accepting of connections.
// It maintains a cache of the last 100 connections.
type ConnectionManager struct {
	connCache  *lru.Cache[string, *RPC]
	dialer     Dialer
	listener   Listener
	handlers   map[string]HandlerFunc
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

type HandlerFunc func(context.Context, *Message) error

func NewConnectionManager(
	dialer Dialer,
	listener Listener,
	publicKey ed25519.PublicKey,
	privateKey ed25519.PrivateKey,
) (*ConnectionManager, error) {
	connCache, err := lru.NewWithEvict(100, func(addr string, rpc *RPC) {
		err := rpc.Close()
		if err != nil {
			// TODO: log error
			fmt.Println("error closing connection on eviction:", err)
			return
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error creating connection cache: %w", err)
	}

	c := &ConnectionManager{
		connCache:  connCache,
		dialer:     dialer,
		listener:   listener,
		publicKey:  publicKey,
		privateKey: privateKey,
		handlers:   map[string]HandlerFunc{},
	}

	if listener != nil {
		go func() {
			// TODO: handle error
			c.handleConnections(context.Background())
		}()
	}

	return c, nil
}

// Dial dials the given address and returns a connection if successful. If the
// address is already in the cache, the cached connection is returned.
func (cm *ConnectionManager) Dial(ctx context.Context, addr NodeAddr) (*RPC, error) {
	// check the cache
	existingConn, ok := cm.connCache.Get(addr.String())
	if ok {
		return existingConn, nil
	}

	// dial the address if it is not in the cache.
	conn, err := cm.dialer.Dial(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("error dialing %s: %w", addr, err)
	}

	// wrap the connection in a chunked connection
	sess := NewSession(conn)
	err = sess.DoServer(cm.publicKey, cm.privateKey)
	if err != nil {
		return nil, fmt.Errorf("error performing handshake: %w", err)
	}

	// wrap the connection in an rpc connection
	rpc := NewRPC(sess)

	// start handling messages
	go func() {
		cm.handleRPC(rpc)
	}()

	// add rpc to cache
	cm.connCache.Add(conn.RemoteAddr().String(), rpc)

	return rpc, nil
}

func (cm *ConnectionManager) handleConnections(ctx context.Context) error {
	errCh := make(chan error)
	// accept inbound connections.
	// if a connection with the same remote address already exists in the cache,
	// it is closed and removed before the new connection is added.
	go func() {
		for {
			conn, err := cm.listener.Accept()
			if err != nil {
				errCh <- fmt.Errorf("error accepting connection: %w", err)
				return
			}

			// start a new session, and perform the server side of the handshake
			// this will also perform the key exchange so after this we should
			// know the public key of the remote peer
			sess := NewSession(conn)
			err = sess.DoServer(cm.publicKey, cm.privateKey)
			if err != nil {
				// TODO: log error
				continue
			}

			remoteAddr := sess.NodeAddr().String()

			// check if a connection with the same remote address already exists in the cache.
			_, connectionExists := cm.connCache.Get(remoteAddr)
			if connectionExists {
				// remove the existing connection from the cache; this will
				// trigger the eviction callback which will close the connection
				cm.connCache.Remove(remoteAddr)
			}

			// wrap the connection in an rpc connection
			rpc := NewRPC(sess)

			// start handling messages
			go func() {
				cm.handleRPC(rpc)
			}()

			// add rpc to cache
			cm.connCache.Add(conn.RemoteAddr().String(), rpc)
		}
	}()

	return <-errCh
}

type MessageWrapper struct {
	Type string `cbor:"$type"`
}

func (cm *ConnectionManager) handleRPC(rpc *RPC) {
	for {
		msg, err := rpc.Read()
		if err != nil {
			// TODO log error
			fmt.Println("error reading message:", err)
			rpc.Close() // TODO handle error
			return
		}

		// assume we're dealing with cbor, and read the type of the message
		wrapper := &MessageWrapper{}
		err = cbor.Unmarshal(msg.Body, wrapper)
		if err != nil {
			// TODO log error
			fmt.Println("error unmarshaling message:", err)
			continue
		}

		// get the handler for the message type
		handler, ok := cm.handlers[wrapper.Type]
		if !ok {
			// TODO log error
			fmt.Println("no handler for message type:", wrapper.Type)
			continue
		}

		// handle the message
		err = handler(context.Background(), msg)
		if err != nil {
			// TODO log error
			fmt.Println("error handling message:", err)
			continue
		}
	}
}

func (cm *ConnectionManager) RegisterHandler(msgType string, handler HandlerFunc) {
	cm.handlers[msgType] = handler
}

func (cm *ConnectionManager) NodeAddr() NodeAddr {
	return NewNodeAddrWithKey(
		cm.listener.NodeAddr().Network(),
		cm.listener.NodeAddr().Address(),
		cm.publicKey,
	)
}

// Close closes all connections in the connection cache.
func (cm *ConnectionManager) Close() error {
	// purge will close all connections in the cache
	cm.connCache.Purge()
	return nil
}
