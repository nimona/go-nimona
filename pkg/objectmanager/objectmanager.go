package objectmanager

import (
	"fmt"
	"sync"
	"time"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/stream"
)

const (
	ErrDone    = errors.Error("done")
	ErrTimeout = errors.Error("request timed out")
)

var (
	objectRequestType      = new(object.Request).Type()
	objectResponseType     = new(object.Response).Type()
	streamRequestType      = new(stream.Request).Type()
	streamResponseType     = new(stream.Response).Type()
	streamSubscriptionType = new(stream.Subscription).Type()
	streamAnnouncementType = new(stream.Announcement).Type()
)

//go:generate mockgen -destination=../objectmanagermock/objectmanagermock_generated.go -package=objectmanagermock -source=objectmanager.go
//go:generate mockgen -destination=../objectmanagerpubsubmock/objectmanagerpubsubmock_generated.go -package=objectmanagerpubsubmock -source=pubsub.go
//go:generate genny -in=$GENERATORS/syncmap_named/syncmap.go -out=subscriptions_generated.go -imp=nimona.io/pkg/crypto -pkg=objectmanager gen "KeyType=object.CID ValueType=stream.Subscription SyncmapName=subscriptions"

type (
	ObjectManager interface {
		Put(
			ctx context.Context,
			o *object.Object,
		) (*object.Object, error)
		Request(
			ctx context.Context,
			cid object.CID,
			peer *peer.ConnectionInfo,
		) (*object.Object, error)
		RequestStream(
			ctx context.Context,
			rootCID object.CID,
			recipients ...*peer.ConnectionInfo,
		) (object.ReadCloser, error)
		AddStreamSubscription(
			ctx context.Context,
			rootCID object.CID,
		) error
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
		newRequestID  func() string
		subscriptions *SubscriptionsMap
	}
	Option func(*manager)
)

// Object manager is responsible for:
// * adding objects (Put) to the store
// * adding objects (Put) to the local peer's content cids

func New(
	ctx context.Context,
	net network.Network,
	res resolver.Resolver,
	str objectstore.Store,
) ObjectManager {
	m := &manager{
		newRequestID: func() string {
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

	subs := m.network.Subscribe()

	if cids, err := m.objectstore.GetPinned(); err == nil {
		m.localpeer.PutCIDs(cids...)
	} else {
		logger.Error("error getting pinned objects", log.Error(err))
	}

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

func (m *manager) isWellKnownEphemeral(
	contentType string,
) bool {
	switch contentType {
	case "nimona.io/stream.Announcement",
		"nimona.io/stream.Request",
		"nimona.io/stream.Response",
		"nimona.io/Request",
		"nimona.io/Response":
		return true
	}
	return false
}

// TODO add support for multiple recipients
// TODO this currently needs to be storing objects for it to work.
func (m *manager) RequestStream(
	ctx context.Context,
	rootCID object.CID,
	recipients ...*peer.ConnectionInfo,
) (object.ReadCloser, error) {
	if len(recipients) == 0 {
		return m.objectstore.GetByStream(rootCID)
	}

	rID := m.newRequestID()
	responses := make(chan stream.Response)

	sub := m.network.Subscribe(
		network.FilterByObjectType(streamResponseType),
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
			if streamResp.RequestID != rID {
				continue
			}
			responses <- *streamResp
			return
		}
	}()

	// TODO support more than 1 recipient
	if len(recipients) > 1 {
		panic(errors.Error("currently only a single recipient is supported"))
	}

	// TODO we should first request and store stream root I guess

	req := stream.Request{
		RequestID: rID,
		RootCID:   rootCID,
	}
	if err := m.network.Send(
		ctx,
		req.ToObject(),
		recipients[0].PublicKey,
	); err != nil {
		return nil, err
	}

	var leaves []object.CID

	select {
	case res := <-responses:
		leaves = res.Leaves
	case <-ctx.Done():
		return nil, ErrTimeout
	}

	if err := m.fetchFromLeaves(ctx, leaves, recipients[0]); err != nil {
		return nil, err
	}

	return m.objectstore.GetByStream(rootCID)
}

func (m *manager) fetchFromLeaves(
	ctx context.Context,
	leaves []object.CID,
	recipient *peer.ConnectionInfo,
) error {
	// TODO refactor to remove buffer
	objectCIDs := make(chan object.CID, 1000)

	go func() {
		for _, l := range leaves {
			objectCIDs <- l
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(len(leaves))

	errorChan := make(chan error)
	doneChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(objectCIDs)
	}()

	go func() {
		for objectCID := range objectCIDs {
			// check if we have object stored
			dCtx := context.New(
				context.WithTimeout(3 * time.Second),
			)
			if obj, err := m.objectstore.Get(objectCID); err == nil {
				// TODO consider checking the whole stream for missing objects
				// TODO consider refactoring, or moving into a goroutine
				for _, group := range obj.Metadata.Parents {
					for _, parent := range group {
						wg.Add(1)
						objectCIDs <- parent
					}
				}
				wg.Done()
				continue
			}
			// TODO consider exluding nexted objects
			fullObj, err := m.Request(
				dCtx,
				objectCID,
				recipient,
			)
			if err != nil {
				wg.Done()
				continue
			}

			// TODO check the validity of the object
			// * it should have objects
			// * it should have a stream root cid
			// * should it be signed?
			// * is its policy valid?
			// TODO consider refactoring, or moving into a goroutine
			for _, group := range fullObj.Metadata.Parents {
				for _, parent := range group {
					objectCIDs <- parent
					wg.Add(1)
				}
			}

			// so we should already have its parents.
			if err := m.storeObject(dCtx, fullObj); err != nil {
				// TODO what do we do now?
				wg.Done()
				continue
			}

			m.pubsub.Publish(fullObj)

			wg.Done()
		}
		close(doneChan)
	}()

	select {
	case <-doneChan:
		return nil
	case err := <-errorChan:
		return err
	case <-ctx.Done():
		return ErrTimeout
	}
}

func (m *manager) Request(
	ctx context.Context,
	cid object.CID,
	pr *peer.ConnectionInfo,
) (*object.Object, error) {
	objCh := make(chan *object.Object)
	errCh := make(chan error)

	rID := m.newRequestID()

	sub := m.network.Subscribe(
		network.FilterByObjectType(objectResponseType),
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
			res := &object.Response{}
			if err := res.FromObject(e.Payload); err != nil {
				continue
			}
			if res.RequestID == rID && res.Object != nil {
				objCh <- res.Object
				break
			}
		}
	}()

	req := &object.Request{
		RequestID: rID,
		ObjectCID: cid,
	}
	if err := m.network.Send(
		ctx,
		req.ToObject(),
		pr.PublicKey,
	); err != nil {
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
				log.String("payload.type", env.Payload.Type),
				log.String("payload.cid", env.Payload.CID().String()),
			)

		logger.Debug("handling object")

		// store object
		// TODO why store here?
		if err := m.storeObject(ctx, env.Payload); err != nil {
			return err
		}

		switch env.Payload.Type {
		case objectRequestType:
			go func() {
				hCtx := context.New(
					context.WithParent(ctx),
				)
				if err := m.handleObjectRequest(
					hCtx,
					env,
				); err != nil {
					logger.Warn(
						"could not handle object request",
						log.Error(err),
					)
				}
			}()
			continue
		case streamRequestType:
			go func() {
				hCtx := context.New(
					context.WithParent(ctx),
				)
				if err := m.handleStreamRequest(
					hCtx,
					env,
				); err != nil {
					logger.Warn(
						"could not handle stream request",
						log.Error(err),
					)
				}
			}()
			continue
		case streamSubscriptionType:
			go func() {
				hCtx := context.New(
					context.WithParent(ctx),
				)
				if err := m.handleStreamSubscription(
					hCtx,
					env,
				); err != nil {
					logger.Warn(
						"could not handle stream request",
						log.Error(err),
					)
				}
			}()
			continue
		case streamAnnouncementType:
			go func() {
				hCtx := context.New(
					context.WithParent(ctx),
				)
				if err := m.handleStreamAnnouncement(
					hCtx,
					env,
				); err != nil {
					logger.Warn(
						"could not handle stream announcement",
						log.Error(err),
					)
				}
			}()
			continue
		}

		// publish to pubsub
		m.pubsub.Publish(env.Payload)
	}
}

// Note: please do not .pubsub.Publish() in here
func (m *manager) storeObject(
	ctx context.Context,
	obj *object.Object,
) error {
	// TODO is registered content type, OR is part of a peristed stream
	ok := m.isRegisteredContentType(obj.Type)
	if !ok {
		return nil
	}

	if skip := m.isWellKnownEphemeral(obj.Type); skip {
		return nil
	}

	logger := log.FromContext(ctx)
	objType := obj.Type
	objCID := obj.CID()

	// store object
	if err := m.objectstore.Put(obj); err != nil {
		logger.Error(
			"error trying to persist incoming object",
			log.String("cid", objCID.String()),
			log.String("type", objType),
			log.Error(err),
		)
		// TODO if we failed to store why are we not returning?
	}

	// add CID to local peer
	// we currently only do this if we are dealing with a single object or
	// a stream root
	if obj.Metadata.Stream.IsEmpty() {
		m.localpeer.PutCIDs(objCID)
	}

	return nil
}

func (m *manager) announceStreamChildren(
	ctx context.Context,
	streamCID object.CID,
	children []object.CID,
) {
	logger := log.FromContext(ctx)

	// find ephemeral subscriptions for this stream
	// TODO do we really need ephemeral subscriptions?
	subscribersMap := map[string]struct{}{}
	m.subscriptions.Range(func(_ object.CID, sub *stream.Subscription) bool {
		// TODO check expiry
		subscribersMap[sub.Metadata.Owner.String()] = struct{}{}
		return true
	})

	// find subscriptions that are attached in the stream
	r, err := m.objectstore.GetByStream(streamCID)
	if err != nil {
		return
	}

	for {
		obj, err := r.Read()
		// TODO do we want to return if error is not EOF?
		if err != nil {
			break
		}
		if obj.Type != streamSubscriptionType {
			continue
		}
		if obj.Metadata.Owner.IsEmpty() {
			continue
		}
		subscribersMap[obj.Metadata.Owner.String()] = struct{}{}
	}

	// remove self
	delete(subscribersMap, m.localpeer.GetPrimaryPeerKey().PublicKey().String())

	subscribers := []string{}
	for subscriber := range subscribersMap {
		subscribers = append(subscribers, subscriber)
	}

	logger.Info("trying to announce",
		log.String("streamCID", streamCID.String()),
		log.Any("subscribers", subscribers),
	)

	if len(subscribers) == 0 {
		return
	}

	// notify subscribers
	announcement := stream.Announcement{
		Metadata: object.Metadata{
			Owner: m.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		StreamCID:  streamCID,
		ObjectCIDs: children,
	}
	for _, subscriber := range subscribers {
		// TODO figure out if subscribers are peers or identities? how?
		// TODO verify that subscriber has access to this object/stream
		k := crypto.PublicKey{}
		if err := k.UnmarshalString(subscriber); err != nil {
			logger.Info(
				"error unmarshaling subscriber key",
				log.String("subscriber", subscriber),
				log.Error(err),
			)
			continue
		}
		err := m.network.Send(ctx, announcement.ToObject(), k)
		if err != nil {
			logger.Info(
				"error sending announcement",
				log.Error(err),
				log.String("subscriber", subscriber),
			)
			continue
		}
		logger.Debug(
			"sent announcement",
			log.Any("sub", subscriber),
			log.Error(err),
		)
	}
}

func (m *manager) handleObjectRequest(
	ctx context.Context,
	env *network.Envelope,
) error {
	logger := log.FromContext(ctx).With(
		log.String("method", "objectmanager.handleObjectRequest"),
	)

	req := &object.Request{}
	if err := req.FromObject(env.Payload); err != nil {
		logger.Warn(
			"received invalid object request",
			log.Error(err),
		)
		return err
	}

	logger = logger.With(
		log.String("req.objectID", req.ObjectCID.String()),
	)

	logger.Info("handling object request")

	resp := &object.Response{
		Metadata: object.Metadata{
			Owner: m.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		Object:    nil,
		RequestID: req.RequestID,
	}

	obj, err := m.objectstore.Get(req.ObjectCID)
	if err != nil {
		logger.Error(
			"error getting object to respond with",
			log.Error(err),
		)
		if err != objectstore.ErrNotFound {
			return err
		}
		if sErr := m.network.Send(
			ctx,
			resp.ToObject(),
			env.Sender,
		); err != nil {
			logger.Info(
				"error sending failure response",
				log.Error(sErr),
			)
		}
		return err
	}

	robj := object.Copy(obj)

	resp.Object = robj

	err = m.network.Send(
		ctx,
		resp.ToObject(),
		env.Sender,
	)

	if err != nil {
		logger.Warn(
			"error sending object response",
			log.Error(err),
		)
		return fmt.Errorf("could not send object: %w", err)
	}

	logger.Info(
		"sent object response",
		log.Error(err),
	)

	return nil
}

func (m *manager) handleStreamRequest(
	ctx context.Context,
	env *network.Envelope,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.FromContext(ctx).With(
		log.String("method", "objectmanager.handleStreamRequest"),
	)

	req := &stream.Request{}
	if err := req.FromObject(env.Payload); err != nil {
		return err
	}

	// start response
	res := &stream.Response{
		Metadata: object.Metadata{
			Owner: m.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		RequestID: req.RequestID,
		RootCID:   req.RootCID,
	}

	leaves, err := m.objectstore.GetStreamLeaves(res.RootCID)
	if err != nil && !errors.Is(err, objectstore.ErrNotFound) {
		return err
	}

	res.Leaves = leaves

	if err := m.network.Send(
		ctx,
		res.ToObject(),
		env.Sender,
	); err != nil {
		logger.Warn(
			"could not send response",
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

	for _, rootCID := range sub.RootCIDs {
		// TODO introduce time-to-live for subscriptions
		m.subscriptions.Put(rootCID, sub)
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

	// TODO check if this a stream we care about using ann.StreamCID

	logger := log.FromContext(ctx).With(
		log.String("method", "objectmanager.handleStreamAnnouncement"),
		log.Any("sender", env.Sender),
	)

	logger.Info("got stream announcement ",
		log.Any("cids", ann.ObjectCIDs),
	)

	// check if we already know about these objects
	allKnown := true
	for _, cid := range ann.ObjectCIDs {
		_, err := m.objectstore.Get(cid)
		if err == objectstore.ErrNotFound {
			allKnown = false
			break
		}
	}

	pr, err := m.resolver.Lookup(
		ctx,
		resolver.LookupByPeerKey(env.Sender),
	)
	if err != nil {
		logger.Warn(
			"error looking up sender, will still attempt to send response",
			log.Error(err),
		)
	}

	// still create a connection in case we still have an open connection
	if len(pr) == 0 {
		pr = []*peer.ConnectionInfo{{
			PublicKey: env.Sender,
		}}
	}

	// fetch announced objects and their parents
	if err := m.fetchFromLeaves(
		ctx,
		ann.ObjectCIDs,
		pr[0],
	); err != nil {
		return err
	}

	// if we didn't know about all of them, announce to other subscribers
	if allKnown {
		return nil
	}

	// announce to subscribers
	go m.announceStreamChildren(
		context.New(
			context.WithCorrelationID(ctx.CorrelationID()),
			// context.WithTimeout(5*time.Second),
		),
		ann.StreamCID,
		ann.ObjectCIDs,
	)

	return nil
}

// Put stores a given object
// TODO(geoah) what happened if the stream graph is not complete? Do we care?
func (m *manager) Put(
	ctx context.Context,
	o *object.Object,
) (*object.Object, error) {
	// if this is not ours, just persist it
	owner := o.Metadata.Owner
	ownPeer := false
	ownIdentity := false
	if !owner.IsEmpty() {
		if k := m.localpeer.GetPrimaryPeerKey().PublicKey(); !k.IsEmpty() {
			ownPeer = owner.Equals(k)
		}
		if k := m.localpeer.GetPrimaryIdentityKey(); !k.IsEmpty() {
			ownIdentity = owner.Equals(k.PublicKey())
		}
	}
	if !ownPeer && !ownIdentity {
		// add to store
		if err := m.storeObject(ctx, o); err != nil {
			return nil, err
		}
		// publish to pubsub
		// TODO why do we publish this?
		m.pubsub.Publish(o)
		return o, nil
	}

	// Note: Please don't add owners as it messes with hypothetical objects
	// TODO sign for owner = identity as well
	// figure out if we need to add parents to the object
	streamCID := o.Metadata.Stream
	if !streamCID.IsEmpty() && len(o.Metadata.Parents) == 0 {
		leaves, err := m.objectstore.GetStreamLeaves(streamCID)
		if err != nil {
			return nil, err
		}
		if len(leaves) == 0 {
			leaves = []object.CID{
				streamCID,
			}
		}
		object.SortCIDs(leaves)
		o.Metadata.Parents = object.Parents{
			"*": leaves,
		}
	}

	// add to store
	if err := m.storeObject(ctx, o); err != nil {
		return nil, err
	}

	if !streamCID.IsEmpty() {
		// announce to subscribers
		m.announceStreamChildren(
			context.New(
				context.WithCorrelationID(ctx.CorrelationID()),
				// TODO timeout?
			),
			o.Metadata.Stream,
			[]object.CID{
				o.CID(),
			},
		)
	}

	// publish to pubsub
	m.pubsub.Publish(o)
	return o, nil
}

func (m *manager) AddStreamSubscription(
	ctx context.Context,
	rootCID object.CID,
) error {
	r, err := m.objectstore.GetByStream(rootCID)
	if err != nil {
		return fmt.Errorf("error trying to get stream objects, %w", err)
	}

	pub := m.localpeer.GetPrimaryPeerKey().PublicKey()

	for {
		o, err := r.Read()
		if errors.Is(err, object.ErrReaderDone) {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading stream objects, %w", err)
		}

		if o.Type != streamSubscriptionType {
			continue
		}

		s := &stream.Subscription{}
		if err := s.FromObject(o); err != nil {
			continue
		}

		if !s.Metadata.Owner.Equals(pub) {
			continue
		}

		// TODO check if the subscription has expired

		// already subscribed

		return nil
	}

	s := &stream.Subscription{
		Metadata: object.Metadata{
			Owner:  pub,
			Stream: rootCID,
		},
		RootCIDs: []object.CID{
			rootCID,
		},
		// TODO add expiry
	}

	if _, err := m.Put(ctx, s.ToObject()); err != nil {
		return fmt.Errorf("error storing subscription, %w", err)
	}

	return nil
}

func (m *manager) Subscribe(
	lookupOptions ...LookupOption,
) ObjectSubscription {
	options := newLookupOptions(lookupOptions...)
	return m.pubsub.Subscribe(options.Filters...)
}
