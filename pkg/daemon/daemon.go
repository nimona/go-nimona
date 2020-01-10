package daemon

import (
	"database/sql"
	"fmt"
	"path"

	_ "github.com/mattn/go-sqlite3"

	"nimona.io/pkg/context"
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

func New(ctx context.Context, config *config.Config) (*Daemon, error) {
	// construct object store
	db, err := sql.Open("sqlite3", path.Join(config.Path, "sqlite3.db"))
	if err != nil {
		return nil, errors.Wrap(errors.New("could not open sql file"), err)
	}

	store, err := sqlobjectstore.New(db)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not start sql store"), err)
	}

	// construct peerstore
	peerstore := discovery.NewPeerStorer(store)

	// construct local info
	localInfo, err := peer.NewLocalPeer(
		config.Peer.AnnounceHostname,
		config.Peer.PeerKey,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not create local info"), err)
	}

	// add content types
	localInfo.AddContentTypes(config.Peer.ContentTypes...)

	// add identity key to local info
	if config.Peer.IdentityKey != "" {
		if err := localInfo.AddIdentityKey(config.Peer.IdentityKey); err != nil {
			return nil, errors.Wrap(errors.New("could not register identity key"), err)
		}
	}

	// add relay addresses to local info
	localInfo.AddAddress("relay", config.Peer.RelayAddresses)

	network, err := net.New(peerstore, localInfo)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not create network"), err)
	}

	// construct tcp transport
	tcpTransport := net.NewTCPTransport(
		localInfo,
		fmt.Sprintf("0.0.0.0:%d", config.Peer.TCPPort),
	)

	// add transports to network
	network.AddTransport("tcps", tcpTransport)

	// construct handshake
	handshakeMiddleware := handshake.New(
		localInfo,
		peerstore,
	)

	// add middleware to network
	network.AddMiddleware(handshakeMiddleware.Handle())

	// construct exchange
	exchange, err := exchange.New(
		ctx,
		config.Peer.PeerKey,
		network,
		store,
		peerstore,
		localInfo,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not construct exchange"), err)
	}

	// construct hyperspace peerstore
	hyperspace, err := hyperspace.NewDiscoverer(
		ctx,
		peerstore,
		exchange,
		localInfo,
		config.Peer.BootstrapAddresses,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not construct hyperspace"), err)
	}

	// construct orchestrator
	orchestrator, err := orchestrator.New(
		store,
		exchange,
		nil,
		localInfo,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not construct orchestrator"), err)
	}

	// add hyperspace provider
	if err := peerstore.AddDiscoverer(hyperspace); err != nil {
		return nil, errors.Wrap(errors.New("could not add hyperspace provider"), err)
	}

	return &Daemon{
		Net:          network,
		Discovery:    peerstore,
		Exchange:     exchange,
		Store:        store,
		Orchestrator: orchestrator,
		LocalPeer:    localInfo,
	}, nil
}
