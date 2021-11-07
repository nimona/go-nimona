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
	"nimona.io/pkg/keystream"
	"nimona.io/pkg/mesh"
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
		Network() mesh.Mesh
		Resolver() resolver.Resolver
		ObjectStore() objectstore.Store
		ObjectManager() objectmanager.ObjectManager
		FeedManager() feedmanager.FeedManager
		KeyStreamManager() keystream.Manager
		// daemon specific methods
		Close()
	}
	daemon struct {
		config          config.Config
		preferences     preferences.Preferences
		configOptions   []config.Option
		mesh            mesh.Mesh
		resolver        resolver.Resolver
		objectstore     objectstore.Store
		objectmanager   objectmanager.ObjectManager
		feedmanager     feedmanager.FeedManager
		keystreamanager keystream.Manager
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

	// construct new mesh
	inet := net.New(cfg.Peer.PrivateKey)
	msh := mesh.New(
		ctx,
		inet,
		cfg.Peer.PrivateKey,
	)

	if cfg.Peer.BindAddress != "" {
		// start listening
		lis, err := msh.Listen(
			ctx,
			cfg.Peer.BindAddress,
			mesh.ListenOnLocalIPs,
			// mesh.ListenOnExternalPort,
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
	msh.RegisterRelays(bootstrapPeers...)

	// construct preferences db
	pdb, err := sql.Open("sqlite", filepath.Join(cfg.Path, "preferences.db"))
	if err != nil {
		return nil, fmt.Errorf("opening sql file for preferences: %w", err)
	}

	// construct preferences
	prf, err := preferences.NewSQLProvider(pdb)
	if err != nil {
		return nil, fmt.Errorf("constructing preferences provider: %w", err)
	}

	// construct object store
	db, err := sql.Open("sqlite", filepath.Join(cfg.Path, "nimona.db"))
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
		inet,
		cfg.Peer.PrivateKey,
		str,
		resolver.WithBoostrapPeers(bootstrapPeers...),
	)

	// register resolver
	msh.RegisterResolver(res)

	// construct manager
	man := objectmanager.New(
		ctx,
		msh,
		res,
		str,
	)

	// construct feed manager
	fdm, err := feedmanager.New(
		ctx,
		msh,
		res,
		str,
		man,
	)
	if err != nil {
		return nil, fmt.Errorf("constructing feed manager, %w", err)
	}

	// construct key stream manager
	ksm, err := keystream.NewKeyManager(
		msh,
		str,
	)
	if err != nil {
		return nil, fmt.Errorf("constructing keystream manager, %w", err)
	}

	d.config = *cfg
	d.preferences = prf
	d.mesh = msh
	d.resolver = res
	d.objectstore = str
	d.objectmanager = man
	d.feedmanager = fdm
	d.keystreamanager = ksm

	return d, nil
}

func (d *daemon) Config() config.Config {
	return d.config
}

func (d *daemon) Preferences() preferences.Preferences {
	return d.preferences
}

func (d *daemon) Network() mesh.Mesh {
	return d.mesh
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

func (d *daemon) KeyStreamManager() keystream.Manager {
	return d.keystreamanager
}

func (d *daemon) Close() {
	if d.listener != nil {
		d.listener.Close() // nolint: errcheck
	}
}
