package main

import (
	"errors"
	"strings"
	"time"

	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/did"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
	"nimona.io/pkg/version"
)

type (
	Provider struct {
		network       network.Network
		resolver      resolver.Resolver
		objectstore   *sqlobjectstore.Store
		objectmanager objectmanager.ObjectManager
		logger        log.Logger
	}
	Config struct{}
)

type InitRequest struct {
	ConfigPath   string   `json:"configPath"`
	ContentTypes []string `json:"contentTypes"`
}

func New(initRequest *InitRequest) *Provider {
	ctx := context.New(
		context.WithCorrelationID("nimona"),
	)

	logger := log.FromContext(ctx).With(
		log.String("build.version", version.Version),
		log.String("build.commit", version.Commit),
		log.String("build.timestamp", version.Date),
	)

	cConfig := &Config{}
	d, err := daemon.New(
		ctx,
		daemon.WithConfigOptions(
			config.WithoutPersistence(),
			config.WithDefaultPath(initRequest.ConfigPath),
			config.WithExtraConfig("CHAT", cConfig),
			config.WithDefaultListenOnLocalIPs(),
			config.WithDefaultListenOnPrivateIPs(),
			config.WithDefaultListenOnExternalPort(),
		),
	)
	if err != nil {
		logger.Fatal("error loading config", log.Error(err))
	}

	nConfig := d.Config()
	net := d.Network()
	str := d.ObjectStore().(*sqlobjectstore.Store)
	res := d.Resolver()
	man := d.ObjectManager()

	log.DefaultLogger.SetLogLevel(nConfig.LogLevel)

	logger = logger.With(
		log.String("peer.publicKey", net.GetPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", net.GetAddresses()),
	)

	logger.Error(
		"ready",
		log.Any("addresses", net.GetAddresses()),
	)

	return &Provider{
		network:       net,
		resolver:      res,
		objectstore:   str,
		objectmanager: man,
		logger:        logger,
	}
}

func (p *Provider) GetConnectionInfo() *peer.ConnectionInfo {
	return p.network.GetConnectionInfo()
}

type GetRequest struct {
	Lookups  []string `json:"lookups"`
	OrderBy  string   `json:"orderBy"`
	OrderDir string   `json:"orderDir"`
	Limit    int      `json:"limit"`
	Offset   int      `json:"offset"`
}

type GetResponse struct {
	ObjectBodies []string `json:"objectBodies"`
}

func (p *Provider) Get(
	ctx context.Context,
	req GetRequest,
) (object.ReadCloser, error) {
	opts := []sqlobjectstore.FilterOption{}
	filterByType := []string{}
	filterByHash := []tilde.Hash{}
	filterByOwner := []did.DID{}
	filterByStreamHash := []tilde.Hash{}
	for _, lookup := range req.Lookups {
		parts := strings.Split(lookup, ":")
		if len(parts) < 2 {
			return nil, errors.New("invalid lookup query")
		}
		prefix := parts[0]
		v := strings.Join(parts[1:], ":")
		switch prefix {
		case "type":
			filterByType = append(
				filterByType,
				v,
			)
		case "hash":
			filterByHash = append(
				filterByHash,
				tilde.Hash(v),
			)
		case "owner":
			k := did.DID{}
			k.UnmarshalString(v) // nolint: errcheck
			filterByOwner = append(
				filterByOwner,
				k,
			)
		case "stream":
			filterByStreamHash = append(
				filterByStreamHash,
				tilde.Hash(v),
			)
		}
		if req.OrderBy != "" {
			opts = append(
				opts,
				sqlobjectstore.FilterOrderBy(req.OrderBy),
			)
		}
		if req.OrderDir != "" {
			opts = append(
				opts,
				sqlobjectstore.FilterOrderDir(req.OrderDir),
			)
		}
		if req.Limit > 0 && req.Offset > 0 {
			opts = append(
				opts,
				sqlobjectstore.FilterLimit(req.Limit, req.Offset),
			)
		}
	}
	if len(filterByType) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByObjectType(filterByType...),
		)
	}
	if len(filterByHash) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByHash(filterByHash...),
		)
	}
	if len(filterByOwner) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByOwner(filterByOwner...),
		)
	}
	if len(filterByStreamHash) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByStreamHash(filterByStreamHash...),
		)
	}
	return p.objectstore.Filter(opts...)
}

type SubscribeRequest struct {
	Lookups []string `json:"lookups"`
}

// payload should start with one of the following:
// - type:<type>
// - hash:<hash>
// - stream:<rootHash>
// - owner:<publicKey>
func (p *Provider) Subscribe(
	ctx context.Context,
	req SubscribeRequest,
) (object.ReadCloser, error) {
	opts := []objectmanager.LookupOption{}
	filterByType := []string{}
	filterByHash := []tilde.Hash{}
	filterByOwner := []did.DID{}
	filterByStreamHash := []tilde.Hash{}
	for _, lookup := range req.Lookups {
		parts := strings.Split(lookup, ":")
		if len(parts) < 2 {
			return nil, errors.New("invalid lookup query")
		}
		prefix := parts[0]
		v := strings.Join(parts[1:], ":")
		switch prefix {
		case "type":
			filterByType = append(
				filterByType,
				v,
			)
		case "hash":
			filterByHash = append(
				filterByHash,
				tilde.Hash(v),
			)
		case "owner":
			k := did.DID{}
			k.UnmarshalString(v) // nolint: errcheck
			filterByOwner = append(
				filterByOwner,
				k,
			)
		case "stream":
			filterByStreamHash = append(
				filterByStreamHash,
				tilde.Hash(v),
			)
		}
	}
	if len(filterByType) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByObjectType(filterByType...),
		)
	}
	if len(filterByHash) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByHash(filterByHash...),
		)
	}
	if len(filterByOwner) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByOwner(filterByOwner...),
		)
	}
	if len(filterByStreamHash) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByStreamHash(filterByStreamHash...),
		)
	}
	reader := p.objectmanager.Subscribe(opts...)
	return reader, nil
}

func (p *Provider) RequestStream(
	ctx context.Context,
	rootHash tilde.Hash,
) error {
	recipients, err := p.resolver.Lookup(
		ctx,
		resolver.LookupByHash(rootHash),
	)
	if err != nil {
		return err
	}
	for _, recipient := range recipients {
		go func(recipient *peer.ConnectionInfo) {
			ctx := context.New(
				context.WithTimeout(10 * time.Second),
			)
			_, err := p.objectmanager.Request(ctx, rootHash, recipient)
			if err != nil {
				return
			}
			r, err := p.objectmanager.RequestStream(ctx, rootHash, recipient)
			if err != nil {
				return
			}
			object.ReadAll(r) // nolint: errcheck
			r.Close()
		}(recipient)
	}
	return nil
}

func (p *Provider) Put(
	ctx context.Context,
	obj *object.Object,
) (*object.Object, error) {
	obj = object.Copy(obj)
	if setOwnerS, ok := obj.Data["_setOwner:s"]; ok {
		if setOwner, ok := setOwnerS.(tilde.String); ok {
			switch setOwner {
			case "@peer":
				obj.Metadata.Owner = p.network.GetPeerKey().PublicKey().DID()
			case "@identity":
				// TODO(geoah): fix identity
				// obj.Metadata.Owner = p.local.GetIdentityPublicKey()
			}
		}
		delete(obj.Data, "_setOwner:s")
	}
	if obj.Metadata.Root.IsEmpty() {
		return obj, p.objectmanager.Put(ctx, obj)
	}
	return p.objectmanager.Append(ctx, obj)
}

func (p *Provider) GetFeedRootHash(
	streamRootObjectType string,
) tilde.Hash {
	v := &feed.FeedStreamRoot{
		ObjectType: streamRootObjectType,
		Metadata: object.Metadata{
			Owner: p.network.GetPeerKey().PublicKey().DID(),
		},
	}
	o, _ := object.Marshal(v)
	return o.Hash()
}
