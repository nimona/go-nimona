package daemon

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"nimona.io/internal/connmanager"
	"nimona.io/pkg/config"
	"nimona.io/pkg/configstore"
	"nimona.io/pkg/context"
	hresolver "nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/keystream"
	"nimona.io/pkg/network"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

type (
	Daemon interface {
		Config() config.Config
		ConfigStore() configstore.Store
		Network() network.Network
		Resolver() resolver.Resolver
		ObjectStore() objectstore.Store
		ObjectManager() objectmanager.ObjectManager
		KeyStreamManager() keystream.Manager
		StreamManager() stream.Manager
		// daemon specific methods
		Close()
	}
	daemon struct {
		config          config.Config
		configstore     configstore.Store
		configOptions   []config.Option
		network         network.Network
		resolver        resolver.Resolver
		objectstore     objectstore.Store
		objectmanager   objectmanager.ObjectManager
		streammanager   stream.Manager
		keystreamanager keystream.Manager
		// internal
		listener connmanager.Listener
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

	// construct configstore db
	pdb, err := sql.Open("sqlite", filepath.Join(cfg.Path, "config.sqlite"))
	if err != nil {
		return nil, fmt.Errorf("opening sql file for configstore: %w", err)
	}

	// construct configstore
	prf, err := configstore.NewSQLProvider(pdb)
	if err != nil {
		return nil, fmt.Errorf("constructing configstore provider: %w", err)
	}

	// construct object store
	db, err := sql.Open("sqlite", filepath.Join(cfg.Path, "object.sqlite"))
	if err != nil {
		return nil, fmt.Errorf("opening sql file: %w", err)
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		return nil, fmt.Errorf("starting sql store: %w", err)
	}

	// construct new network
	net := network.New(
		ctx,
		cfg.Peer.PrivateKey,
		str,
	)

	if cfg.Peer.BindAddress != "" {
		// start listening
		lis, err := net.Listen(
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
	net.RegisterRelays(bootstrapPeers...)

	// construct new composite resolver
	res := resolver.New()

	// construct new stream manager
	sm, err := stream.NewManager(
		ctx,
		net,
		res,
		str,
	)
	if err != nil {
		return nil, fmt.Errorf("constructing stream manager, %w", err)
	}

	// construct key stream manager
	ksm, err := keystream.NewKeyManager(
		net,
		str,
		sm,
		prf,
	)
	if err != nil {
		return nil, fmt.Errorf("constructing keystream manager, %w", err)
	}

	// construct new resolver
	hres := hresolver.New(
		ctx,
		net,
		cfg.Peer.PrivateKey,
		str,
		ksm,
		hresolver.WithBoostrapPeers(bootstrapPeers...),
	)

	// register resolver
	res.RegisterResolver(hres)

	// register resolver
	net.RegisterResolver(res)

	// construct manager
	man := objectmanager.New(
		ctx,
		net,
		res,
		str,
	)
	if err != nil {
		return nil, fmt.Errorf("constructing object manager, %w", err)
	}

	d.config = *cfg
	d.configstore = prf
	d.network = net
	d.resolver = res
	d.objectstore = str
	d.objectmanager = man
	d.keystreamanager = ksm
	d.streammanager = sm

	return d, nil
}

func (d *daemon) Config() config.Config {
	return d.config
}

func (d *daemon) ConfigStore() configstore.Store {
	return d.configstore
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

func (d *daemon) KeyStreamManager() keystream.Manager {
	return d.keystreamanager
}

func (d *daemon) StreamManager() stream.Manager {
	return d.streammanager
}

func (d *daemon) Close() {
	if d.listener != nil {
		d.listener.Close() // nolint: errcheck
	}
}
