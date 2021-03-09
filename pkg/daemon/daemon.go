package daemon

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"nimona.io/internal/net"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

type (
	Daemon interface {
		Config() config.Config
		Network() network.Network
		Resolver() resolver.Resolver
		LocalPeer() localpeer.LocalPeer
		ObjectStore() objectstore.Store
		ObjectManager() objectmanager.ObjectManager
	}
	daemon struct {
		config        config.Config
		configOptions []config.Option
		network       network.Network
		resolver      resolver.Resolver
		localpeer     localpeer.LocalPeer
		objectstore   objectstore.Store
		objectmanager objectmanager.ObjectManager
		// internal
		listener net.Listener
	}
	Option func(d *daemon) error
)

func New(ctx context.Context, opts ...Option) (Daemon, error) {
	d := &daemon{}

	// apply options
	for _, o := range opts {
		if err := o(d); err != nil {
			return nil, err
		}
	}

	// load config with given options
	cfg, err := config.New(d.configOptions...)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.PutPrimaryPeerKey(cfg.Peer.PrivateKey)

	// construct new network
	ntw := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	if cfg.Peer.BindAddress != "" {
		// start listening
		lis, err := ntw.Listen(
			ctx,
			cfg.Peer.BindAddress,
			network.ListenOnLocalIPs,
			// network.ListenOnExternalPort,
		)
		if err != nil {
			return nil, fmt.Errorf("listening: %w", err)
		}
		d.listener = lis
	}

	// convert shorthands into connection infos
	bootstrapPeers := []*peer.ConnectionInfo{}
	for _, s := range cfg.Peer.Bootstraps {
		bootstrapPeer, err := s.ConnectionInfo()
		if err != nil {
			return nil, fmt.Errorf("parsing bootstraps: %w", err)
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// add bootstrap peers as relays
	local.PutRelays(bootstrapPeers...)

	// construct object store
	db, err := sql.Open("sqlite3", filepath.Join(cfg.Path, "chat.db"))
	if err != nil {
		return nil, fmt.Errorf("opening sql file: %w", err)
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		return nil, fmt.Errorf("starting sql store: %w", err)
	}

	// construct new resolver
	res := resolver.New(
		ctx,
		ntw,
		resolver.WithBoostrapPeers(bootstrapPeers...),
	)

	// construct manager
	man := objectmanager.New(
		ctx,
		ntw,
		res,
		str,
	)

	d.config = *cfg
	d.network = ntw
	d.resolver = res
	d.localpeer = local
	d.objectstore = str
	d.objectmanager = man

	return d, nil
}

func (d *daemon) Config() config.Config {
	return d.config
}

func (d *daemon) Network() network.Network {
	return d.network
}

func (d *daemon) Resolver() resolver.Resolver {
	return d.resolver
}

func (d *daemon) LocalPeer() localpeer.LocalPeer {
	return d.localpeer
}

func (d *daemon) ObjectStore() objectstore.Store {
	return d.objectstore
}

func (d *daemon) ObjectManager() objectmanager.ObjectManager {
	return d.objectmanager
}

func (d *daemon) Close() {
	if d.listener != nil {
		d.listener.Close() // nolint: errcheck
	}
}
