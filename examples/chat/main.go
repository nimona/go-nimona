package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"nimona.io/internal/net"
	"nimona.io/pkg/config"
	"nimona.io/pkg/configstore"
	"nimona.io/pkg/context"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/keystream"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
	"nimona.io/pkg/tilde"
	"nimona.io/pkg/version"
)

type Config struct {
	Nonce string `envconfig:"NONCE" json:"nonce"`
}

type chat struct {
	network       network.Network
	objectmanager objectmanager.ObjectManager
	objectstore   objectstore.Store
	resolver      resolver.Resolver
	logger        log.Logger
}

func (c *chat) subscribe(
	conversationRootHash tilde.Digest,
) (chan interface{}, error) {
	objects := make(chan *object.Object)
	events := make(chan interface{})

	// handle objects from subscriptions or store
	go func() {
		for o := range objects {
			switch o.Type {
			case ConversationMessageAddedType:
				v := &ConversationMessageAdded{}
				object.Unmarshal(o, v)
				if v.Body == "" || v.Metadata.Timestamp == "" {
					fmt.Println("> Received message without date or body")
					continue
				}
				if v.Metadata.Owner.IsEmpty() {
					fmt.Println("> Received unsigned message")
					continue
				}
				events <- v
			case ConversationNicknameUpdatedType:
				v := &ConversationNicknameUpdated{}
				object.Unmarshal(o, v)
				if v.Nickname == "" {
					fmt.Println("> Received nickname update without nickname")
					continue
				}
				if v.Metadata.Owner.IsEmpty() {
					fmt.Println("> Received unsigned nickname update")
					continue
				}
				events <- v
			}
		}
	}()

	// get objects from db first
	or, err := c.objectstore.GetByStream(conversationRootHash)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			o, err := or.Read()
			if err != nil {
				break
			}
			objects <- o
		}
		// subscribe to conversation updates
		sub := c.objectmanager.Subscribe(
			objectmanager.FilterByStreamHash(conversationRootHash),
		)
		for {
			o, err := sub.Read()
			if err != nil {
				break
			}
			objects <- o
		}
	}()

	// create subscription for stream
	go func() {
		// add a subscription to the stream if one doesn't already exist
		or, err := c.objectstore.GetByStream(conversationRootHash)
		if err != nil {
			c.logger.Fatal("error checking for subscription", log.Error(err))
		}
		alreadySubscribed := false
		for {
			o, err := or.Read()
			if err != nil {
				break
			}
			if o.Type == stream.SubscriptionType {
				s := &stream.Subscription{}
				if err := object.Unmarshal(o, s); err != nil {
					continue
				}
				if s.Metadata.Owner.Equals(
					c.network.GetPeerKey().PublicKey().DID(),
				) {
					alreadySubscribed = true
					or.Close()
					break
				}
			}
		}
		if !alreadySubscribed {
			ctx := context.New(context.WithTimeout(time.Second * 5))
			so := object.MustMarshal(&stream.Subscription{
				Metadata: object.Metadata{
					Owner: c.network.GetPeerKey().PublicKey().DID(),
					Root:  conversationRootHash,
				},
				RootHashes: []tilde.Digest{
					conversationRootHash,
				},
			})
			if _, err := c.objectmanager.Append(ctx, so); err != nil {
				c.logger.Fatal("could not persist conversation sub", log.Error(err))
			}
		}

		// sync conversation
		queryCtx := context.New(context.WithTimeout(time.Second * 5))
		peers, err := c.resolver.LookupByContent(
			queryCtx,
			conversationRootHash,
		)
		if err != nil {
			c.logger.Error("could not find any peers that have this hash",
				log.String("hash", string(conversationRootHash)),
			)
			return
		}
		for _, p := range peers {
			reqCtx := context.New(context.WithTimeout(time.Second * 5))
			cr, err := c.objectmanager.RequestStream(
				reqCtx,
				conversationRootHash,
				p.PublicKey.DID(),
			)
			if err != nil {
				c.logger.Warn(
					"could not ask peer for stream",
					log.String("peer", p.PublicKey.String()),
				)
				continue
			}
			for {
				o, err := cr.Read()
				if err != nil {
					break
				}
				c.objectmanager.Put(
					context.New(),
					o,
				)
			}
			cr.Close()
		}
	}()
	return events, nil
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

	cConfig := &Config{
		Nonce: "hello-world!!1",
	}
	nConfig, err := config.New(
		config.WithExtraConfig("CHAT", cConfig),
	)
	if err != nil {
		logger.Fatal("error loading config", log.Error(err))
	}

	log.DefaultLogger.SetLogLevel(nConfig.LogLevel)

	// construct new network
	nnet := net.New(nConfig.Peer.PrivateKey)
	net := network.New(
		ctx,
		nnet,
		nConfig.Peer.PrivateKey,
	)

	if nConfig.Peer.BindAddress != "" {
		// start listening
		lis, err := net.Listen(
			ctx,
			nConfig.Peer.BindAddress,
			network.ListenOnLocalIPs,
			// network.ListenOnExternalPort,
		)
		if err != nil {
			logger.Fatal("error while listening", log.Error(err))
		}
		defer lis.Close() // nolint: errcheck
	}

	// convert shorthands into connection infos
	bootstrapPeers := []*peer.ConnectionInfo{}
	for _, s := range nConfig.Peer.Bootstraps {
		bootstrapPeer, err := s.GetConnectionInfo()
		if err != nil {
			logger.Fatal("error parsing bootstrap peer", log.Error(err))
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// add bootstrap peers as relays
	net.RegisterRelays(bootstrapPeers...)

	// construct object store
	db, err := sql.Open("sqlite3", filepath.Join(nConfig.Path, "chat.sqlite"))
	if err != nil {
		logger.Fatal("error opening sql file", log.Error(err))
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		logger.Fatal("error starting sql store", log.Error(err))
	}

	// construct configstore db
	pdb, err := sql.Open("sqlite", filepath.Join(nConfig.Path, "config.sqlite"))
	if err != nil {
		logger.Fatal("opening sql file for configstore", log.Error(err))
	}

	// construct configstore
	prf, err := configstore.NewSQLProvider(pdb)
	if err != nil {
		logger.Fatal("constructing configstore provider", log.Error(err))
	}

	// construct hypothetical root in order to get a root hash
	conversationRoot := ConversationStreamRoot{
		Nonce: cConfig.Nonce,
	}

	conversationRootObject := object.MustMarshal(conversationRoot)
	conversationRootHash := conversationRootObject.Hash()

	// construct key stream manager
	ksm, err := keystream.NewKeyManager(
		net,
		str,
		prf,
	)
	if err != nil {
		logger.Fatal("constructing keystream manager", log.Error(err))
	}

	// construct new resolver
	res := resolver.New(
		ctx,
		nnet,
		nConfig.Peer.PrivateKey,
		str,
		ksm,
		resolver.WithBoostrapPeers(bootstrapPeers...),
	)

	// construct manager
	man := objectmanager.New(
		ctx,
		net,
		res,
		str,
	)

	// register conversation in object manager
	if err := man.Put(ctx, conversationRootObject); err != nil {
		logger.Fatal("could not persist conversation root", log.Error(err))
	}

	logger = logger.With(
		log.String("peer.publicKey", net.GetPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", net.GetAddresses()),
	)

	// ready
	logger.Info("ready")

	c := &chat{
		network:       net,
		objectmanager: man,
		objectstore:   str,
		resolver:      res,
		logger:        logger,
	}

	events, err := c.subscribe(conversationRootHash)
	if err != nil {
		logger.Fatal("error subscribing to conversation", log.Error(err))
	}

	app := NewApp(conversationRootHash.String())
	app.Chat = c
	go app.Show()

	go func() {
		for {
			select {
			case nickname := <-app.Channels.SelfNicknameUpdated:
				if _, err := man.Append(
					context.New(
						context.WithTimeout(time.Second*5),
					),
					object.MustMarshal(&ConversationNicknameUpdated{
						Metadata: object.Metadata{
							Owner:     net.GetPeerKey().PublicKey().DID(),
							Root:      conversationRootHash,
							Timestamp: time.Now().Format(time.RFC3339),
						},
						Nickname: nickname,
					}),
				); err != nil {
					logger.Warn(
						"error putting message",
						log.Error(err),
					)
				}
			case input := <-app.Channels.InputLines:
				if _, err := man.Append(
					context.New(
						context.WithTimeout(time.Second*5),
					),
					object.MustMarshal(&ConversationMessageAdded{
						Metadata: object.Metadata{
							Owner:     net.GetPeerKey().PublicKey().DID(),
							Root:      conversationRootHash,
							Timestamp: time.Now().Format(time.RFC3339),
						},
						Body: input,
					}),
				); err != nil {
					logger.Warn(
						"error putting message",
						log.Error(err),
					)
				}
			}
		}
	}()

	for event := range events {
		switch v := event.(type) {
		case *ConversationMessageAdded:
			t, err := time.Parse(time.RFC3339, v.Metadata.Timestamp)
			if err != nil {
				continue
			}
			app.Channels.MessageAdded <- &Message{
				Hash:             object.MustMarshal(v).Hash().String(),
				ConversationHash: v.Metadata.Root.String(),
				SenderKey:        v.Metadata.Owner.String(),
				Body:             strings.TrimSpace(v.Body),
				Created:          t.UTC(),
			}
		case *ConversationNicknameUpdated:
			updated, _ := time.Parse(time.RFC3339, v.Metadata.Timestamp)
			app.Channels.ParticipantUpdated <- &Participant{
				ConversationHash: v.Metadata.Root.String(),
				Key:              v.Metadata.Owner.String(),
				Nickname:         v.Nickname,
				Updated:          updated.UTC(),
			}
		}
	}
}

func first(t string, i int) string {
	if len(t) <= i {
		return t
	}
	return t[:i]
}
