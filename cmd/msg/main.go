package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/daemon/config"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/log"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/version"
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

	sub := d.Exchange.Subscribe(exchange.FilterByObjectType("nimona.io/msg"))
	go func() {
		for {
			e, err := sub.Next()
			if err != nil {
				panic(err)
			}
			fmt.Println(">", e.Sender.String(), e.Payload.Get("body:s").(string))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		switch text {
		case "/info":
			fmt.Println("addresses", d.LocalPeer.GetAddresses())
			fmt.Println("peer", config.Peer.PeerKey.PublicKey().String())
			fmt.Println("identity", config.Peer.IdentityKey.PublicKey().String())
			continue
		}
		ps := strings.Split(text, " ")
		if len(ps) < 2 {
			fmt.Println("missing recipient")
			continue
		}
		recipient := ps[0]
		body := strings.Join(ps[1:], " ")
		msg := Msg{
			Datetime: time.Now().Unix(),
			Body:     body,
		}
		err := d.Exchange.Send(
			context.New(),
			msg.ToObject(),
			peer.LookupByOwner(crypto.PublicKey(recipient)),
			exchange.WithAsync(),
		)
		go func() {
			if err != nil {
				fmt.Println("* could not send message;", err)
			}
		}()
	}
}
