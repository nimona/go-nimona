package daemon

import (
	"database/sql"
	"path/filepath"

	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
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
		return nil, errors.Wrap(errors.New("error loading config"), err)
	}

	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.PutPrimaryPeerKey(cfg.Peer.PrivateKey)

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
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
			return nil, errors.Wrap(errors.New("error while listening"), err)
		}
		defer lis.Close() // nolint: errcheck
	}

	// convert shorthands into connection infos
	bootstrapPeers := []*peer.ConnectionInfo{}
	for _, s := range cfg.Peer.Bootstraps {
		bootstrapPeer, err := s.ConnectionInfo()
		if err != nil {
			return nil, errors.Wrap(errors.New("error parsing bootstraps"), err)
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// add bootstrap peers as relays
	local.PutRelays(bootstrapPeers...)

	// construct object store
	db, err := sql.Open("sqlite3", filepath.Join(cfg.Path, "chat.db"))
	if err != nil {
		return nil, errors.Wrap(errors.New("error opening sql file"), err)
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		return nil, errors.Wrap(errors.New("error starting sql store"), err)
	}

	// construct new resolver
	res := resolver.New(
		ctx,
		net,
		resolver.WithBoostrapPeers(bootstrapPeers...),
	)

	// construct manager
	man := objectmanager.New(
		ctx,
		net,
		res,
		str,
	)

	d.config = *cfg
	d.network = net
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
