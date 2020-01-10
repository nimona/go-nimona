package net

import (
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"nimona.io/pkg/context"
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
	Dial(ctx context.Context, peer *peer.Peer) (*Connection, error)
	Listen(ctx context.Context) (chan *Connection, error)

	AddMiddleware(handler MiddlewareHandler)
	AddTransport(tag string, tsp Transport)
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
	p *peer.Peer,
) (*Connection, error) {
	logger := log.FromContext(ctx).With(
		log.String("peer", p.PublicKey().String()),
		log.Strings("addresses", p.Addresses),
	)

	logger.Debug("dialing")

	for _, address := range p.Addresses {
		addressType := strings.Split(address, ":")[0]
		t, ok := n.transports.Load(addressType)
		if !ok {
			logger.Info("not sure how to dial",
				log.String("type", addressType),
			)
			return nil, ErrNoAddresses
		}

		trsp := t.(Transport)
		conn, err := trsp.Dial(ctx, address)
		if err != nil {
			return nil, err
		}

		for _, mh := range n.middleware {
			conn, err = mh(ctx, conn)
			if err != nil {
				return nil, err
			}
		}
		return conn, nil
	}

	logger.Error("could not dial peer")
	return nil, ErrAllAddressesFailed
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
