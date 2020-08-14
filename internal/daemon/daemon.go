package daemon

import (
	"database/sql"
	"fmt"
	"path"

	// required for sqlobjectstore
	_ "github.com/mattn/go-sqlite3"

	"nimona.io/internal/daemon/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/nat"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"
)

type Daemon struct {
	Net           net.Network
	Resolver      resolver.Resolver
	Exchange      exchange.Exchange
	Store         *sqlobjectstore.Store
	ObjectManager objectmanager.ObjectManager
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
	for i, rp := range cfg.Peer.RelayKeys {
		eventbus.DefaultEventbus.Publish(
			eventbus.RelayAdded{
				Peer: &peer.Peer{
					Metadata: object.Metadata{
						Owners: []crypto.PublicKey{
							crypto.PublicKey(rp),
						},
					},
					Addresses: []string{
						cfg.Peer.RelayAddresses[i],
					},
				},
			},
		)
	}

	keychain.DefaultKeychain.Put(
		keychain.PrimaryPeerKey,
		cfg.Peer.PeerKey,
	)

	// get temp bootstrap peers from cfg
	bootstrapPeers := make([]*peer.Peer, len(cfg.Peer.BootstrapKeys))
	for i, k := range cfg.Peer.BootstrapKeys {
		bootstrapPeers[i] = &peer.Peer{
			Metadata: object.Metadata{
				Owners: []crypto.PublicKey{
					crypto.PublicKey(k),
				},
			},
		}
	}
	for i, a := range cfg.Peer.BootstrapAddresses {
		bootstrapPeers[i].Addresses = []string{a}
	}

	// construct resolver
	rs := resolver.New(
		ctx,
		resolver.WithExchange(exchange.DefaultExchange),
		resolver.WithKeychain(keychain.DefaultKeychain),
		resolver.WithEventbus(eventbus.DefaultEventbus),
		resolver.WithBoostrapPeers(bootstrapPeers),
	)

	// construct objectmanager
	om := objectmanager.New(
		ctx,
		objectmanager.WithExchange(exchange.DefaultExchange),
		objectmanager.WithStore(st),
	)

	if _, err := net.DefaultNetwork.Listen(
		ctx,
		fmt.Sprintf("0.0.0.0:%d", cfg.Peer.TCPPort),
	); err != nil {
		return nil, err
	}

	if cfg.Peer.UPNP {
		nat.MapExternalPort(cfg.Peer.TCPPort) // nolint: errcheck
	}

	return &Daemon{
		Net:           net.DefaultNetwork,
		Exchange:      exchange.DefaultExchange,
		Resolver:      rs,
		Store:         st,
		ObjectManager: om,
	}, nil
}
