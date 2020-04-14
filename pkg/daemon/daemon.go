package daemon

import (
	"database/sql"
	"fmt"
	"path"

	// required for sqlobjectstore
	_ "github.com/mattn/go-sqlite3"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon/config"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/hyperspace"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/middleware/handshake"
	"nimona.io/pkg/net"
	"nimona.io/pkg/orchestrator"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

type Daemon struct {
	Net          net.Network
	Discovery    discovery.Discoverer
	Exchange     exchange.Exchange
	Store        *sqlobjectstore.Store
	Orchestrator orchestrator.Orchestrator
	LocalPeer    *peer.LocalPeer
}

func New(ctx context.Context, cfg *config.Config) (*Daemon, error) {
	// construct object store
	db, err := sql.Open("sqlite3", path.Join(cfg.Path, "sqlite3.db"))
	if err != nil {
		return nil, errors.Wrap(errors.New("could not open sql file"), err)
	}

	st, err := sqlobjectstore.New(db)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not start sql store"), err)
	}

	// construct peerstore
	ps := discovery.NewPeerStorer(st)

	// construct local info
	li, err := peer.NewLocalPeer(
		cfg.Peer.AnnounceHostname,
		cfg.Peer.PeerKey,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not create local info"), err)
	}

	// add content types
	li.AddContentTypes(cfg.Peer.ContentTypes...)

	// add identity key to local info
	if cfg.Peer.IdentityKey != "" {
		if err := li.AddIdentityKey(cfg.Peer.IdentityKey); err != nil {
			return nil, errors.Wrap(errors.New("could not register identity key"), err)
		}
	}

	// add relay peers to local info
	for _, rp := range cfg.Peer.RelayAddresses {
		li.AddRelays(rp)
	}

	network, err := net.New(ps, li)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not create network"), err)
	}

	// construct tcp transport
	tcpTransport := net.NewTCPTransport(
		li,
		fmt.Sprintf("0.0.0.0:%d", cfg.Peer.TCPPort),
	)

	// add transports to network
	network.AddTransport("tcps", tcpTransport)

	// construct handshake
	handshakeMiddleware := handshake.New(
		li,
		ps,
	)

	// add middleware to network
	network.AddMiddleware(handshakeMiddleware.Handle())

	// construct exchange
	ex, err := exchange.New(
		ctx,
		cfg.Peer.PeerKey,
		network,
		st,
		ps,
		li,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not construct exchange"), err)
	}

	// get temp bootstrap peers from cfg
	bootstrapPeers := make([]*peer.Peer, len(cfg.Peer.BootstrapKeys))
	for i, k := range cfg.Peer.BootstrapKeys {
		bootstrapPeers[i] = &peer.Peer{
			Owners: []crypto.PublicKey{
				crypto.PublicKey(k),
			},
		}
	}
	for i, a := range cfg.Peer.BootstrapAddresses {
		bootstrapPeers[i].Addresses = []string{a}
	}

	// construct hyperspace peerstore
	hs, err := hyperspace.NewDiscoverer(
		ctx,
		ps,
		ex,
		li,
		bootstrapPeers,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not construct hyperspace"), err)
	}

	// construct orchestrator
	or, err := orchestrator.New(
		st,
		ex,
		nil,
		li,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not construct orchestrator"), err)
	}

	// add hyperspace provider
	if err := ps.AddDiscoverer(hs); err != nil {
		return nil, errors.Wrap(errors.New("could not add hyperspace provider"), err)
	}

	return &Daemon{
		Net:          network,
		Discovery:    ps,
		Exchange:     ex,
		Store:        st,
		Orchestrator: or,
		LocalPeer:    li,
	}, nil
}
