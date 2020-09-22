package objectmanager

import (
	"strings"
	"sync"
	"time"

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
	streamAnnouncementType = stream.Announcement{}.GetType()
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
			excludeNested bool,
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

// Object manager is responsible for:
// * adding objects (Put) to the store
// * adding objects (Put) to the local peer's content hashes

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
		hashes, err := m.objectstore.GetPinned()
		if err != nil {
			logger.Error("error getting pinned objects", log.Error(err))
			return
		}
		if len(hashes) == 0 {
			return
		}
		m.localpeer.PutContentHashes(hashes...)
	}()

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
// TODO this currently needs to be storing objects for it to work.
func (m *manager) RequestStream(
	ctx context.Context,
	rootHash object.Hash,
	recipients ...*peer.Peer,
) (object.ReadCloser, error) {
	if len(recipients) == 0 {
		return m.objectstore.GetByStream(rootHash)
	}

	nonce := m.newNonce()
	responses := make(chan stream.Response)

	sub := m.network.Subscribe(
		network.FilterByObjectType(new(stream.Response).GetType()),
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

	// TODO support more than 1 recipient
	if len(recipients) > 1 {
		panic(errors.New("currently only a single recipient is supported"))
	}

	req := stream.Request{
		Nonce:    nonce,
		RootHash: rootHash,
	}
	if err := m.network.Send(
		ctx,
		req.ToObject(),
		recipients[0],
	); err != nil {
		return nil, err
	}

	// TODO refactor to remove buffer
	objectHashes := make(chan object.Hash, 1000)

	wg := sync.WaitGroup{}

	select {
	case res := <-responses:
		// TODO consider refactoring, or moving into a goroutine
		for _, leaf := range res.Leaves {
			objectHashes <- leaf
		}
		wg.Add(len(res.Leaves))
	case <-ctx.Done():
		return nil, ErrTimeout
	}

	errorChan := make(chan error)
	doneChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(objectHashes)
	}()

	go func() {
		for objectHash := range objectHashes {
			// check if we have object stored
			dCtx := context.New(
				context.WithTimeout(3 * time.Second),
			)
			if _, err := m.objectstore.Get(objectHash); err == nil {
				// TODO consider checking the whole stream for missing objects
				// parents := obj.GetParents()
				// TODO consider refactoring, or moving into a goroutine
				// for _, parent := range parents {
				// 	objectHashes <- parent
				// }
				// wg.Add(len(parents))
				wg.Done()
				continue
			}
			// TODO consider exluding nexted objects
			fullObj, err := m.Request(
				dCtx,
				objectHash,
				recipients[0],
				false,
			)
			if err != nil {
				wg.Done()
				continue
			}

			parents := fullObj.GetParents()
			// TODO check the validity of the object
			// * it should have objects
			// * it should have a stream root hash
			// * should it be signed?
			// * is its policy valid?
			// TODO consider refactoring, or moving into a goroutine
			for _, parent := range parents {
				objectHashes <- parent
				wg.Add(len(parents))
			}

			// so we should already have its parents.
			if err := m.storeObject(dCtx, *fullObj); err != nil {
				// TODO what do we do now?
				wg.Done()
				continue
			}

			wg.Done()
		}
		close(doneChan)
	}()

	select {
	case <-doneChan:
		return m.objectstore.GetByStream(rootHash)
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, ErrTimeout
	}
}

func (m *manager) Request(
	ctx context.Context,
	hash object.Hash,
	pr *peer.Peer,
	excludeNested bool,
) (*object.Object, error) {
	objCh := make(chan *object.Object)
	errCh := make(chan error)

	sub := m.network.Subscribe(
	// network.FilterByObjectHash(hash),
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
		ObjectHash:            hash,
		ExcludedNestedObjects: excludeNested,
	}.ToObject(), pr); err != nil {
		return nil, err
	}

	select {
	case err := <-errCh:
		return nil, err
	case obj := <-objCh:
		// TODO verify we have all parents?
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
		case streamAnnouncementType:
			if err := m.handleStreamAnnouncement(
				ctx,
				env,
			); err != nil {
				logger.Warn(
					"could not handle stream announcement",
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
	// TODO is registered content type, OR is part of a peristed stream
	ok := m.isRegisteredContentType(obj.GetType())
	if !ok {
		return nil
	}

	logger := log.FromContext(ctx)
	objType := obj.GetType()
	objHash := obj.Hash()

	// TODO should we be de-reffing the object? I think so at least.

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
	if err := m.objectstore.Put(*mainObj); err != nil {
		logger.Error(
			"error trying to persist incoming object",
			log.String("hash", objHash.String()),
			log.String("type", objType),
			log.Error(err),
		)
	}

	if m.localpeer.GetPrimaryIdentityKey().IsEmpty() {
		return nil
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
		// TODO figure out if subscribers are peers or identities? how?
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

	robj := &obj

	switch req.ExcludedNestedObjects {
	case true:
		robj, _, err = object.UnloadReferences(ctx, obj)
		if err != nil {
			return err
		}
	case false:
		robj, err = object.LoadReferences(ctx, hash, func(
			ctx context.Context, hash object.Hash,
		) (*object.Object, error) {
			obj, err := m.objectstore.Get(hash)
			return &obj, err
		})
		if err != nil {
			return err
		}
	}

	if err := m.network.Send(
		ctx,
		*robj,
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

	// get the whole stream
	or, err := m.objectstore.GetByStream(req.RootHash)
	if err != nil {
		return err
	}
	os, err := object.ReadAll(or)
	if err != nil {
		return err
	}
	leaves := stream.GetStreamLeaves(os)
	leaveHashes := []object.Hash{}
	for _, o := range leaves {
		leaveHashes = append(leaveHashes, o.Hash())
	}

	res := &stream.Response{
		Metadata: object.Metadata{
			Owner: m.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		Nonce: req.Nonce,
		// RootHash: req.RootHash, // TODO ADD THIS
		Leaves: leaveHashes,
	}

	if err := m.network.Send(
		ctx,
		res.ToObject(),
		// TODO we should probably resolve this peer
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

func (m *manager) handleStreamAnnouncement(
	ctx context.Context,
	env *network.Envelope,
) error {
	ann := &stream.Announcement{}
	if err := ann.FromObject(env.Payload); err != nil {
		return err
	}

	// TODO check if this a stream we care about
	// TODO maybe verify nested objects
	// TODO request missing objects if missing

	// go through all objects
	for _, obj := range ann.Objects {
		// store object
		if err := m.storeObject(ctx, *obj); err != nil {
			// TODO should we move on?
			return err
		}
		// publish to pubsub
		m.pubsub.Publish(*obj)
	}

	return nil
}

// Put stores a given object
// TODO(geoah) what happened if the stream graph is not complete? Do we care?
func (m *manager) Put(
	ctx context.Context,
	o object.Object,
) (object.Object, error) {
	// if this is not ours, just persist it
	// TODO check identity as well?
	if o.GetOwner() != m.localpeer.GetPrimaryPeerKey().PublicKey() {
		// add to store
		if err := m.storeObject(ctx, o); err != nil {
			return o, err
		}
		// publish to pubsub
		m.pubsub.Publish(o)
		return o, nil
	}
	// Note: Please don't add owners as it messes with hypothetical objects
	// if o.GetOwner() == m.localpeer.GetPrimaryPeerKey().PublicKey() {
	// 	sig, err := object.NewSignature(
	// 		m.localpeer.GetPrimaryPeerKey(),
	// 		o,
	// 	)
	// 	if err != nil {
	// 		return o, errors.Wrap(
	// 			errors.New("unable to sign object"),
	// 			err,
	// 		)
	// 	}
	// 	o = o.SetSignature(sig)
	// }
	// TODO sign for owner = identity as well
	// figure out if we need to add parents to the object
	streamHash := o.GetStream()
	if !streamHash.IsEmpty() && len(o.GetParents()) == 0 {
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
	// TODO consider removing the err return from announceObject
	// TODO make async
	go m.announceObject(
		context.New(
			context.WithTimeout(1*time.Second),
		),
		o,
	) // nolint: errcheck
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
