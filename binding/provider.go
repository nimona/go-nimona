package main

import (
	"errors"
	"strings"
	"time"

	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/version"
)

type (
	Provider struct {
		local         localpeer.LocalPeer
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
	local := d.LocalPeer()
	net := d.Network()
	str := d.ObjectStore().(*sqlobjectstore.Store)
	res := d.Resolver()
	man := d.ObjectManager()

	log.DefaultLogger.SetLogLevel(nConfig.LogLevel)

	logger = logger.With(
		log.String("peer.publicKey", local.GetPrimaryPeerKey().PublicKey().String()),
		log.Strings("peer.addresses", local.GetAddresses()),
	)

	logger.Error(
		"ready",
		log.Any("addresses", local.GetAddresses()),
	)

	local.PutContentTypes(initRequest.ContentTypes...)

	return &Provider{
		local:         local,
		network:       net,
		resolver:      res,
		objectstore:   str,
		objectmanager: man,
		logger:        logger,
	}
}

func (p *Provider) GetConnectionInfo() *peer.ConnectionInfo {
	return p.local.ConnectionInfo()
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
	filterByCID := []object.CID{}
	filterByOwner := []crypto.PublicKey{}
	filterByStreamCID := []object.CID{}
	for _, lookup := range req.Lookups {
		parts := strings.Split(lookup, ":")
		if len(parts) < 2 {
			return nil, errors.New("invalid lookup query")
		}
		prefix := parts[0]
		value := strings.Join(parts[1:], ":")
		switch prefix {
		case "type":
			filterByType = append(
				filterByType,
				value,
			)
		case "cid":
			filterByCID = append(
				filterByCID,
				object.CID(value),
			)
		case "owner":
			k := crypto.PublicKey{}
			k.UnmarshalString(value) // nolint: errcheck
			filterByOwner = append(
				filterByOwner,
				k,
			)
		case "stream":
			filterByStreamCID = append(
				filterByStreamCID,
				object.CID(value),
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
	if len(filterByCID) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByCID(filterByCID...),
		)
	}
	if len(filterByOwner) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByOwner(filterByOwner...),
		)
	}
	if len(filterByStreamCID) > 0 {
		opts = append(
			opts,
			sqlobjectstore.FilterByStreamCID(filterByStreamCID...),
		)
	}
	return p.objectstore.Filter(opts...)
}

type SubscribeRequest struct {
	Lookups []string `json:"lookups"`
}

// payload should start with one of the following:
// - type:<type>
// - cid:<cid>
// - stream:<rootHash>
// - owner:<publicKey>
func (p *Provider) Subscribe(
	ctx context.Context,
	req SubscribeRequest,
) (object.ReadCloser, error) {
	opts := []objectmanager.LookupOption{}
	filterByType := []string{}
	filterByCID := []object.CID{}
	filterByOwner := []crypto.PublicKey{}
	filterByStreamCID := []object.CID{}
	for _, lookup := range req.Lookups {
		parts := strings.Split(lookup, ":")
		if len(parts) < 2 {
			return nil, errors.New("invalid lookup query")
		}
		prefix := parts[0]
		value := strings.Join(parts[1:], ":")
		switch prefix {
		case "type":
			filterByType = append(
				filterByType,
				value,
			)
		case "cid":
			filterByCID = append(
				filterByCID,
				object.CID(value),
			)
		case "owner":
			k := crypto.PublicKey{}
			k.UnmarshalString(value) // nolint: errcheck
			filterByOwner = append(
				filterByOwner,
				k,
			)
		case "stream":
			filterByStreamCID = append(
				filterByStreamCID,
				object.CID(value),
			)
		}
	}
	if len(filterByType) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByObjectType(filterByType...),
		)
	}
	if len(filterByCID) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByCID(filterByCID...),
		)
	}
	if len(filterByOwner) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByOwner(filterByOwner...),
		)
	}
	if len(filterByStreamCID) > 0 {
		opts = append(
			opts,
			objectmanager.FilterByStreamCID(filterByStreamCID...),
		)
	}
	reader := p.objectmanager.Subscribe(opts...)
	return reader, nil
}

func (p *Provider) RequestStream(
	ctx context.Context,
	rootHash object.CID,
) error {
	recipients, err := p.resolver.Lookup(
		ctx,
		resolver.LookupByCID(rootHash),
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
		if setOwner, ok := setOwnerS.(object.String); ok {
			switch setOwner {
			case "@peer":
				obj.Metadata.Owner = p.local.GetPrimaryPeerKey().PublicKey()
			case "@identity":
				obj.Metadata.Owner = p.local.GetIdentityPublicKey()
			}
		}
		delete(obj.Data, "_setOwner:s")
	}
	return p.objectmanager.Put(ctx, obj)
}

func (p *Provider) GetFeedRootCID(
	streamRootObjectType string,
) object.CID {
	v := &feed.FeedStreamRoot{
		ObjectType: streamRootObjectType,
		Metadata: object.Metadata{
			Owner: p.local.GetPrimaryPeerKey().PublicKey(),
		},
	}
	return v.ToObject().CID()
}
