package main

import (
	"fmt"
	"os"

	"nimona.io/internal/api"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/internal/daemon"
	"nimona.io/internal/daemon/config"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/internal/version"
)

func main() {
	ctx := context.New(
		context.WithCorrelationID("nimona"),
	)

	nodeAlias := os.Getenv("NIMONA_ALIAS")
	if nodeAlias != "" {
		log.DefaultLogger = log.DefaultLogger.With(
			log.String("$alias", nodeAlias),
		)
	}

	logger := log.FromContext(ctx).With(
		log.String("build.version", version.Version),
		log.String("build.commit", version.Commit),
		log.String("build.timestamp", version.Date),
	)

	// load config
	logger.Info("loading config file")
	config := config.New()
	if err := config.Load(); err != nil {
		logger.Fatal("could not load config file", log.Error(err))
	}

	// create peer key pair if it does not exist
	if config.Peer.PeerKey == "" {
		logger.Info("creating new peer key pair")
		peerKey, err := crypto.GenerateEd25519PrivateKey()
		if err != nil {
			logger.Fatal("could not generate peer key", log.Error(err))
		}
		config.Peer.PeerKey = peerKey
	}

	// create identity key pair if it does not exist
	// TODO this is temporary
	if config.Peer.IdentityKey == "" {
		logger.Info("creating new identity key pair")
		identityKey, err := crypto.GenerateEd25519PrivateKey()
		if err != nil {
			logger.Fatal("could not generate identity key", log.Error(err))
		}
		config.Peer.IdentityKey = identityKey
	}

	// update config
	if err := config.Update(); err != nil {
		logger.Fatal("could not update config", log.Error(err))
	}

	logger.Info("loaded config", log.Any("config", config))

	d, err := daemon.New(ctx, config)
	if err != nil {
		logger.Fatal("could not construct daemon", log.Error(err))
	}

	// print some info
	nlogger := logger.With(
		log.Strings("addresses", d.Net.Addresses()),
		log.String("peer", config.Peer.PeerKey.PublicKey().String()),
	)

	ik := config.Peer.IdentityKey
	if ik != "" {
		nlogger = nlogger.With(
			log.String("identity", ik.PublicKey().String()),
		)
	}

	nlogger.Info("starting HTTP API")

	// construct api server
	apiServer := api.New(
		config,
		config.Peer.PeerKey,
		keychain.DefaultKeychain,
		d.Net,
		d.Resolver,
		d.Exchange,
		d.Store,
		d.ObjectManager,
		version.Version,
		version.Commit,
		version.Date,
		config.API.Token,
	)

	apiAddress := fmt.Sprintf("%s:%d", config.API.Host, config.API.Port)
	logger.Info(
		"starting http server",
		log.String("address", apiAddress),
	)
	apiServer.Serve(apiAddress) // nolint: errcheck
}
