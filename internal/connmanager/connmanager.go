package connmanager

import (
	"crypto/ed25519"
	"crypto/tls"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

//go:generate genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub_objects_generated.go -pkg=connmanager gen "ObjectType=*object.Object Name=Object name=object"

var (
	connConnOutCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_net_conn_out_total",
			Help: "Total number of outgoing connections",
		},
	)
	connConnIncCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_net_conn_in_total",
			Help: "Total number of incoming connections",
		},
	)
	connDialAttemptCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_net_dial_attempt_total",
			Help: "Total number of dial attempts",
		},
	)
	connDialSuccessCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_net_dial_success_total",
			Help: "Total number of successful dials",
		},
	)
	connDialErrorCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_net_dial_failed_total",
			Help: "Total number of failed dials",
		},
	)
	connDialBlockedCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_net_dial_blocked_total",
			Help: "Total number of failed dials due to all addresses blocked",
		},
	)
)

type (
	// ConnManager interface
	ConnManager interface {
		Dial(
			ctx context.Context,
			peer *peer.ConnectionInfo,
		) (Connection, error)
		Listen(
			ctx context.Context,
			bindAddress string,
			listenConfig *ListenConfig,
		) (Listener, error)
		RegisterConnectionHandler(
			handler ConnectionHandler,
		)
		Addresses() []string
	}
	ConnectionHandler func(Connection)
)

// New creates a new transport
func New(
	peerKey crypto.PrivateKey,
) ConnManager {
	n := &network{
		peerKey: peerKey,
		transports: map[string]Transport{
			"tcps": &tcpTransport{
				peerKey: peerKey,
			},
		},
		listeners:        []*listener{},
		blocklist:        cache.New(time.Second*5, time.Second*60),
		connections:      map[string]*connection{},
		connectionsMutex: &sync.RWMutex{},
		connHandlers:     []ConnectionHandler{},
		connHandlerMutex: sync.RWMutex{},
	}
	return n
}

// network allows dialing and listening for p2p connections
type network struct {
	peerKey    crypto.PrivateKey
	transports map[string]Transport
	listeners  []*listener
	attempts   attemptsMap
	blocklist  *cache.Cache

	// connections
	connections      map[string]*connection
	connectionsMutex *sync.RWMutex

	// handlers
	connHandlers     []ConnectionHandler
	connHandlerMutex sync.RWMutex
}

// Dial to a peer and return a net.Conn or error
func (n *network) Dial(
	ctx context.Context,
	p *peer.ConnectionInfo,
) (Connection, error) {
	logger := log.FromContext(ctx).With(
		log.String("peer", p.Metadata.Owner.String()),
		log.Strings("addresses", p.Addresses),
	)

	// TODO: Can we dial a peer without an owner/public-key?

	pubKey, err := crypto.PublicKeyFromDID(p.Metadata.Owner)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key from did: %w", err)
	}

	n.connectionsMutex.RLock()
	conn, ok := n.connections[pubKey.String()]
	if ok {
		if !conn.IsClosed() {
			n.connectionsMutex.RUnlock()
			return conn, nil
		}
	}
	n.connectionsMutex.RUnlock()

	if len(p.Addresses) == 0 {
		return nil, ErrNoAddresses
	}

	logger.Debug("dialing")
	connDialAttemptCounter.Inc()

	// keep a flag on whether all addresses where blocked so we can return
	// an ErrAllAddressesBlocked error
	allBlocked := true

	// go through all addresses and try to dial them
	for _, address := range p.Addresses {
		// check if address is currently blocked
		if n.isAddressBlocked(*pubKey, address) {
			logger.Debug("address is blocked, skipping")
			continue
		}
		// get protocol from address
		addressType := strings.Split(address, ":")[0]
		trsp, ok := n.transports[addressType]
		if !ok {
			logger.Debug("not sure how to dial",
				log.String("type", addressType),
			)
			continue
		}

		// reset blocked flag
		allBlocked = false

		// dial address
		conn, err := trsp.Dial(ctx, address)
		if err != nil {
			// blocking address
			attempts, backoff := n.blockAddress(
				*pubKey,
				address,
			)
			logger.Error("could not dial address, blocking",
				log.Int("failedAttempts", attempts),
				log.String("backoff", backoff.String()),
				log.String("type", addressType),
				log.Error(err),
			)
			continue
		}

		// check negotiated key against dialed
		if !conn.remotePeerKey.Equals(*pubKey) {
			n.blockAddress(
				*pubKey,
				address,
			)
			logger.Error("remote didn't match expect key, blocking",
				log.String("expected", pubKey.String()),
				log.String("received", conn.remotePeerKey.String()),
			)
			continue
		}

		// try to write something
		ping := &object.Object{
			Type: "ping",
			Data: tilde.Map{
				"dt": tilde.String(time.Now().Format(time.RFC3339)),
			},
		}
		if err := conn.Write(ctx, ping); err != nil {
			n.blockAddress(
				*pubKey,
				address,
			)
			logger.Error("could not actually write to remote, blocking")
			continue
		}

		// at this point we consider the connection successful, so we can
		// reset the failed attempts
		n.attempts.Put(address, 0)
		n.attempts.Put(pubKey.String(), 0)

		connDialSuccessCounter.Inc()
		connConnOutCounter.Inc()

		n.handleNewConnection(conn)

		return conn, nil
	}

	err = ErrAllAddressesFailed
	if allBlocked {
		err = ErrAllAddressesBlocked
		connDialBlockedCounter.Inc()
	} else {
		connDialErrorCounter.Inc()
	}

	logger.Error("could not dial peer", log.Error(err))
	return nil, err
}

func (n *network) isAddressBlocked(
	publicKey crypto.PublicKey,
	address string,
) bool {
	_, blocked := n.blocklist.Get(publicKey.String() + "/" + address)
	return blocked
}

func (n *network) blockAddress(
	publicKey crypto.PublicKey,
	address string,
) (int, time.Duration) {
	pk := publicKey.String() + "/" + address
	baseBackoff := float64(time.Second * 1)
	maxBackoff := float64(time.Minute * 10)
	attempts, _ := n.attempts.Get(pk)
	attempts++
	backoff := baseBackoff * math.Pow(1.5, float64(attempts))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	n.attempts.Put(pk, attempts)
	n.blocklist.Set(pk, attempts, time.Duration(backoff))
	return attempts, time.Duration(backoff)
}

// func (n *network) Accept() (*Connection, error) {
// 	conn := <-n.connections
// 	return conn, nil
// }

type ListenConfig struct {
	BindLocal   bool
	BindPrivate bool
	BindIPV6    bool
}

// Listen
// TODO do we need to return a listener?
func (n *network) Listen(
	ctx context.Context,
	bindAddress string,
	listenConfig *ListenConfig,
) (Listener, error) {
	mlst := &listener{
		addresses: []string{},
		listeners: []net.Listener{},
	}
	k := n.peerKey
	for pt, tsp := range n.transports {
		lst, err := tsp.Listen(ctx, bindAddress, k)
		if err != nil {
			return nil, err
		}

		if listenConfig == nil {
			listenConfig = &ListenConfig{}
		}

		addresses := GetAddresses(
			pt,
			lst,
			listenConfig.BindLocal,
			listenConfig.BindPrivate,
			listenConfig.BindIPV6,
		)
		mlst.listeners = append(mlst.listeners, lst)
		mlst.addresses = append(mlst.addresses, addresses...)

		n.listeners = append(n.listeners, mlst)

		// TODO goroutine never ends
		go func() {
			for {
				rawConn, err := lst.Accept()
				if err != nil {
					// we need to check whether the error is temporary,
					// a non-temporary error would be for example a closed
					// listener
					errIsTemp := true
					if opErr, ok := err.(*net.OpError); ok {
						errIsTemp = opErr.Temporary()
					}
					// if the error is not temporary stop trying to accept
					// connections from this lsitener
					if !errIsTemp {
						return
					}
					// else, just move on
					continue
				}

				conn := newConnection(rawConn, true)
				conn.remoteAddress = "tcps:" + rawConn.RemoteAddr().String()
				conn.localAddress = rawConn.LocalAddr().String()

				if tlsConn, ok := rawConn.(*tls.Conn); ok {
					if err := tlsConn.Handshake(); err != nil {
						// not currently supported
						// TODO find a way to surface this error
						conn.Close() // nolint: errcheck
						continue
					}
					state := tlsConn.ConnectionState()
					certs := state.PeerCertificates
					if len(certs) != 1 {
						// not currently supported
						// TODO find a way to surface this error
						conn.Close() // nolint: errcheck
						continue
					}
					pubKey, ok := certs[0].PublicKey.(ed25519.PublicKey)
					if !ok {
						// not currently supported
						// TODO find a way to surface this error
						conn.Close() // nolint: errcheck
						continue
					}
					conn.remotePeerKey = crypto.NewEd25519PublicKeyFromRaw(
						pubKey,
					)
				} else {
					// not currently supported
					// TODO find a way to surface this error
					conn.Close() // nolint: errcheck
					continue
				}

				// TODO check if the remote key is the expected one

				connConnIncCounter.Inc()

				n.handleNewConnection(conn)
			}
		}()
	}
	// block our own addresses, just in case anyone tries to dial them
	for _, addr := range mlst.addresses {
		n.blocklist.Set(k.PublicKey().String()+"/"+addr, 0, cache.NoExpiration)
	}
	return mlst, nil
}

func (n *network) Addresses() []string {
	addrs := []string{}
	for _, l := range n.listeners {
		addrs = append(addrs, l.Addresses()...)
	}
	return addrs
}

func (n *network) RegisterConnectionHandler(
	handler ConnectionHandler,
) {
	n.connHandlerMutex.Lock()
	n.connHandlers = append(n.connHandlers, handler)
	n.connHandlerMutex.Unlock()
}

func (n *network) handleNewConnection(conn *connection) {
	// add connection to list of connections
	n.connectionsMutex.Lock()
	n.connections[conn.remotePeerKey.String()] = conn
	n.connectionsMutex.Unlock()
	// call all connection handlers
	n.connHandlerMutex.Lock()
	for _, handler := range n.connHandlers {
		handler(conn)
	}
	n.connHandlerMutex.Unlock()
}
