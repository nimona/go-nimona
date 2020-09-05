package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"

	"nimona.io/internal/rand"
	"nimona.io/internal/version"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"
)

type config struct {
	PeerPrivateKey  crypto.PrivateKey  `envconfig:"PEER_PRIVATE_KEY"`
	BindAddress     string             `envconfig:"BIND_ADDRESS"`
	AnnounceAddress string             `envconfig:"ANNOUNCE_ADDRESS"`
	BootstrapPeers  []peer.Shorthand   `envconfig:"BOOTSTRAP_PEERS"` // shorthands
	PingPeers       []crypto.PublicKey `envconfig:"PING_PEERS"`
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
	res := resolver.New(
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

	// construct object store
	db, err := sql.Open("sqlite3", "sqlite3.db")
	if err != nil {
		logger.Fatal("error opening sql file", log.Error(err))
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		logger.Fatal("error starting sql store", log.Error(err))
	}

	// construct manager
	man := objectmanager.New(
		ctx,
		net,
		res,
		str,
	)

	allPingsReceived := make(chan struct{}, 1)
	allPingsSent := make(chan struct{}, 1)

	// listen for pings
	go func() {
		pingedFromPeers := map[crypto.PublicKey]bool{} // [key]pinged
		for _, p := range cfg.PingPeers {
			pingedFromPeers[p] = false
		}
		sub := man.Subscribe(
			objectmanager.FilterByObjectType("ping"),
		)
		defer sub.Cancel()
		for {
			env, err := sub.Next()
			if err != nil {
				if err != object.ErrReaderDone {
					logger.Error("error while reading pings", log.Error(err))
				}
				return
			}
			fmt.Printf(
				"%s received ping from %s\n",
				local.GetPrimaryPeerKey().PublicKey(),
				env.GetOwner(),
			)
			pingedFromPeers[env.GetOwner()] = true
			// check if all have pinged us
			allPinged := true
			for _, pinged := range pingedFromPeers {
				if !pinged {
					allPinged = false
					break
				}
			}
			if allPinged {
				close(allPingsReceived)
				return
			}
		}
	}()

	ping := func(peerKey crypto.PublicKey) error {
		sctx := context.New(
			context.WithParent(ctx),
			context.WithTimeout(time.Second*5),
		)
		recipients, err := res.Lookup(
			sctx,
			resolver.LookupByOwner(peerKey),
		)
		if err != nil {
			return err
		}
		for recipient := range recipients {
			if err := net.Send(
				sctx,
				new(object.Object).
					SetType("ping").
					SetOwner(local.GetPrimaryPeerKey().PublicKey()).
					Set("nonce:s", rand.String(8)),
				recipient,
			); err != nil {
				logger.Error(
					"error sending ping to peer",
					log.String("publicKey", recipient.PublicKey().String()),
					log.Strings("addresses", recipient.Addresses),
					log.Error(err),
				)
				continue
			}
			fmt.Printf(
				"%s sent ping to %s\n",
				recipient.PublicKey().String(),
				local.GetPrimaryPeerKey().PublicKey(),
			)
			return nil
		}
		return errors.New("unable to ping")
	}

	go func() {
		pingPeers := map[crypto.PublicKey]bool{} // [key]pinged
		for _, p := range cfg.PingPeers {
			pingPeers[p] = false
		}
		for {
			time.Sleep(time.Second)
			leftToPing := 0
			for peerKey, pinged := range pingPeers {
				if pinged {
					continue
				}
				leftToPing++
				if err := ping(peerKey); err != nil {
					logger.Error(
						"error trying to ping peer",
						log.String("publicKey", peerKey.String()),
						log.Error(err),
					)
					continue
				}
				pingPeers[peerKey] = true
			}
			if leftToPing == 0 {
				close(allPingsSent)
				return
			}
		}
	}()

	// and wait both channels to close
	<-allPingsSent
	fmt.Println("all pings sent")
	<-allPingsReceived
	fmt.Println("all pings received")

	// finally terminate everything
	lis.Close() // nolint: errcheck
}
