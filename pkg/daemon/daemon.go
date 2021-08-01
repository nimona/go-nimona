package daemon

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"nimona.io/internal/net"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/feedmanager"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/network"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/preferences"
	"nimona.io/pkg/sqlobjectstore"
)

type (
	Daemon interface {
		Config() config.Config
		Preferences() preferences.Preferences
		Network() network.Network
		Resolver() resolver.Resolver
		ObjectStore() objectstore.Store
		ObjectManager() objectmanager.ObjectManager
		FeedManager() feedmanager.FeedManager
		// daemon specific methods
		Close()
	}
	daemon struct {
		config        config.Config
		preferences   preferences.Preferences
		configOptions []config.Option
		network       network.Network
		resolver      resolver.Resolver
		objectstore   objectstore.Store
		objectmanager objectmanager.ObjectManager
		feedmanager   feedmanager.FeedManager
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

	// construct new network
	ntw := network.New(
		ctx,
		network.WithPeerKey(cfg.Peer.PrivateKey),
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
		bootstrapPeer, err := s.GetConnectionInfo()
		if err != nil {
			return nil, fmt.Errorf("parsing bootstraps: %w", err)
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// add bootstrap peers as relays
	ntw.RegisterRelays(bootstrapPeers...)

	// construct preferences db
	pdb, err := sql.Open("sqlite3", filepath.Join(cfg.Path, "preferences.db"))
	if err != nil {
		return nil, fmt.Errorf("opening sql file for preferences: %w", err)
	}

	// construct preferences
	prf, err := preferences.NewSQLProvider(pdb)
	if err != nil {
		return nil, fmt.Errorf("constructing preferences provider: %w", err)
	}

	// construct object store
	db, err := sql.Open("sqlite3", filepath.Join(cfg.Path, "nimona.db"))
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
		str,
		resolver.WithBoostrapPeers(bootstrapPeers...),
	)

	// construct manager
	man := objectmanager.New(
		ctx,
		ntw,
		res,
		str,
	)

	// construct feed manager
	fdm, err := feedmanager.New(
		ctx,
		ntw,
		res,
		str,
		man,
	)
	if err != nil {
		return nil, fmt.Errorf("constructing feed manager, %w", err)
	}

	d.config = *cfg
	d.preferences = prf
	d.network = ntw
	d.resolver = res
	d.objectstore = str
	d.objectmanager = man
	d.feedmanager = fdm

	return d, nil
}

func (d *daemon) Config() config.Config {
	return d.config
}

func (d *daemon) Preferences() preferences.Preferences {
	return d.preferences
}

func (d *daemon) Network() network.Network {
	return d.network
}

func (d *daemon) Resolver() resolver.Resolver {
	return d.resolver
}

func (d *daemon) ObjectStore() objectstore.Store {
	return d.objectstore
}

func (d *daemon) ObjectManager() objectmanager.ObjectManager {
	return d.objectmanager
}

func (d *daemon) FeedManager() feedmanager.FeedManager {
	return d.feedmanager
}

func (d *daemon) Close() {
	if d.listener != nil {
		d.listener.Close() // nolint: errcheck
	}
}
