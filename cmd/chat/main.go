package main

import (
	"database/sql"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"

	"nimona.io/internal/version"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

func init() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
}

// nolint: lll
type config struct {
	Peer struct {
		PrivateKey  crypto.PrivateKey `envconfig:"PRIVATE_KEY"`
		BindAddress string            `envconfig:"BIND_ADDRESS" default:"0.0.0.0:0"`
		Bootstraps  []peer.Shorthand  `envconfig:"BOOTSTRAPS"`
	} `envconfig:"PEER"`
	Chat struct {
		Nonce string `envconfig:"NONCE"`
	} `envconfig:"CHAT"`
}

type chat struct {
	local         localpeer.LocalPeer
	objectmanager objectmanager.ObjectManager
	objectstore   objectstore.Store
	resolver      resolver.Resolver
	logger        log.Logger
}

func (c *chat) subscribe(
	conversationRootHash object.Hash,
) (chan interface{}, error) {
	objects := make(chan object.Object)
	events := make(chan interface{})

	// handle objects from subscriptions or store
	go func() {
		typeConversationMessageAdded := new(ConversationMessageAdded).GetType()
		for o := range objects {
			switch o.GetType() {
			case typeConversationMessageAdded:
				v := &ConversationMessageAdded{}
				v.FromObject(o)
				if v.Body == "" || v.Datetime == "" {
					fmt.Println("> Received message without date of body")
					continue
				}
				if v.Metadata.Owner.IsEmpty() {
					fmt.Println("> Received unsigned message")
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
			objects <- *o
		}
		// subscribe to conversation updates
		sub := c.objectmanager.Subscribe(
			objectmanager.FilterByStreamHash(conversationRootHash),
		)
		for {
			o, err := sub.Next()
			if err != nil {
				break
			}
			objects <- o
		}
	}()

	// create subscription for stream
	go func() {
		// add a subscription to the stream
		// TODO check if we are already subscribed
		ctx := context.New(context.WithTimeout(time.Second * 5))
		if _, err := c.objectmanager.Put(ctx, stream.Subscription{
			Metadata: object.Metadata{
				Owner:  c.local.GetPrimaryPeerKey().PublicKey(),
				Stream: conversationRootHash,
			},
			RootHashes: []object.Hash{
				conversationRootHash,
			},
		}.ToObject()); err != nil {
			c.logger.Fatal("could not persist conversation sub", log.Error(err))
		}

		// sync conversation
		queryCtx := context.New(context.WithTimeout(time.Second * 5))
		peers, err := c.resolver.Lookup(
			queryCtx,
			resolver.LookupByContentHash(conversationRootHash),
		)
		if err != nil {
			c.logger.Error("could not find any peers that have this hash")
			return
		}
		for p := range peers {
			reqCtx := context.New(context.WithTimeout(time.Second * 5))
			cr, err := c.objectmanager.RequestStream(
				reqCtx,
				conversationRootHash,
				p,
			)
			if err != nil {
				c.logger.Warn(
					"could not ask peer for stream",
					log.String("peer", p.PublicKey().String()),
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
					*o,
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

	cfg := &config{}
	if err := envconfig.Process("nimona", cfg); err != nil {
		logger.Fatal("error processing config", log.Error(err))
	}

	if cfg.Peer.PrivateKey.IsEmpty() {
		k, err := crypto.GenerateEd25519PrivateKey()
		if err != nil {
			logger.Fatal("missing peer key and unable to generate one")
		}
		cfg.Peer.PrivateKey = k
	}

	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.PutPrimaryPeerKey(cfg.Peer.PrivateKey)

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	if cfg.Peer.BindAddress != "" {
		// start listening
		lis, err := net.Listen(
			ctx,
			cfg.Peer.BindAddress,
			network.ListenOnLocalIPs,
			network.ListenOnPrivateIPs,
			network.ListenOnExternalPort,
		)
		if err != nil {
			logger.Fatal("error while listening", log.Error(err))
		}
		defer lis.Close() // nolint: errcheck
	}

	// make sure we have some bootstrap peers to start with
	if len(cfg.Peer.Bootstraps) == 0 {
		cfg.Peer.Bootstraps = []peer.Shorthand{
			"ed25519.CJi6yjjXuNBFDoYYPrp697d6RmpXeW8ZUZPmEce9AgEc@tcps:asimov.node.nimona.io:22581",
			"ed25519.6fVWVAK2DVGxBhtVBvzNWNKBWk9S83aQrAqGJfrxr75o@tcps:egan.node.nimona.io:22581",
			"ed25519.7q7YpmPNQmvSCEBWW8ENw8XV8MHzETLostJTYKeaRTcL@tcps:sloan.node.nimona.io:22581",
		}
	}

	// convert shorthands into peers
	bootstrapPeers := []*peer.Peer{}
	for _, s := range cfg.Peer.Bootstraps {
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
	db, err := sql.Open("sqlite3", "chat.db")
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

	// if no noce is specified use a default
	if cfg.Chat.Nonce == "" {
		cfg.Chat.Nonce = "hello-world!!1"
	}

	// construct hypothetical root in order to get a root hash
	conversationRoot := ConversationStreamRoot{
		Nonce: cfg.Chat.Nonce,
	}

	// register types so object manager persists them
	local.PutContentTypes(
		new(ConversationStreamRoot).GetType(),
		new(ConversationMessageAdded).GetType(),
		new(stream.Subscription).GetType(),
	)

	conversationRootObject := conversationRoot.ToObject()
	conversationRootHash := conversationRootObject.Hash()

	// register conversation in object manager
	if _, err := man.Put(ctx, conversationRootObject); err != nil {
		logger.Fatal("could not persist conversation root", log.Error(err))
	}

	c := &chat{
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
		for input := range app.Channels.InputLines {
			if _, err := man.Put(
				context.New(
					context.WithTimeout(time.Second*5),
				),
				ConversationMessageAdded{
					Metadata: object.Metadata{
						Owner:  local.GetPrimaryPeerKey().PublicKey(),
						Stream: conversationRootHash,
					},
					Body:     input,
					Datetime: time.Now().Format(time.RFC3339Nano),
				}.ToObject(),
			); err != nil {
				logger.Warn(
					"error putting message",
					log.Error(err),
				)
			}
		}
	}()

	for event := range events {
		switch v := event.(type) {
		case *ConversationMessageAdded:
			t, err := time.Parse(time.RFC3339Nano, v.Datetime)
			if err != nil {
				continue
			}
			usr := last(v.Metadata.Owner.String(), 8)
			app.Channels.MessageAdded <- &Message{
				Hash:             v.ToObject().Hash().String(),
				ConversationHash: v.Metadata.Stream.String(),
				SenderHash:       v.Metadata.Owner.String(),
				SenderNickname:   usr,
				Body:             strings.TrimSpace(v.Body),
				Created:          t,
			}
		}
	}
}

func last(t string, i int) string {
	if len(t) <= i {
		return t
	}
	return t[len(t)-i:]
}
