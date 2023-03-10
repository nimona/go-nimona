package nimona

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/golang-lru/v2/simplelru"

	"nimona.io/internal/xsync"
)

// SessionManager manages the dialing and accepting of connections.
// It maintains a cache of the last 100 connections.
type SessionManager struct {
	connCache  *simplelru.LRU[connCacheKey, *Session]
	dialer     Dialer
	listener   Listener
	handlers   map[string]RequestHandlerFunc
	publicKey  PublicKey
	privateKey PrivateKey

	resolver   Resolver
	aliases    xsync.Map[IdentityAlias, *IdentityInfo]
	providers  xsync.Map[IdentityAlias, *IdentityInfo]
	identities xsync.Map[DocumentID, *IdentityInfo]
}

type RequestHandlerFunc func(context.Context, *Request) error

type connCacheKey struct {
	publicKeyInHex string
}

func NewSessionManager(
	dialer Dialer,
	listener Listener,
	publicKey PublicKey,
	privateKey PrivateKey,
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
		resolver:   &ResolverHTTP{},
	}

	if listener != nil {
		go func() {
			// nolint:errcheck // TODO: handle error
			c.handleConnections()
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

func (cm *SessionManager) connCacheKey(k PublicKey) connCacheKey {
	return connCacheKey{
		publicKeyInHex: fmt.Sprintf("%x", k),
	}
}

type RequestRecipientFn func(*requestRecipient)

type requestRecipient struct {
	Alias    *IdentityAlias
	Identity *Identity
	PeerAddr *PeerAddr
}

func FromAlias(alias IdentityAlias) RequestRecipientFn {
	return func(r *requestRecipient) {
		r.Alias = &alias
	}
}

func FromIdentity(identity Identity) RequestRecipientFn {
	return func(r *requestRecipient) {
		r.Identity = &identity
	}
}

func FromPeerAddr(peerAddr PeerAddr) RequestRecipientFn {
	return func(r *requestRecipient) {
		r.PeerAddr = &peerAddr
	}
}

func (cm *SessionManager) Request(
	ctx context.Context,
	req *Document,
	rfn RequestRecipientFn,
) (*Response, error) {
	rec := &requestRecipient{}
	rfn(rec)

	switch {
	case rec.Alias != nil:
		return cm.requestFromAlias(ctx, *rec.Alias, req)
	case rec.Identity != nil:
		return cm.requestFromIdentity(ctx, rec.Identity, req)
	case rec.PeerAddr != nil:
		return cm.requestFromPeerAddr(ctx, *rec.PeerAddr, req)
	default:
		return nil, fmt.Errorf("no recipient specified")
	}
}

func (cm *SessionManager) requestFromAlias(
	ctx context.Context,
	alias IdentityAlias,
	req *Document,
) (*Response, error) {
	// resolve the alias
	info, err := cm.LookupAlias(alias)
	if err != nil {
		return nil, fmt.Errorf("error looking up alias %s: %w", alias, err)
	}

	return cm.requestFromIdentity(ctx, &info.Identity, req)
}

func (cm *SessionManager) requestFromIdentity(
	ctx context.Context,
	identity *Identity,
	req *Document,
) (*Response, error) {
	// resolve the identity
	info, err := cm.LookupIdentity(NewDocumentID(identity.Document()))
	if err != nil {
		return nil, fmt.Errorf("error looking up identity %s: %w", identity, err)
	}

	return cm.requestFromPeerAddr(ctx, info.PeerAddresses[0], req)
}

func (cm *SessionManager) requestFromPeerAddr(
	ctx context.Context,
	addr PeerAddr,
	req *Document,
) (*Response, error) {
	ses, err := cm.Dial(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("error dialing %s: %w", addr, err)
	}

	return ses.Request(ctx, req)
}

func (cm *SessionManager) handleConnections() error {
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

func (cm *SessionManager) LookupAlias(alias IdentityAlias) (*IdentityInfo, error) {
	if info, ok := cm.aliases.Load(alias); ok {
		return info, nil
	}

	identityInfo, err := cm.resolver.ResolveIdentityAlias(alias)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve provider alias: %w", err)
	}

	cm.aliases.Store(alias, identityInfo)
	cm.identities.Store(NewDocumentID(identityInfo.Identity.Document()), identityInfo)

	// TODO add recursive lookup for user identities
	if identityInfo.Identity.Use == "provider" {
		cm.providers.Store(alias, identityInfo)
	}

	return identityInfo, nil
}

func (cm *SessionManager) LookupIdentity(identityID DocumentID) (*IdentityInfo, error) {
	if info, ok := cm.identities.Load(identityID); ok {
		return info, nil
	}

	var identityInfo *IdentityInfo
	cm.providers.Range(func(key IdentityAlias, providerInfo *IdentityInfo) bool {
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		defer cf()

		for _, addr := range providerInfo.PeerAddresses {
			rctx := &RequestContext{}
			doc, err := RequestDocument(ctx, rctx, cm, identityID, FromPeerAddr(addr))
			if err != nil {
				continue
			}

			err = identityInfo.FromDocument(doc)
			if err != nil {
				continue
			}

			return false
		}

		return true
	})

	if identityInfo == nil {
		return nil, fmt.Errorf("unable to resolve identity %s", identityID)
	}

	cm.identities.Store(identityID, identityInfo)

	// TODO verify the alias is indeed correct before storing or returning it
	if identityInfo.Alias.Hostname != "" {
		cm.aliases.Store(identityInfo.Alias, identityInfo)
	}

	if identityInfo.Identity.Use == "provider" {
		cm.providers.Store(identityInfo.Alias, identityInfo)
	}

	return identityInfo, nil
}
