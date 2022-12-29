package nimona

import (
	"context"
	"fmt"

	"github.com/hashicorp/golang-lru/v2/simplelru"
	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

// SessionManager manages the dialing and accepting of connections.
// It maintains a cache of the last 100 connections.
type SessionManager struct {
	connCache  *simplelru.LRU[connCacheKey, *Session]
	dialer     Dialer
	listener   Listener
	handlers   map[string]RequestHandlerFunc
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

type RequestHandlerFunc func(context.Context, *Request) error

type connCacheKey struct {
	publicKeyInHex string
}

func NewSessionManager(
	dialer Dialer,
	listener Listener,
	publicKey ed25519.PublicKey,
	privateKey ed25519.PrivateKey,
) (*SessionManager, error) {
	connCache, err := simplelru.NewLRU(100, func(_ connCacheKey, ses *Session) {
		err := ses.Close()
		if err != nil {
			// TODO: log error
			fmt.Println("error closing connection on eviction:", err)
			return
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error creating connection cache: %w", err)
	}

	c := &SessionManager{
		connCache:  connCache,
		dialer:     dialer,
		listener:   listener,
		publicKey:  publicKey,
		privateKey: privateKey,
		handlers:   map[string]RequestHandlerFunc{},
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
func (cm *SessionManager) Dial(
	ctx context.Context,
	addr PeerAddr,
) (*Session, error) {
	// check the cache
	existingConn, ok := cm.connCache.Get(cm.connCacheKey(addr.PublicKey))
	if ok {
		return existingConn, nil
	}

	// dial the address if it is not in the cache.
	conn, err := cm.dialer.Dial(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("error dialing %s: %w", addr, err)
	}

	// wrap the connection in a chunked connection
	ses := NewSession(conn)
	err = ses.DoServer(cm.publicKey, cm.privateKey)
	if err != nil {
		return nil, fmt.Errorf("error performing handshake: %w", err)
	}

	// start handling messages
	go func() {
		cm.handleSession(ses)
	}()

	// add ses to cache
	cm.connCache.Add(cm.connCacheKey(ses.PublicKey()), ses)

	return ses, nil
}

func (cm *SessionManager) connCacheKey(k ed25519.PublicKey) connCacheKey {
	return connCacheKey{
		publicKeyInHex: fmt.Sprintf("%x", k),
	}
}

func (cm *SessionManager) Request(
	ctx context.Context,
	addr PeerAddr,
	req Cborer,
) (*Response, error) {
	ses, err := cm.Dial(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("error dialing %s: %w", addr, err)
	}

	return ses.Request(ctx, req)
}

func (cm *SessionManager) handleConnections(ctx context.Context) error {
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
			ses := NewSession(conn)
			err = ses.DoServer(cm.publicKey, cm.privateKey)
			if err != nil {
				// TODO: log error
				continue
			}

			// check if a connection with the same remote address already exists
			// in the cache.
			connCacheKey := cm.connCacheKey(ses.PublicKey())
			_, connectionExists := cm.connCache.Get(connCacheKey)
			if connectionExists {
				// remove the existing connection from the cache; this will
				// trigger the eviction callback which will close the connection
				cm.connCache.Remove(connCacheKey)
			}

			// start handling messages
			go func() {
				cm.handleSession(ses)
			}()

			// add ses to cache
			cm.connCache.Add(connCacheKey, ses)
		}
	}()

	return <-errCh
}

func (cm *SessionManager) handleSession(ses *Session) {
	for {
		req, err := ses.Read()
		if err != nil {
			// TODO log error
			fmt.Println("error reading message:", err)
			ses.Close() // TODO handle error
			return
		}

		// get the handler for the message type
		handler, ok := cm.handlers[req.Type]
		if !ok {
			// TODO log error
			fmt.Println("no handler for message type:", req.Type)
			continue
		}

		// handle the message
		err = handler(context.Background(), req)
		if err != nil {
			// TODO log error
			fmt.Println("error handling message:", err)
			continue
		}
	}
}

func (cm *SessionManager) RegisterHandler(
	msgType string,
	handler RequestHandlerFunc,
) {
	cm.handlers[msgType] = handler
}

func (cm *SessionManager) PeerAddr() PeerAddr {
	return PeerAddr{
		Network:   cm.listener.PeerAddr().Network,
		Address:   cm.listener.PeerAddr().Address,
		PublicKey: cm.publicKey,
	}
}

// Close closes all connections in the connection cache.
func (cm *SessionManager) Close() error {
	// purge will close all connections in the cache
	cm.connCache.Purge()
	return nil
}
