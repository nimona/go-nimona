package daemon

import (
	"database/sql"
	"fmt"
	"path"

	"nimona.io/pkg/network"

	// required for sqlobjectstore
	_ "github.com/mattn/go-sqlite3"

	"nimona.io/internal/daemon/config"
	"nimona.io/internal/nat"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"
)

type Daemon struct {
	Resolver      resolver.Resolver
	Network       network.Network
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

	local := localpeer.New()

	// add identity key to local info
	if cfg.Peer.IdentityKey != "" {
		local.PutPrimaryIdentityKey(cfg.Peer.IdentityKey)
	}

	if cfg.Peer.AnnounceHostname != "" {
		local.PutAddresses(fmt.Sprintf(
			"tcps:%s:%d",
			cfg.Peer.AnnounceHostname,
			cfg.Peer.TCPPort,
		))
	}

	// add relay peers
	for i, rp := range cfg.Peer.RelayKeys {
		local.PutRelays(
			&peer.Peer{
				Metadata: object.Metadata{
					Owner: crypto.PublicKey(rp),
				},
				Addresses: []string{
					cfg.Peer.RelayAddresses[i],
				},
			},
		)
	}

	local.PutPrimaryPeerKey(cfg.Peer.PeerKey)

	// get temp bootstrap peers from cfg
	bootstrapPeers := make([]*peer.Peer, len(cfg.Peer.BootstrapKeys))
	for i, k := range cfg.Peer.BootstrapKeys {
		bootstrapPeers[i] = &peer.Peer{
			Metadata: object.Metadata{
				Owner: crypto.PublicKey(k),
			},
		}
	}
	for i, a := range cfg.Peer.BootstrapAddresses {
		bootstrapPeers[i].Addresses = []string{a}
	}

	net := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	// construct resolver
	rs := resolver.New(
		ctx,
		net,
		resolver.WithBoostrapPeers(bootstrapPeers),
	)

	// construct objectmanager
	om := objectmanager.New(
		ctx,
		net,
		rs,
		st,
	)

	if _, err := net.Listen(
		ctx,
		fmt.Sprintf("0.0.0.0:%d", cfg.Peer.TCPPort),
	); err != nil {
		return nil, err
	}

	if cfg.Peer.UPNP {
		addr, _, _ := nat.MapExternalPort(cfg.Peer.TCPPort) // nolint: errcheck
		local.PutAddresses("tcps:" + addr)
	}

	return &Daemon{
		Network:       net,
		Resolver:      rs,
		Store:         st,
		ObjectManager: om,
	}, nil
}
