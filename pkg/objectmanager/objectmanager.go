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
	"nimona.io/pkg/log"
	"nimona.io/pkg/mesh"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/stream"
	"nimona.io/pkg/tilde"
)

const (
	ErrDone        = errors.Error("done")
	ErrTimeout     = errors.Error("request timed out")
	ErrMissingRoot = errors.Error("missing root")
)

//go:generate mockgen -destination=../objectmanagermock/objectmanagermock_generated.go -package=objectmanagermock -source=objectmanager.go
//go:generate mockgen -destination=../objectmanagerpubsubmock/objectmanagerpubsubmock_generated.go -package=objectmanagerpubsubmock -source=pubsub.go
//go:generate genny -in=$GENERATORS/syncmap_named/syncmap.go -out=subscriptions_generated.go -imp=nimona.io/pkg/crypto -pkg=objectmanager gen "KeyType=tilde.Digest ValueType=stream.Subscription SyncmapName=subscriptions"

type (
	ObjectManager interface {
		Put(
			ctx context.Context,
			o *object.Object,
		) error
		Append(
			ctx context.Context,
			o *object.Object,
		) (*object.Object, error)
		Request(
			ctx context.Context,
			hash tilde.Digest,
			peer *peer.ConnectionInfo,
		) (*object.Object, error)
		RequestStream(
			ctx context.Context,
			rootHash tilde.Digest,
			recipients ...*peer.ConnectionInfo,
		) (object.ReadCloser, error)
		AddStreamSubscription(
			ctx context.Context,
			rootHash tilde.Digest,
		) error
		Subscribe(
			lookupOptions ...LookupOption,
		) ObjectSubscription
	}

	manager struct {
		mesh          mesh.Mesh
		objectstore   objectstore.Store
		resolver      resolver.Resolver
		pubsub        ObjectPubSub
		newRequestID  func() string
		subscriptions *SubscriptionsMap
	}
	Option func(*manager)
)

// Object manager is responsible for:
// * adding objects (Put) to the store
// * adding objects (Put) to the local peer's content hashes

func New(
	ctx context.Context,
	msh mesh.Mesh,
	res resolver.Resolver,
	str objectstore.Store,
) ObjectManager {
	m := &manager{
		newRequestID: func() string {
			return rand.String(16)
		},
		pubsub:        NewObjectPubSub(),
		subscriptions: &SubscriptionsMap{},
		mesh:          msh,
		resolver:      res,
		objectstore:   str,
	}

	logger := log.
		FromContext(ctx).
		Named("objectmanager").
		With(
			log.String("method", "objectmanager.New"),
		)

	subs := m.mesh.Subscribe()

	go func() {
		if err := m.handleObjects(subs); err != nil {
			logger.Error("handling object requests failed", log.Error(err))
		}
	}()

	return m
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
	rootHash tilde.Digest,
	recipients ...*peer.ConnectionInfo,
) (object.ReadCloser, error) {
	if len(recipients) == 0 {
		return m.objectstore.GetByStream(rootHash)
	}

	rID := m.newRequestID()
	responses := make(chan stream.Response)

	sub := m.mesh.Subscribe(
		mesh.FilterByObjectType(stream.ResponseType),
	)

	go func() {
		defer sub.Cancel()
		for {
			env, err := sub.Next()
			if err != nil || env == nil {
				return
			}
			streamResp := &stream.Response{}
			if err := object.Unmarshal(env.Payload, streamResp); err != nil {
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

	req := &stream.Request{
		RequestID: rID,
		RootHash:  rootHash,
	}
	ro, err := object.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err := m.mesh.Send(
		ctx,
		ro,
		recipients[0].PublicKey,
	); err != nil {
		return nil, err
	}

	var leaves []tilde.Digest

	select {
	case res := <-responses:
		leaves = res.Leaves
	case <-ctx.Done():
		return nil, ErrTimeout
	}

	if err := m.fetchFromLeaves(ctx, leaves, recipients[0]); err != nil {
		return nil, err
	}

	return m.objectstore.GetByStream(rootHash)
}

func (m *manager) fetchFromLeaves(
	ctx context.Context,
	leaves []tilde.Digest,
	recipient *peer.ConnectionInfo,
) error {
	// TODO refactor to remove buffer
	objectHashes := make(chan tilde.Digest, 1000)

	go func() {
		for _, l := range leaves {
			objectHashes <- l
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(len(leaves))

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
			if obj, err := m.objectstore.Get(objectHash); err == nil {
				// TODO consider checking the whole stream for missing objects
				// TODO consider refactoring, or moving into a goroutine
				for _, group := range obj.Metadata.Parents {
					for _, parent := range group {
						wg.Add(1)
						objectHashes <- parent
					}
				}
				wg.Done()
				continue
			}
			// TODO consider exluding nexted objects
			fullObj, err := m.Request(
				dCtx,
				objectHash,
				recipient,
			)
			if err != nil {
				wg.Done()
				continue
			}

			// TODO check the validity of the object
			// * it should have objects
			// * it should have a stream root hash
			// * should it be signed?
			// * is its policy valid?
			// TODO consider refactoring, or moving into a goroutine
			for _, group := range fullObj.Metadata.Parents {
				for _, parent := range group {
					objectHashes <- parent
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
	hash tilde.Digest,
	pr *peer.ConnectionInfo,
) (*object.Object, error) {
	objCh := make(chan *object.Object)
	errCh := make(chan error)

	rID := m.newRequestID()

	sub := m.mesh.Subscribe(
		mesh.FilterByObjectType(object.ResponseType),
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
			if err := object.Unmarshal(e.Payload, res); err != nil {
				errCh <- err
				break
			}
			if res.RequestID == rID && res.Object != nil {
				objCh <- res.Object
				break
			}
		}
	}()

	req := &object.Request{
		RequestID:  rID,
		ObjectHash: hash,
	}
	ro, err := object.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err := m.mesh.Send(
		ctx,
		ro,
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
	sub mesh.EnvelopeSubscription,
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
				log.String("payload.hash", env.Payload.Hash().String()),
			)

		logger.Debug("handling object")

		switch env.Payload.Type {
		case object.RequestType:
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
		case stream.RequestType:
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
		case stream.SubscriptionType:
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
		case stream.AnnouncementType:
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
	if skip := m.isWellKnownEphemeral(obj.Type); skip {
		return nil
	}

	logger := log.FromContext(ctx)
	objType := obj.Type
	objHash := obj.Hash()

	// store object
	if err := m.objectstore.Put(obj); err != nil {
		logger.Error(
			"error trying to persist incoming object",
			log.String("hash", objHash.String()),
			log.String("type", objType),
			log.Error(err),
		)
		// TODO if we failed to store why are we not returning?
	}

	return nil
}

func (m *manager) announceStreamChildren(
	ctx context.Context,
	streamHash tilde.Digest,
	children []tilde.Digest,
) {
	logger := log.FromContext(ctx)

	// find ephemeral subscriptions for this stream
	// TODO do we really need ephemeral subscriptions?
	subscribersMap := map[string]struct{}{}
	m.subscriptions.Range(func(_ tilde.Digest, sub *stream.Subscription) bool {
		// TODO check expiry
		subscribersMap[sub.Metadata.Owner.String()] = struct{}{}
		return true
	})

	// find subscriptions that are attached in the stream
	r, err := m.objectstore.GetByStream(streamHash)
	if err != nil {
		return
	}

	for {
		obj, err := r.Read()
		// TODO do we want to return if error is not EOF?
		if err != nil {
			break
		}
		if obj.Type != stream.SubscriptionType {
			continue
		}
		if obj.Metadata.Owner.IsEmpty() {
			continue
		}
		subscribersMap[obj.Metadata.Owner.String()] = struct{}{}
	}

	// remove self
	delete(subscribersMap, m.mesh.GetPeerKey().PublicKey().String())

	subscribers := []string{}
	for subscriber := range subscribersMap {
		subscribers = append(subscribers, subscriber)
	}

	logger.Info("trying to announce",
		log.String("streamHash", streamHash.String()),
		log.Any("subscribers", subscribers),
	)

	if len(subscribers) == 0 {
		return
	}

	// notify subscribers
	announcement := &stream.Announcement{
		Metadata: object.Metadata{
			Owner: m.mesh.GetPeerKey().PublicKey().DID(),
		},
		StreamHash:   streamHash,
		ObjectHashes: children,
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
		ao, err := object.Marshal(announcement)
		if err != nil {
			logger.Info(
				"error marshaling announcement",
				log.Error(err),
				log.String("subscriber", subscriber),
			)
			continue
		}
		err = m.mesh.Send(ctx, ao, k)
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
	env *mesh.Envelope,
) error {
	logger := log.FromContext(ctx).With(
		log.String("method", "objectmanager.handleObjectRequest"),
	)

	req := &object.Request{}
	if err := object.Unmarshal(env.Payload, req); err != nil {
		logger.Warn(
			"received invalid object request",
			log.Error(err),
		)
		return err
	}

	logger = logger.With(
		log.String("req.objectID", req.ObjectHash.String()),
	)

	logger.Info("handling object request")

	resp := &object.Response{
		Metadata: object.Metadata{
			Owner: m.mesh.GetPeerKey().PublicKey().DID(),
		},
		Object:    nil,
		RequestID: req.RequestID,
	}

	obj, err := m.objectstore.Get(req.ObjectHash)
	if err != nil {
		logger.Error(
			"error getting object to respond with",
			log.Error(err),
		)
		if err != objectstore.ErrNotFound {
			return err
		}
		ro, err := object.Marshal(resp)
		if err != nil {
			return err
		}
		if sErr := m.mesh.Send(
			ctx,
			ro,
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

	ro, err := object.Marshal(resp)
	if err != nil {
		return err
	}
	err = m.mesh.Send(
		ctx,
		ro,
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
	env *mesh.Envelope,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.FromContext(ctx).With(
		log.String("method", "objectmanager.handleStreamRequest"),
	)

	req := &stream.Request{}
	if err := object.Unmarshal(env.Payload, req); err != nil {
		return err
	}

	// start response
	res := &stream.Response{
		Metadata: object.Metadata{
			Owner: m.mesh.GetPeerKey().PublicKey().DID(),
		},
		RequestID: req.RequestID,
		RootHash:  req.RootHash,
	}

	leaves, err := m.objectstore.GetStreamLeaves(res.RootHash)
	if err != nil && !errors.Is(err, objectstore.ErrNotFound) {
		return err
	}

	res.Leaves = leaves

	ro, err := object.Marshal(res)
	if err != nil {
		return err
	}
	if err := m.mesh.Send(
		ctx,
		ro,
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
	env *mesh.Envelope,
) error {
	sub := &stream.Subscription{}
	if err := object.Unmarshal(env.Payload, sub); err != nil {
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
	env *mesh.Envelope,
) error {
	ann := &stream.Announcement{}
	if err := object.Unmarshal(env.Payload, ann); err != nil {
		return err
	}

	// TODO check if this a stream we care about using ann.StreamHash

	logger := log.FromContext(ctx).With(
		log.String("method", "objectmanager.handleStreamAnnouncement"),
		log.Any("sender", env.Sender),
	)

	logger.Info("got stream announcement ",
		log.Any("hashes", ann.ObjectHashes),
	)

	// check if we already know about these objects
	allKnown := true
	for _, hash := range ann.ObjectHashes {
		_, err := m.objectstore.Get(hash)
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
		ann.ObjectHashes,
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
		ann.StreamHash,
		ann.ObjectHashes,
	)

	return nil
}

// Put stores a given object as-is, and announces it to any subscribers.
func (m *manager) Put(
	ctx context.Context,
	o *object.Object,
) error {
	// add to store
	if err := m.storeObject(ctx, o); err != nil {
		return err
	}
	// publish to pubsub
	m.pubsub.Publish(o)
	return nil
}

// Append stores a given object that is part of a stream.
// If the object's parents and sequence are missing they will be added.
// TODO should we check if the given parents and sequence are valid?
// TODO should we check if the stream is valid?
func (m *manager) Append(
	ctx context.Context,
	o *object.Object,
) (*object.Object, error) {
	root := o.Metadata.Root

	// check and add parents
	if o.Metadata.Root.IsEmpty() {
		return nil, ErrMissingRoot
	}
	if len(o.Metadata.Parents) == 0 {
		leaves, err := m.objectstore.GetStreamLeaves(root)
		if err != nil {
			return nil, err
		}
		if len(leaves) == 0 {
			leaves = []tilde.Digest{
				root,
			}
		}
		tilde.SortDigests(leaves)
		o.Metadata.Parents = object.Parents{
			"*": leaves,
		}
	}

	// check and add sequence
	// if a sequence has not been set, we find the highest sequence of the
	// direct parents and add one to it
	if o.Metadata.Sequence == 0 {
		for _, ph := range o.Metadata.Parents["*"] {
			p, err := m.objectstore.Get(ph)
			if err != nil {
				return nil, fmt.Errorf("missing parent, %w", err)
			}
			if p.Metadata.Sequence > o.Metadata.Sequence {
				o.Metadata.Sequence = p.Metadata.Sequence
			}
		}
		o.Metadata.Sequence++
	}

	// add to store
	if err := m.storeObject(ctx, o); err != nil {
		return nil, err
	}

	// announce to subscribers
	m.announceStreamChildren(
		context.New(
			context.WithCorrelationID(ctx.CorrelationID()),
			// TODO timeout?
		),
		root,
		[]tilde.Digest{
			o.Hash(),
		},
	)

	// publish to pubsub
	m.pubsub.Publish(o)
	return o, nil
}

func (m *manager) AddStreamSubscription(
	ctx context.Context,
	rootHash tilde.Digest,
) error {
	r, err := m.objectstore.GetByStream(rootHash)
	if err != nil {
		return fmt.Errorf("error trying to get stream objects, %w", err)
	}

	pub := m.mesh.GetPeerKey().PublicKey()

	for {
		o, err := r.Read()
		if errors.Is(err, object.ErrReaderDone) {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading stream objects, %w", err)
		}

		if o.Type != stream.SubscriptionType {
			continue
		}

		s := &stream.Subscription{}
		if err := object.Unmarshal(o, s); err != nil {
			continue
		}

		if !s.Metadata.Owner.Equals(pub.DID()) {
			continue
		}

		// TODO check if the subscription has expired

		// already subscribed

		return nil
	}

	s := &stream.Subscription{
		Metadata: object.Metadata{
			Owner: pub.DID(),
			Root:  rootHash,
		},
		RootHashes: []tilde.Digest{
			rootHash,
		},
		// TODO add expiry
	}

	so, err := object.Marshal(s)
	if err != nil {
		return err
	}

	if err := m.Put(ctx, so); err != nil {
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
