package net

import (
	"expvar"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/zserge/metric"

	"nimona.io/pkg/context"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/peer"
)

var UseUPNP = false

// TODO remove UseUPNP and replace with option
// nolint: gochecknoinits
func init() {
	UseUPNP, _ = strconv.ParseBool(os.Getenv("UPNP"))

	connConnOutCounter := metric.NewCounter("2m1s", "15m30s", "1h1m")
	expvar.Publish("nm:net.conn.out", connConnOutCounter)

	connConnIncCounter := metric.NewCounter("2m1s", "15m30s", "1h1m")
	expvar.Publish("nm:net.conn.in", connConnIncCounter)

	connDialCounter := metric.NewCounter("2m1s", "15m30s", "1h1m")
	expvar.Publish("nm:net.dial", connDialCounter)

	connBlacklistCounter := metric.NewCounter("2m1s", "15m30s", "1h1m")
	expvar.Publish("nm:net.conn.dial.blacklist", connBlacklistCounter)
}

// Network interface
type Network interface {
	Dial(ctx context.Context, peer *peer.Peer) (*Connection, error)
	Listen(ctx context.Context) (chan *Connection, error)

	AddMiddleware(handler MiddlewareHandler)
	AddTransport(tag string, tsp Transport)
}

// New creates a new p2p network using an address book
func New(
	discover discovery.Discoverer,
	local *peer.LocalPeer,
) (Network, error) {
	return &network{
		discoverer: discover,
		middleware: []MiddlewareHandler{},
		local:      local,
		midLock:    &sync.RWMutex{},
		transports: &sync.Map{},
		attempts:   newAttemptsMap(),
		blacklist:  cache.New(time.Second*5, time.Second*60),
	}, nil
}

// network allows dialing and listening for p2p connections
type network struct {
	discoverer discovery.Discoverer
	local      *peer.LocalPeer
	midLock    *sync.RWMutex
	transports *sync.Map
	middleware []MiddlewareHandler
	attempts   *attemptsMap
	blacklist  *cache.Cache
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
	expvar.Get("nm:net.dial").(metric.Metric).Add(1)

	// keep a flag on whether all addresses where blacklisted so we can return
	// an ErrMissingSignature error
	allBlacklisted := true

	// go through all addresses and try to dial them
	for _, address := range p.Addresses {
		// check if address is currently blacklisted
		if _, blacklisted := n.blacklist.Get(address); blacklisted {
			logger.Debug("address is blacklisted, skipping")
			continue
		}
		// get protocol from address
		addressType := strings.Split(address, ":")[0]
		t, ok := n.transports.Load(addressType)
		if !ok {
			logger.Debug("not sure how to dial",
				log.String("type", addressType),
			)
			continue
		}

		// reset blacklist flag
		allBlacklisted = false

		// dial address
		trsp := t.(Transport)
		conn, err := trsp.Dial(ctx, address)
		if err != nil {
			// blacklist address
			expvar.Get("nm:net.conn.dial.blacklist").(metric.Metric).Add(1)
			attempts, backoff := n.exponentialyBlacklist(address)
			logger.Error("could not dial address, blacklisting",
				log.Int("failedAttempts", attempts),
				log.String("backoff", backoff.String()),
				log.String("type", addressType),
				log.Error(err),
			)
			continue
		}

		// pass connection to all middleware
		var merr error
		for _, mh := range n.middleware {
			conn, merr = mh(ctx, conn)
			if merr != nil {
				break
			}
		}
		if merr != nil {
			conn.Close() // nolint: errcheck
			logger.Info("could not handle middleware",
				log.String("type", addressType),
				log.Error(err),
			)
			continue
		}

		// at this point we consider the connection successful, so we can
		// reset the failed attempts
		n.attempts.Put(address, 0)
		n.attempts.Put(p.PublicKey().String(), 0)

		expvar.Get("nm:net.conn.out").(metric.Metric).Add(1)

		return conn, nil
	}

	err := ErrAllAddressesFailed
	if allBlacklisted {
		err = ErrAllAddressesBlacklisted
	}

	logger.Error("could not dial peer", log.Error(err))
	return nil, err
}

func (n *network) exponentialyBlacklist(k string) (int, time.Duration) {
	baseBackoff := float64(time.Second * 1)
	maxBackoff := float64(time.Minute * 10)
	attempts, _ := n.attempts.Get(k)
	attempts++
	backoff := baseBackoff * math.Pow(1.5, float64(attempts))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	n.attempts.Put(k, attempts)
	n.blacklist.Set(k, attempts, time.Duration(backoff))
	return attempts, time.Duration(backoff)
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

				expvar.Get("nm:net.conn.in").(metric.Metric).Add(1)

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
