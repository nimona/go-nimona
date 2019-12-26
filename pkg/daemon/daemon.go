package daemon

import (
	ssql "database/sql"
	"fmt"
	"path"
	"strings"

	"nimona.io/pkg/errors"

	_ "github.com/mattn/go-sqlite3"

	"nimona.io/pkg/store/sql"
	"nimona.io/pkg/context"
	"nimona.io/pkg/daemon/config"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/hyperspace"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/middleware/handshake"
	"nimona.io/pkg/net"
	"nimona.io/pkg/orchestrator"
	"nimona.io/pkg/peer"
)

type Daemon struct {
	Net          net.Network
	Discovery    discovery.Discoverer
	Exchange     exchange.Exchange
	Store        *sql.Store
	Orchestrator orchestrator.Orchestrator
	LocalPeer    *peer.LocalPeer
}

func New(ctx context.Context, config *config.Config) (*Daemon, error) {
	// make sure relays are valid
	for i, addr := range config.Peer.RelayAddresses {
		if !strings.HasPrefix(addr, "relay:") {
			config.Peer.RelayAddresses[i] = "relay:" + addr
		}
	}

	// construct discoverer
	discoverer := discovery.NewDiscoverer()

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

	network, err := net.New(discoverer, localInfo)
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
		discoverer,
	)

	// add middleware to network
	network.AddMiddleware(handshakeMiddleware.Handle())

	// construct graph store
	db, err := ssql.Open("sqlite3", path.Join(config.Path, "sqlite3.db"))
	if err != nil {
		return nil, errors.Wrap(errors.New("could not open sql file"), err)
	}

	store, err := sql.New(db)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not start sql store"), err)
	}

	// construct exchange
	exchange, err := exchange.New(
		ctx,
		config.Peer.PeerKey,
		network,
		store,
		discoverer,
		localInfo,
	)
	if err != nil {
		return nil, errors.Wrap(errors.New("could not construct exchange"), err)
	}

	// construct hyperspace discoverer
	hyperspace, err := hyperspace.NewDiscoverer(
		ctx,
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
	if err := discoverer.AddProvider(hyperspace); err != nil {
		return nil, errors.Wrap(errors.New("could not add hyperspace provider"), err)
	}

	return &Daemon{
		Net:          network,
		Discovery:    discoverer,
		Exchange:     exchange,
		Store:        store,
		Orchestrator: orchestrator,
		LocalPeer:    localInfo,
	}, nil
}
