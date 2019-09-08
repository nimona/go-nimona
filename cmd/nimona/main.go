package main

import (
	"fmt"
	"os"
	"strings"

	"nimona.io/internal/api"
	"nimona.io/internal/context"
	"nimona.io/internal/log"
	"nimona.io/internal/store/graph"
	"nimona.io/internal/store/kv"
	"nimona.io/internal/version"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/hyperspace"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/middleware/handshake"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object/dag"
	"nimona.io/pkg/peer"
)

func main() {
	cfgFile := os.Getenv("NIMONA_CONFIG")
	if cfgFile == "" {
		cfgFile = ".nimona/config.json"
	}

	ctx := context.New(
		context.WithCorrelationID("nimona"),
	)

	logger := log.FromContext(ctx).With(
		log.String("configFile", cfgFile),
		log.String("build.version", version.Version),
		log.String("build.commit", version.Commit),
		log.String("build.timestamp", version.Date),
	)

	// load config
	logger.Info("loading config file")
	config, err := LoadConfig(cfgFile)
	if err != nil {
		logger.Fatal("could not load config file", log.Error(err))
	}

	// create peer key pair if it does not exist
	if config.Daemon.PeerKey == nil {
		logger.Info("creating new peer key pair")
		peerKey, err := crypto.GenerateKey()
		if err != nil {
			logger.Fatal("could not generate peer key", log.Error(err))
		}
		config.Daemon.PeerKey = peerKey
	}

	// create identity key pair if it does not exist
	// TODO this is temporary
	if config.Daemon.IdentityKey == nil {
		logger.Info("creating new identity key pair")
		identityKey, err := crypto.GenerateKey()
		if err != nil {
			logger.Fatal("could not generate identity key", log.Error(err))
		}
		config.Daemon.IdentityKey = identityKey
	}

	// make sure relays are valid
	for i, addr := range config.Daemon.RelayAddresses {
		if !strings.HasPrefix(addr, "relay:") {
			config.Daemon.RelayAddresses[i] = "relay:" + addr
		}
	}

	// update config
	if err := UpdateConfig(cfgFile, config); err != nil {
		logger.Fatal("could not update config", log.Error(err))
	}

	logger.Info("loaded config", log.Any("config", config))

	// start daemon

	// construct discoverer
	discoverer := discovery.NewDiscoverer()

	// construct local info
	localInfo, err := peer.NewLocalPeer(
		config.Daemon.AnnounceHostname,
		config.Daemon.PeerKey,
	)
	if err != nil {
		logger.Fatal("could not create local info", log.Error(err))
	}

	// add identity key to local info
	if err := localInfo.AddIdentityKey(config.Daemon.IdentityKey); err != nil {
		logger.Fatal("could not register identity key", log.Error(err))
	}

	// add relay addresses to local info
	localInfo.AddAddress("relay", config.Daemon.RelayAddresses)

	network, err := net.New(discoverer, localInfo)
	if err != nil {
		logger.Fatal("could not create network", log.Error(err))
	}

	// construct tcp transport
	tcpTransport := net.NewTCPTransport(
		localInfo,
		fmt.Sprintf("0.0.0.0:%d", config.Daemon.TCPPort),
	)

	// construct http transport
	httpTransport := net.NewHTTPTransport(
		localInfo,
		fmt.Sprintf("0.0.0.0:%d", config.Daemon.HTTPPort),
	)

	// add transports to network
	network.AddTransport("tcps", tcpTransport)
	network.AddTransport("https", httpTransport)

	// construct handshake
	handshakeMiddleware := handshake.New(
		localInfo,
		discoverer,
	)

	// add middleware to network
	network.AddMiddleware(handshakeMiddleware.Handle())

	// construct graph store
	graphStore := graph.New(kv.NewMemory())

	// construct exchange
	exchange, err := exchange.New(
		ctx,
		config.Daemon.PeerKey,
		network,
		graphStore,
		discoverer,
		localInfo,
	)
	if err != nil {
		logger.Fatal("could not construct exchange", log.Error(err))
	}

	// construct dag
	dag, err := dag.New(
		graphStore,
		exchange,
		nil,
		localInfo,
	)
	if err != nil {
		logger.Fatal("could not construct dag", log.Error(err))
	}

	// construct hyperspace discoverer
	hyperspace, err := hyperspace.NewDiscoverer(
		ctx,
		network,
		exchange,
		localInfo,
		config.Daemon.BootstrapAddresses,
	)
	if err != nil {
		logger.Fatal("could not construct hyperspace", log.Error(err))
	}

	// add hyperspace provider
	if err := discoverer.AddProvider(hyperspace); err != nil {
		logger.Fatal("could not add hyperspace provider", log.Error(err))
	}

	// print some info
	fmt.Println("Started daemon")
	fmt.Println("* Peer fingerprint:\n  *", config.Daemon.PeerKey.Fingerprint())
	ik := config.Daemon.IdentityKey
	if ik != nil {
		fmt.Println("* Identity fingerprint:\n  *", ik.PublicKey.Fingerprint())
	}
	peerAddresses := localInfo.GetAddresses()
	fmt.Println("* Peer addresses:")
	if len(peerAddresses) > 0 {
		for _, addr := range peerAddresses {
			fmt.Println("  *", addr)
		}
	} else {
		fmt.Println("  * No addresses available")
	}

	// construct api server
	apiServer := api.New(
		config.Daemon.PeerKey,
		network,
		discoverer,
		exchange,
		localInfo,
		graphStore,
		dag,
		version.Version,
		version.Commit,
		version.Date,
		config.API.Token,
	)

	fmt.Println("* HTTP API address:")
	fmt.Printf("  * http://localhost:%d\n", config.API.Port)
	apiServer.Serve(fmt.Sprintf("0.0.0.0:%d", config.API.Port))
}
