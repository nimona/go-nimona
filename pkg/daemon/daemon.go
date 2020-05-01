package daemon

import (
	"database/sql"
	"fmt"
	"path"

	"nimona.io/pkg/nat"

	// required for sqlobjectstore
	_ "github.com/mattn/go-sqlite3"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon/config"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/hyperspace"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
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

	// add identity key to local info
	if cfg.Peer.IdentityKey != "" {
		keychain.DefaultKeychain.Put(
			keychain.IdentityKey,
			cfg.Peer.IdentityKey,
		)
	}

	if cfg.Peer.AnnounceHostname != "" {
		eventbus.DefaultEventbus.Publish(
			eventbus.NetworkAddressAdded{
				Address: fmt.Sprintf(
					"%s:%d",
					cfg.Peer.AnnounceHostname,
					cfg.Peer.TCPPort,
				),
			},
		)
	}

	// add relay peers
	for _, rp := range cfg.Peer.RelayAddresses {
		eventbus.DefaultEventbus.Publish(
			eventbus.RelayAdded{
				PublicKey: rp,
			},
		)
	}

	keychain.DefaultKeychain.Put(
		keychain.PrimaryPeerKey,
		cfg.Peer.PeerKey,
	)

	network := net.New(
		net.WithKeychain(keychain.DefaultKeychain),
		net.WithEventBus(eventbus.DefaultEventbus),
	)

	// construct exchange
	ex, err := exchange.New(
		ctx,
		eventbus.DefaultEventbus,
		keychain.DefaultKeychain,
		network,
		st,
		ps,
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
	hs, err := hyperspace.New(
		ctx,
		ps,
		keychain.DefaultKeychain,
		eventbus.DefaultEventbus,
		ex,
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
		keychain.DefaultKeychain,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not construct orchestrator"), err)
	}

	// add hyperspace provider
	if err := ps.AddDiscoverer(hs); err != nil {
		return nil, errors.Wrap(errors.New("could not add hyperspace provider"), err)
	}

	if _, err := network.Listen(
		ctx,
		fmt.Sprintf("0.0.0.0:%d", cfg.Peer.TCPPort),
	); err != nil {
		return nil, err
	}

	if cfg.Peer.UPNP {
		nat.MapExternalPort(cfg.Peer.TCPPort) // nolint: errcheck
	}

	return &Daemon{
		Net:          network,
		Discovery:    ps,
		Exchange:     ex,
		Store:        st,
		Orchestrator: or,
	}, nil
}
