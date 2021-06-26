package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
	"nimona.io/pkg/version"
)

type Config struct {
	Nonce string `envconfig:"NONCE" json:"nonce"`
}

type chat struct {
	network       network.Network
	local         localpeer.LocalPeer
	objectmanager objectmanager.ObjectManager
	objectstore   objectstore.Store
	resolver      resolver.Resolver
	logger        log.Logger
}

func (c *chat) subscribe(
	conversationRootHash chore.Hash,
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
				if v.Body == "" || v.Metadata.Datetime == "" {
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
					c.local.GetPeerKey().PublicKey(),
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
					Owner:  c.local.GetPeerKey().PublicKey(),
					Stream: conversationRootHash,
				},
				RootHashes: []chore.Hash{
					conversationRootHash,
				},
			})
			if _, err := c.objectmanager.Put(ctx, so); err != nil {
				c.logger.Fatal("could not persist conversation sub", log.Error(err))
			}
		}

		// sync conversation
		queryCtx := context.New(context.WithTimeout(time.Second * 5))
		peers, err := c.resolver.Lookup(
			queryCtx,
			resolver.LookupByHash(conversationRootHash),
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
				p,
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

	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.SetPeerKey(nConfig.Peer.PrivateKey)

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
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
	db, err := sql.Open("sqlite3", filepath.Join(nConfig.Path, "chat.db"))
	if err != nil {
		logger.Fatal("error opening sql file", log.Error(err))
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		logger.Fatal("error starting sql store", log.Error(err))
	}

	// construct hypothetical root in order to get a root hash
	conversationRoot := ConversationStreamRoot{
		Nonce: cConfig.Nonce,
	}

	conversationRootObject := object.MustMarshal(conversationRoot)
	conversationRootHash := conversationRootObject.Hash()

	// construct new resolver
	res := resolver.New(
		ctx,
		net,
		str,
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
	if _, err := man.Put(ctx, conversationRootObject); err != nil {
		logger.Fatal("could not persist conversation root", log.Error(err))
	}

	logger = logger.With(
		log.String("peer.publicKey", local.GetPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", net.GetAddresses()),
	)

	// ready
	logger.Info("ready")

	c := &chat{
		network:       net,
		local:         local,
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
				if _, err := man.Put(
					context.New(
						context.WithTimeout(time.Second*5),
					),
					object.MustMarshal(&ConversationNicknameUpdated{
						Metadata: object.Metadata{
							Owner:    local.GetPeerKey().PublicKey(),
							Stream:   conversationRootHash,
							Datetime: time.Now().Format(time.RFC3339),
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
				if _, err := man.Put(
					context.New(
						context.WithTimeout(time.Second*5),
					),
					object.MustMarshal(&ConversationMessageAdded{
						Metadata: object.Metadata{
							Owner:    local.GetPeerKey().PublicKey(),
							Stream:   conversationRootHash,
							Datetime: time.Now().Format(time.RFC3339),
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
			t, err := time.Parse(time.RFC3339, v.Metadata.Datetime)
			if err != nil {
				continue
			}
			app.Channels.MessageAdded <- &Message{
				Hash:             object.MustMarshal(v).Hash().String(),
				ConversationHash: v.Metadata.Stream.String(),
				SenderKey:        v.Metadata.Owner.String(),
				Body:             strings.TrimSpace(v.Body),
				Created:          t.UTC(),
			}
		case *ConversationNicknameUpdated:
			updated, _ := time.Parse(time.RFC3339, v.Metadata.Datetime)
			app.Channels.ParticipantUpdated <- &Participant{
				ConversationHash: v.Metadata.Stream.String(),
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
