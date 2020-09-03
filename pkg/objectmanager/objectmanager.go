package objectmanager

import (
	"strings"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/stream"
)

var (
	ErrDone    = errors.New("done")
	ErrTimeout = errors.New("request timed out")
)

var (
	objectRequestType      = object.Request{}.GetType()
	streamRequestType      = stream.Request{}.GetType()
	streamSubscriptionType = stream.Subscription{}.GetType()
)

//go:generate $GOBIN/mockgen -destination=../objectmanagermock/objectmanagermock_generated.go -package=objectmanagermock -source=objectmanager.go
//go:generate $GOBIN/mockgen -destination=../objectmanagerpubsubmock/objectmanagerpubsubmock_generated.go -package=objectmanagerpubsubmock -source=pubsub_generated.go
//go:generate $GOBIN/genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub_generated.go -pkg objectmanager -imp=nimona.io/pkg/object gen "ObjectType=object.Object Name=Object name=object"
//go:generate $GOBIN/genny -in=$GENERATORS/syncmap_named/syncmap.go -out=subscriptions_generated.go -imp=nimona.io/pkg/crypto -pkg=objectmanager gen "KeyType=object.Hash ValueType=stream.Subscription SyncmapName=subscriptions"

type (
	ObjectManager interface {
		Put(
			ctx context.Context,
			o object.Object,
		) (object.Object, error)
		Request(
			ctx context.Context,
			hash object.Hash,
			peer *peer.Peer,
		) (*object.Object, error)
		RequestStream(
			ctx context.Context,
			rootHash object.Hash,
			recipients ...*peer.Peer,
		) (object.ReadCloser, error)
		Subscribe(
			lookupOptions ...LookupOption,
		) ObjectSubscription
	}

	manager struct {
		network       network.Network
		objectstore   objectstore.Store
		localpeer     localpeer.LocalPeer
		resolver      resolver.Resolver
		pubsub        ObjectPubSub
		newNonce      func() string
		subscriptions *SubscriptionsMap
	}
	Option func(*manager)
)

func New(
	ctx context.Context,
	net network.Network,
	res resolver.Resolver,
	str objectstore.Store,
) ObjectManager {
	m := &manager{
		newNonce: func() string {
			return rand.String(16)
		},
		pubsub:        NewObjectPubSub(),
		subscriptions: &SubscriptionsMap{},
		network:       net,
		resolver:      res,
		localpeer:     net.LocalPeer(),
		objectstore:   str,
	}

	logger := log.
		FromContext(ctx).
		Named("objectmanager").
		With(
			log.String("method", "objectmanager.New"),
		)

	subs := m.network.Subscribe(
		network.FilterByObjectType("**"),
	)

	go func() {
		if err := m.handleObjects(subs); err != nil {
			logger.Error("handling object requests failed", log.Error(err))
		}
	}()

	return m
}

func (m *manager) isRegisteredContentType(
	contentType string,
) bool {
	contentTypes := m.localpeer.GetContentTypes()
	for _, ct := range contentTypes {
		if contentType == ct {
			return true
		}
	}
	return false
}

// TODO add support for multiple recipients
func (m *manager) RequestStream(
	ctx context.Context,
	rootHash object.Hash,
	recipients ...*peer.Peer,
) (object.ReadCloser, error) {
	nonce := m.newNonce()
	responses := make(chan stream.Response)

	sub := m.network.Subscribe(
		network.FilterByObjectType(stream.Response{}.GetType()),
	)

	go func() {
		defer sub.Cancel()
		for {
			env, err := sub.Next()
			if err != nil || env == nil {
				return
			}
			streamResp := &stream.Response{}
			if err := streamResp.FromObject(env.Payload); err != nil {
				continue
			}
			if streamResp.Nonce != nonce {
				continue
			}
			responses <- *streamResp
			return
		}
	}()

	if len(recipients) == 0 {
		return m.objectstore.GetByStream(rootHash)
	}

	if len(recipients) > 1 {
		panic(errors.New("currently only a single recipient is supported"))
	}

	req := stream.Request{
		RootHash: rootHash,
	}
	if err := m.network.Send(
		ctx,
		req.ToObject(),
		recipients[0],
	); err != nil {
		return nil, err
	}

	var objectHashes []object.Hash

	select {
	case res := <-responses:
		objectHashes = append(objectHashes, res.Children...)

	case <-ctx.Done():
		return nil, ErrTimeout
	}

	requestHandler := func(
		ctx context.Context,
		objectHash object.Hash,
	) (*object.Object, error) {
		// TODO(geoah) check store before requesting the object
		// obj, err := m.objectstore.Get(objectHash)
		// if err == nil {
		// 	return &obj, nil
		// }
		return m.Request(
			ctx,
			objectHash,
			recipients[0],
		)
	}

	objectChan := make(chan *object.Object)
	errorChan := make(chan error)
	closeChan := make(chan struct{})

	reader := object.NewReadCloser(
		ctx,
		objectChan,
		errorChan,
		closeChan,
	)

	go func() {
		defer close(objectChan)
		defer close(errorChan)
		for _, objectHash := range objectHashes {
			fullObj, err := object.LoadReferences(
				ctx,
				objectHash,
				requestHandler,
			)
			if err != nil {
				errorChan <- err
				return
			}
			// TODO check the validity of each event, they should be ordered
			// so we should already have its parents.
			select {
			case objectChan <- fullObj:
				// all good
			case <-ctx.Done():
				return
			case <-closeChan:
				return
			}
		}
	}()

	return reader, nil
}

func (m *manager) Request(
	ctx context.Context,
	hash object.Hash,
	pr *peer.Peer,
) (*object.Object, error) {
	objCh := make(chan *object.Object)
	errCh := make(chan error)

	sub := m.network.Subscribe(
		network.FilterByObjectHash(hash),
	)
	defer sub.Cancel()

	go func() {
		for {
			e, err := sub.Next()
			if err != nil {
				errCh <- err
				break
			}
			if e == nil {
				break
			}
			if e.Payload.Hash() == hash {
				objCh <- &e.Payload
				break
			}
		}
	}()

	if err := m.network.Send(ctx, object.Request{
		ObjectHash: hash,
	}.ToObject(), pr); err != nil {
		return nil, err
	}

	select {
	case err := <-errCh:
		return nil, err
	case obj := <-objCh:
		return obj, nil
	case <-ctx.Done():
		return nil, ErrTimeout
	}
}

func (m *manager) handleObjects(
	sub network.EnvelopeSubscription,
) error {
	for {
		env, err := sub.Next()
		if err != nil {
			return err
		}

		ctx := context.Background()
		logger := log.
			FromContext(ctx).
			Named("objectmanager").
			With(
				log.String("method", "objectmanager.handleObjects"),
				log.String("payload", env.Payload.GetType()),
			)

		logger.Debug("handling object")

		// store object
		if err := m.storeObject(ctx, env.Payload); err != nil {
			return err
		}

		switch env.Payload.GetType() {
		case objectRequestType:
			if err := m.handleObjectRequest(
				ctx,
				env,
			); err != nil {
				logger.Warn(
					"could not handle object request",
					log.Error(err),
				)
			}
		case streamRequestType:
			if err := m.handleStreamRequest(
				ctx,
				env,
			); err != nil {
				logger.Warn(
					"could not handle stream request",
					log.Error(err),
				)
			}
		case streamSubscriptionType:
			if err := m.handleStreamSubscription(
				ctx,
				env,
			); err != nil {
				logger.Warn(
					"could not handle stream request",
					log.Error(err),
				)
			}
		}

		// publish to pubsub
		m.pubsub.Publish(env.Payload)
	}
}

func (m *manager) storeObject(
	ctx context.Context,
	obj object.Object,
) error {
	ok := m.isRegisteredContentType(obj.GetType())
	if !ok {
		return nil
	}

	logger := log.FromContext(ctx)
	objType := obj.GetType()
	objHash := obj.Hash()

	// deref nested objects
	mainObj, refObjs, err := object.UnloadReferences(ctx, obj)
	if err != nil {
		logger.Error(
			"error unloading nested objects",
			log.String("hash", objHash.String()),
			log.String("type", objType),
			log.Error(err),
		)
		return err
	}

	// store nested objects
	for _, refObj := range refObjs {
		// TODO reconsider ttls for nested objects
		if err := m.objectstore.PutWithTimeout(refObj, 0); err != nil {
			logger.Error(
				"error trying to persist incoming nested object",
				log.String("hash", refObj.Hash().String()),
				log.String("type", refObj.GetType()),
				log.Error(err),
			)
		}
	}

	// store primary object
	if err := m.objectstore.PutWithTimeout(*mainObj, 0); err != nil {
		logger.Error(
			"error trying to persist incoming object",
			log.String("hash", objHash.String()),
			log.String("type", objType),
			log.Error(err),
		)
	}

	// TODO check if object already exists in feed

	// add to feed
	feedStreamHash := getFeedRootHash(
		m.localpeer.GetPrimaryIdentityKey().PublicKey(),
		getTypeForFeed(objType),
	)
	feedEvent := feed.Added{
		Metadata: object.Metadata{
			Stream: feedStreamHash,
		},
		ObjectHash: []object.Hash{
			objHash,
		},
	}.ToObject()
	or, err := m.objectstore.GetByStream(feedStreamHash)
	if err != nil && err != objectstore.ErrNotFound {
		return err
	}
	if err == objectstore.ErrNotFound {
		feedEvent = feedEvent.SetParents([]object.Hash{
			feedStreamHash,
		})
	} else {
		os, err := object.ReadAll(or)
		if err != nil {
			return err
		}
		if len(os) > 0 {
			parents := stream.GetStreamLeaves(os)
			parentHashes := make([]object.Hash, len(parents))
			for i, p := range parents {
				parentHashes[i] = p.Hash()
			}
			feedEvent = feedEvent.SetParents(parentHashes)
		}
	}
	if err := m.objectstore.Put(feedEvent); err != nil {
		return err
	}

	return nil
}

// TODO should announceObject be best effort and not return an error?
func (m *manager) announceObject(
	ctx context.Context,
	obj object.Object,
) error {
	// if this is not part of a stream, we're done here
	if obj.GetStream().IsEmpty() {
		return nil
	}

	// find ephemeral subscriptions for this stream
	subscribersMap := map[crypto.PublicKey]struct{}{}
	m.subscriptions.Range(func(_ object.Hash, sub *stream.Subscription) bool {
		// TODO check expiry
		subscribersMap[sub.Metadata.Owner] = struct{}{}
		return true
	})

	// find subscriptions that are attached in the stream
	r, err := m.objectstore.GetByStream(obj.GetStream())
	if err != nil {
		return err
	}
	for {
		obj, err := r.Read()
		// TODO do we want to return if error is not EOF?
		if err != nil {
			break
		}
		if obj.GetType() != streamSubscriptionType {
			continue
		}
		if obj.GetOwner().IsEmpty() {
			continue
		}
		subscribersMap[obj.GetOwner()] = struct{}{}
	}
	subscribers := []crypto.PublicKey{}
	for subscriber := range subscribersMap {
		subscribers = append(subscribers, subscriber)
	}

	if len(subscribers) == 0 {
		return nil
	}

	// notify subscribers
	announcement := stream.Announcement{
		Metadata: object.Metadata{
			Owner: m.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		Objects: []*object.Object{
			&obj,
		},
	}.ToObject()
	for _, subscriber := range subscribers {
		// TODO verify that subscriber has access to this object/stream
		peers, err := m.resolver.Lookup(
			ctx,
			resolver.LookupByOwner(subscriber),
		)
		if err != nil {
			// TODO log error
			continue
		}
		for peer := range peers {
			if err := m.network.Send(ctx, announcement, peer); err != nil {
				// TODO log error
				continue
			}
		}
	}

	return nil
}

func (m *manager) handleObjectRequest(
	ctx context.Context,
	env *network.Envelope,
) error {
	req := &object.Request{}
	if err := req.FromObject(env.Payload); err != nil {
		return err
	}

	hash := req.ObjectHash
	obj, err := m.objectstore.Get(hash)
	if err != nil {
		return err
	}

	if err := m.network.Send(
		ctx,
		obj,
		&peer.Peer{
			Metadata: object.Metadata{
				Owner: env.Sender,
			},
		},
	); err != nil {
		return errors.Wrap(
			errors.Error("could not send object"),
			err,
		)
	}

	return nil
}

func (m *manager) handleStreamRequest(
	ctx context.Context,
	env *network.Envelope,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.FromContext(ctx)

	req := &stream.Request{}
	if err := req.FromObject(env.Payload); err != nil {
		return err
	}

	// get the entire graph for this stream
	or, err := m.objectstore.GetByStream(req.Metadata.Stream)
	if err != nil {
		return err
	}

	// get only the object hashes
	hs := []object.Hash{}
	for {
		o, err := or.Read()
		if err == object.ErrReaderDone {
			break
		}
		if err != nil {
			return err
		}
		hs = append(hs, o.Hash())
	}

	res := &stream.Response{
		Metadata: object.Metadata{
			Stream: req.Metadata.Stream,
		},
		Nonce:    req.Nonce,
		Children: hs,
	}

	if err := m.network.Send(
		ctx,
		res.ToObject(),
		&peer.Peer{
			Metadata: object.Metadata{
				Owner: env.Sender,
			},
		},
	); err != nil {
		logger.Warn(
			"objectmanager.handleStreamRequest could not send response",
			log.Error(err),
		)
		return err
	}

	return nil
}

func (m *manager) handleStreamSubscription(
	ctx context.Context,
	env *network.Envelope,
) error {
	sub := &stream.Subscription{}
	if err := sub.FromObject(env.Payload); err != nil {
		return err
	}

	for _, rootHash := range sub.RootHashes {
		// TODO introduce time-to-live for subscriptions
		m.subscriptions.Put(rootHash, sub)
	}

	return nil
}

// Put stores a given object
// TODO(geoah) what happened if the stream graph is not complete? Do we care?
func (m *manager) Put(
	ctx context.Context,
	o object.Object,
) (object.Object, error) {
	// add owners
	// TODO should we be adding owners?
	if ik := m.localpeer.GetPrimaryIdentityKey(); !ik.IsEmpty() {
		o = o.SetOwner(
			ik.PublicKey(),
		)
	} else {
		o = o.SetOwner(
			m.localpeer.GetPrimaryPeerKey().PublicKey(),
		)
	}
	// figure out if we need to add parents to the object
	streamHash := o.GetStream()
	if !streamHash.IsEmpty() {
		or, err := m.objectstore.GetByStream(streamHash)
		if err != nil && err != objectstore.ErrNotFound {
			return o, err
		}
		os, err := object.ReadAll(or)
		if err != nil {
			return o, err
		}
		if len(os) > 0 {
			parents := stream.GetStreamLeaves(os)
			parentHashes := make([]object.Hash, len(parents))
			for i, p := range parents {
				parentHashes[i] = p.Hash()
			}
			o = o.SetParents(parentHashes)
		}
	}
	// add to store
	if err := m.storeObject(ctx, o); err != nil {
		return o, err
	}
	// announce to subscribers
	if err := m.announceObject(ctx, o); err != nil {
		return o, err
	}
	// publish to pubsub
	m.pubsub.Publish(o)
	return o, nil
}

func (m *manager) Subscribe(
	lookupOptions ...LookupOption,
) ObjectSubscription {
	options := newLookupOptions(lookupOptions...)
	return m.pubsub.Subscribe(options.Filters...)
}

func getFeedRootHash(owner crypto.PublicKey, feedType string) object.Hash {
	r := feed.FeedStreamRoot{
		Type: feedType,
		Metadata: object.Metadata{
			Owner: owner,
		},
	}
	return r.ToObject().Hash()
}

func getTypeForFeed(objectType string) string {
	pt := object.ParseType(objectType)
	return strings.TrimLeft(pt.Namespace+"/"+pt.Object, "/")
}
