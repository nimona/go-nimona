package net

import (
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/peer"
)

var UseUPNP = false

func init() {
	UseUPNP, _ = strconv.ParseBool(os.Getenv("UPNP"))
}

// Network interface
type Network interface {
	Dial(ctx context.Context, address string, options ...Option) (*Connection, error)
	Listen(ctx context.Context) (chan *Connection, error)

	AddMiddleware(handler MiddlewareHandler)
	AddTransport(tag string, tsp Transport)
}

type (
	// Options (mostly) for Dial()
	Options struct {
		LocalDiscoveryOnly bool
	}
	Option func(*Options)
)

func WithLocalDiscoveryOnly() Option {
	return func(options *Options) {
		options.LocalDiscoveryOnly = true
	}
}

// New creates a new p2p network using an address book
func New(discover discovery.Discoverer, local *peer.LocalPeer) (Network, error) {
	return &network{
		discoverer: discover,
		middleware: []MiddlewareHandler{},
		local:      local,
		midLock:    &sync.RWMutex{},
		transports: &sync.Map{},
	}, nil
}

// network allows dialing and listening for p2p connections
type network struct {
	discoverer discovery.Discoverer
	local      *peer.LocalPeer
	midLock    *sync.RWMutex
	transports *sync.Map
	middleware []MiddlewareHandler
}

func (n *network) AddMiddleware(handler MiddlewareHandler) {
	n.midLock.Lock()
	defer n.midLock.Unlock()
	n.middleware = append(n.middleware, handler)
}

func (n *network) AddTransport(tag string, tsp Transport) {
	n.transports.Store(tag, tsp)
}

// Dial to a peer and return a net.Conn or error
func (n *network) Dial(
	ctx context.Context,
	address string,
	opts ...Option,
) (*Connection, error) {
	logger := log.FromContext(ctx).With(
		log.String("address", address),
	)

	logger.Debug("dialing")

	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	var conn *Connection
	var err error

	addressType := strings.Split(address, ":")[0]
	switch addressType {
	case "peer":
		conn, err = n.dialPeer(ctx, address, options.LocalDiscoveryOnly)
	default:
		t, ok := n.transports.Load(addressType)
		if !ok {
			logger.Info("not sure how to dial",
				log.String("type", addressType),
			)
			return nil, ErrNoAddresses
		}

		trsp := t.(Transport)

		conn, err = trsp.Dial(ctx, address)
		if err != nil {
			return nil, err
		}

		for _, mh := range n.middleware {
			conn, err = mh(ctx, conn)
			if err != nil {
				return nil, err
			}
		}
	}
	if err != nil {
		logger.Error("could not dial address", log.Error(err))
		return nil, err
	}

	return conn, nil
}

// Listen
// TODO do we need to return a listener?
func (n *network) Listen(ctx context.Context) (chan *Connection, error) {
	logger := log.FromContext(ctx)
	cconn := make(chan *Connection, 10)

	n.transports.Range(func(key, value interface{}) bool {
		tsp := value.(Transport)
		chConn, err := tsp.Listen(ctx)
		if err != nil {
			// TODO log
			return true
		}
		go func() {
			for {
				conn := <-chConn
				n.midLock.RLock()
				failed := false

				for _, mh := range n.middleware {
					conn, err = mh(ctx, conn)
					if err != nil {
						if errors.CausedBy(err, io.EOF) {
							break
						}
						logger.Error(
							"middleware failure",
							log.Error(err),
						)

						if conn != nil {
							conn.conn.Close() // nolint: errcheck
						}
						failed = true
						break
					}
				}
				n.midLock.RUnlock()

				if !failed {
					cconn <- conn
				}
			}
		}()
		return true
	})

	return cconn, nil
}

func (n *network) dialPeer(
	ctx context.Context,
	address string,
	localDiscoveryOnly bool,
) (*Connection, error) {
	logger := log.FromContext(ctx).With(
		log.String("address", address),
		log.Bool("localDiscoveryOnly", localDiscoveryOnly),
	)

	pkey := strings.Replace(address, "peer:", "", 1)
	key := crypto.PublicKey(pkey)

	if n.local.GetPeerKey().Equals(key) {
		return nil, errors.New("cannot dial our own peer")
	}

	logger.Debug("dialing peer")

	opts := []discovery.Option{}
	if localDiscoveryOnly {
		opts = append(opts, discovery.Local())
	}
	ps, err := n.discoverer.FindByPublicKey(ctx, key, opts...)
	if err != nil {
		return nil, err
	}

	logger.Debug("got peer infos", log.Int("n", len(ps)))

	for _, p := range ps {
		for _, addr := range p.Addresses {
			logger.Debug("trying to dial peer",
				log.String("peer.key", p.PublicKey().String()),
				log.Strings("peer.addresses", p.Addresses),
			)
			conn, err := n.Dial(ctx, addr)
			if err == nil {
				return conn, nil
			}
		}
	}

	return nil, ErrAllAddressesFailed
}
