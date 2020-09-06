package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"

	"nimona.io/internal/version"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
)

type config struct {
	PeerPrivateKey  crypto.PrivateKey `envconfig:"PEER_PRIVATE_KEY"`
	BindAddress     string            `envconfig:"BIND_ADDRESS"`
	AnnounceAddress string            `envconfig:"ANNOUNCE_ADDRESS"`
	BootstrapPeers  []peer.Shorthand  `envconfig:"BOOTSTRAP_PEERS"` // shorthands
}

func main() {
	ctx := context.New(
		context.WithCorrelationID("nimona"),
	)

	logger := log.FromContext(ctx).With(
		log.String("build.version", version.Version),
		log.String("build.commit", version.Commit),
		log.String("build.timestamp", version.Date),
	)

	cfg := &config{
		BindAddress: "0.0.0.0:0",
	}
	if err := envconfig.Process("nimona", cfg); err != nil {
		logger.Fatal("error processing config", log.Error(err))
	}

	if cfg.PeerPrivateKey.IsEmpty() {
		logger.Fatal("missing peer private key")
	}

	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.PutPrimaryPeerKey(cfg.PeerPrivateKey)

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	// start listening
	lis, err := net.Listen(
		ctx,
		cfg.BindAddress,
	)
	if err != nil {
		logger.Fatal("error while listening", log.Error(err))
	}

	// add announce address
	if cfg.AnnounceAddress != "" {
		local.PutAddresses("tcps:" + cfg.AnnounceAddress)
	}

	// convert shorthands into peers
	bootstrapPeers := []*peer.Peer{}
	for _, s := range cfg.BootstrapPeers {
		bootstrapPeer, err := s.Peer()
		if err != nil {
			logger.Fatal("error parsing bootstrap peer", log.Error(err))
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// construct new resolver
	resolver.New(
		ctx,
		net,
		resolver.WithBoostrapPeers(bootstrapPeers),
	)

	logger = logger.With(
		log.String("peer.privateKey", local.GetPrimaryPeerKey().String()),
		log.String("peer.publicKey", local.GetPrimaryPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", local.GetAddresses()),
	)

	logger.Info("ready")

	// register for termination signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// and wait for one
	<-sigs

	// finally terminate everything
	logger.Info("shutting down")
	lis.Close() // nolint: errcheck
}
