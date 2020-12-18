package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"nimona.io/internal/version"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

var (
	objectPinRequestType   = new(PinRequest).Type()
	objectPinResponseType  = new(PinResponse).Type()
	objectListRequestType  = new(ListRequest).Type()
	objectListResponseType = new(ListResponse).Type()
)

func main() {
	ctx := context.New(
		context.WithCorrelationID("nimona"),
	)
	logger := log.FromContext(ctx).With(
		log.String("build.version", version.Version),
		log.String("build.commit", version.Commit),
		log.String("build.timestamp", version.Date),
	)

	pinConfig := &Config{}
	nimConfig, err := config.New(
		config.WithExtraConfig("PINNING", pinConfig),
	)
	if err != nil {
		logger.Fatal("error parsing config", log.Error(err))
	}

	// construct local peer
	local := localpeer.New()

	// attach peer private key from config
	local.PutPrimaryPeerKey(nimConfig.Peer.PrivateKey)
	local.PutContentTypes(
		objectPinRequestType,
		objectPinResponseType,
		objectListRequestType,
		objectListResponseType,
	)

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	// start listening
	lis, err := net.Listen(
		ctx,
		nimConfig.Peer.BindAddress,
		network.ListenOnLocalIPs,
		network.ListenOnPrivateIPs,
	)
	if err != nil {
		logger.Fatal("error while listening", log.Error(err))
	}

	// convert shorthands into peers
	bootstrapPeers := []*peer.ConnectionInfo{}
	for _, s := range nimConfig.Peer.Bootstraps {
		bootstrapPeer, err := s.ConnectionInfo()
		if err != nil {
			logger.Fatal("error parsing bootstrap peer", log.Error(err))
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// construct new resolver
	res := resolver.New(
		ctx,
		net,
		resolver.WithBoostrapPeers(bootstrapPeers...),
	)

	logger = logger.With(
		log.String("peer.publicKey", local.GetPrimaryPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", local.GetAddresses()),
	)

	// construct object store
	db, err := sql.Open("sqlite3", "file_transfer.db")
	if err != nil {
		logger.Fatal("error opening sql file", log.Error(err))
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		logger.Fatal("error starting sql store", log.Error(err))
	}

	// construct object manager
	man := objectmanager.New(
		ctx,
		net,
		res,
		str,
	)

	srv := &Service{
		logger:        logger,
		local:         local,
		objectmanager: man,
		objectstore:   str,
		network:       net,
		resolver:      res,
		listener:      lis,
		config:        pinConfig,
		nimonaConfig:  nimConfig,
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "pin":
		if len(args) != 2 {
			logger.Fatal("invalid number of arguments")
		}
		err := srv.Pin(
			context.New(
				context.WithTimeout(5*time.Second),
			),
			crypto.PublicKey(args[0]),
			object.Hash(args[1]),
		)
		if err != nil {
			fmt.Println("error pinning object, err:", err)
			return
		}
		fmt.Println("successfully pinned object")
	case "list":
		if len(args) != 1 {
			logger.Fatal("invalid number of arguments")
		}
		hashes, err := srv.List(
			context.New(
				context.WithTimeout(5*time.Second),
			),
			crypto.PublicKey(args[0]),
		)
		if err != nil {
			logger.Fatal(
				"unable to list pinned objects",
				log.String("publicKey", args[0]),
				log.Error(err),
			)
		}
		fmt.Println("Pinned objects:")
		for _, hash := range hashes {
			fmt.Println("-", hash)
		}
	case "serve":
		if len(args) != 0 {
			logger.Fatal("invalid number of arguments")
		}
		fmt.Println("public key:", srv.local.ConnectionInfo().PublicKey)
		fmt.Println("pinned hashes:")
		for _, hash := range srv.local.GetContentHashes() {
			fmt.Println("-", hash)
		}
		srv.Serve()
	default:
		fmt.Println("unknown command")
		return
	}

}
