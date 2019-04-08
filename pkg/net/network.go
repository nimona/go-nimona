package net

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	igd "github.com/emersion/go-upnp-igd"
	"go.uber.org/zap"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/net/peer"
)

var (
	UseUPNP = false
)

func init() {
	UseUPNP, _ = strconv.ParseBool(os.Getenv("UPNP"))
}

// Network interface
type Network interface {
	Dial(ctx context.Context, address string) (*Connection, error)
	Listen(ctx context.Context, addrress string) (chan *Connection, error)

	GetPeerInfo() *peer.PeerInfo
	AttachMandate(m *crypto.Mandate) error
	AddMiddleware(handler MiddlewareHandler)
}

// New creates a new p2p network using an address book
func New(hostname string, discover discovery.Discoverer, local NetState,
	relayAddresses []string) (Network, error) {

	return &network{
		discoverer:     discover,
		hostname:       hostname,
		relayAddresses: relayAddresses,
		middleware:     []MiddlewareHandler{},
		local:          local,
	}, nil
}

// network allows dialing and listening for p2p connections
type network struct {
	discoverer     discovery.Discoverer
	hostname       string
	relayAddresses []string
	local          NetState
	middleware     []MiddlewareHandler
}

func (n *network) AddMiddleware(handler MiddlewareHandler) {
	n.middleware = append(n.middleware, handler)
}

// Dial to a peer and return a net.Conn or error
func (n *network) Dial(ctx context.Context, address string) (
	*Connection, error) {
	logger := log.Logger(ctx)

	addressType := strings.Split(address, ":")[0]
	switch addressType {
	case "peer":
		return n.dialPeer(ctx, address)
	case "tcps":
		return n.dialAddress(ctx, address)
	default:
		logger.Info("not sure how to dial",
			zap.String("address", address),
			zap.String("type", addressType))
	}

	return nil, ErrNoAddresses
}

// Listen on an address
// TODO do we need to return a listener?
func (n *network) Listen(ctx context.Context, address string) (chan *Connection, error) {
	logger := log.Logger(ctx).Named("network")
	cert, err := crypto.GenerateCertificate(n.local.GetPeerKey())
	if err != nil {
		return nil, err
	}

	now := time.Now()
	config := tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	config.NextProtos = []string{"nimona/1"} // TODO(geoah) is this of any actual use?
	config.Time = func() time.Time { return now }
	config.Rand = rand.Reader
	tcpListener, err := tls.Listen("tcp", address, &config)
	if err != nil {
		return nil, err
	}

	port := tcpListener.Addr().(*net.TCPAddr).Port
	logger.Info("Listening and service nimona", zap.Int("port", port))
	addresses := GetAddresses(tcpListener)
	devices := make(chan igd.Device, 10)

	if n.hostname != "" {
		addresses = append(addresses, fmtAddress(n.hostname, port))
	}

	if UseUPNP {
		logger.Info("Trying to find external IP and open port")
		go func() {
			if err := igd.Discover(devices, 2*time.Second); err != nil {
				logger.Error("could not discover devices", zap.Error(err))
			}
		}()
		for device := range devices {
			externalAddress, err := device.GetExternalIPAddress()
			if err != nil {
				logger.Error("could not get external ip", zap.Error(err))
				continue
			}
			desc := "nimona"
			ttl := time.Hour * 24 * 365
			if _, err := device.AddPortMapping(igd.TCP, port, port, desc, ttl); err != nil {
				logger.Error("could not add port mapping", zap.Error(err))
			} else {
				addresses = append(addresses, fmtAddress(externalAddress.String(), port))
			}
		}
	}

	logger.Info("Started listening", zap.Strings("addresses", addresses))
	n.local.AddAddress(addresses...)
	n.local.AddAddress(n.relayAddresses...)

	cconn := make(chan *Connection, 10)
	go func() {

		for {
			tcpConn, err := tcpListener.Accept()
			if err != nil {
				log.DefaultLogger.Warn(
					"could not accept connection", zap.Error(err))
				// TODO close conn?
				return
			}

			conn := &Connection{
				Conn:          tcpConn,
				RemotePeerKey: nil,
				IsIncoming:    true,
			}

			for _, mh := range n.middleware {
				conn, err = mh(ctx, conn)
			if err != nil {
					log.DefaultLogger.Error(
						"middleware failure", zap.Error((err)))
			}
			}

			cconn <- conn
		}
	}()

	return cconn, nil
}

// GetPeerInfo returns the local peer info
func (n *network) GetPeerInfo() *peer.PeerInfo {
	// TODO cache peer info and reuse
	return n.local.GetPeerInfo()
}

func (n *network) AttachMandate(m *crypto.Mandate) error {
	return n.local.AttachMandate(m)
}

func (n *network) dialPeer(ctx context.Context, address string) (
	*Connection, error) {
	logger := log.Logger(ctx)

	peerID := strings.Replace(address, "peer:", "", 1)
	if peerID == n.local.GetPeerKey().GetPublicKey().HashBase58() {
		return nil, errors.New("cannot dial our own peer")
	}
	logger.Debug("dialing peer", zap.String("peer", address))
	q := &peer.PeerInfoRequest{
		SignerKeyHash: peerID,
	}
	ps, err := n.discoverer.Discover(q)
	if err != nil {
		return nil, err
	}

	if len(ps) == 0 || len(ps[0].Addresses) == 0 {
		return nil, ErrNoAddresses
	}

	// TODO we should probably try all results
	for _, addr := range ps[0].Addresses {
		conn, err := n.Dial(ctx, addr)
		if err == nil {
			return conn, nil
		}
	}

	return nil, ErrAllAddressesFailed

}

func (n *network) dialAddress(ctx context.Context, address string) (
	*Connection, error) {

	// find dialer
	// execute middleware
	config := tls.Config{
		InsecureSkipVerify: true,
	}
	addr := strings.Replace(address, "tcps:", "", 1)
	dialer := net.Dialer{Timeout: time.Second}

	tcpConn, err := tls.DialWithDialer(&dialer, "tcp", addr, &config)
	if err != nil {
		return nil, err
	}

	if tcpConn == nil {
		return nil, ErrAllAddressesFailed
	}

	conn := &Connection{
		Conn:          tcpConn,
		RemotePeerKey: nil, // we don't really know who the other side is
		IsOutgoing:    true,
	}

	for _, mh := range n.middleware {
		conn, err = mh(ctx, conn)
	if err != nil {
		return nil, err
	}
	}

	return conn, nil
}
