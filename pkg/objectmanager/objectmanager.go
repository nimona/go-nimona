package objectmanager

import (
	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/stream"
)

var (
	ErrDone    = errors.New("done")
	ErrTimeout = errors.New("request timed out")
)

var (
	objectRequestType = object.Request{}.GetType()
	streamRequestType = stream.Request{}.GetType()
)

//go:generate $GOBIN/mockgen -destination=../objectmanagermock/objectmanagermock_generated.go -package=objectmanagermock -source=objectmanager.go
//go:generate $GOBIN/mockgen -destination=../objectmanagerpubsubmock/objectmanagerpubsubmock_generated.go -package=objectmanagerpubsubmock -source=pubsub_generated.go
//go:generate $GOBIN/genny -in=$GENERATORS/pubsub/pubsub.go -out=pubsub_generated.go -pkg objectmanager -imp=nimona.io/pkg/object gen "ObjectType=object.Object Name=Object name=object"

type (
	ObjectManager interface {
		Request(
			ctx context.Context,
			hash object.Hash,
			peer *peer.Peer,
		) (*object.Object, error)
		RequestStream(
			ctx context.Context,
			rootHash object.Hash,
			peer *peer.Peer,
		) (object.ReferencesResults, error)
	}
	manager struct {
		store    objectstore.Store
		exchange exchange.Exchange
		keychain keychain.Keychain
		pubsub   ObjectPubSub
		newNonce func() string
	}
	Option        func(*manager)
	StreamResults struct {
		context context.Context
		next    chan objectRefOrErr
	}
	objectRefOrErr struct {
		object *object.Object
		err    error
	}
)

func (r *StreamResults) Next() (*object.Object, error) {
	select {
	case n, ok := <-r.next:
		if !ok {
			return nil, ErrDone
		}
		return n.object, n.err
	case <-r.context.Done():
		return nil, ErrTimeout
	}
}

func New(
	ctx context.Context,
	opts ...Option,
) ObjectManager {
	m := &manager{
		newNonce: func() string {
			return rand.String(16)
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	logger := log.
		FromContext(ctx).
		Named("objectmanager").
		With(
			log.String("method", "objectmanager.New"),
		)

	subs := m.exchange.Subscribe(
		exchange.FilterByObjectType(objectRequestType),
	)

	go func() {
		if err := m.handleObjects(subs); err != nil {
			logger.Error("handling object requests failed", log.Error(err))
		}
	}()

	return m
}

func (m *manager) RequestStream(
	ctx context.Context,
	rootHash object.Hash,
	recipient *peer.Peer,
) (object.ReferencesResults, error) {
	nonce := m.newNonce()
	responses := make(chan stream.Response)

	sub := m.exchange.Subscribe(
		exchange.FilterByObjectType(stream.Response{}.GetType()),
	)
	defer sub.Cancel()

	go func() {
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

	req := stream.Request{
		RootHash: rootHash,
	}
	if err := m.exchange.Send(
		ctx,
		req.ToObject(),
		recipient,
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

	next := make(chan objectRefOrErr)

	requestHandler := func(
		ctx context.Context,
		objectHash object.Hash,
	) (*object.Object, error) {
		// TODO(geoah) check store before requesting the object
		// obj, err := m.store.Get(objectHash)
		// if err == nil {
		// 	return &obj, nil
		// }
		return m.Request(
			ctx,
			objectHash,
			recipient,
		)
	}

	go func() {
		for _, objectHash := range objectHashes {
			fullObj, err := object.LoadReferences(
				ctx,
				objectHash,
				requestHandler,
			)
			// TODO check the validity of each event, they should be ordered
			// so we should already have its parents.
			next <- objectRefOrErr{
				object: fullObj,
				err:    err,
			}
			if err != nil {
				break
			}
		}
		close(next)
	}()

	return &StreamResults{
		context: ctx,
		next:    next,
	}, nil
}

func (m *manager) Request(
	ctx context.Context,
	hash object.Hash,
	pr *peer.Peer,
) (*object.Object, error) {
	objCh := make(chan *object.Object)
	errCh := make(chan error)

	sub := m.exchange.Subscribe(
		// exchange.FilterByObjectType(object.Response{}.GetType()),
		exchange.FilterByObjectHash(hash),
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

	if err := m.exchange.Send(ctx, object.Request{
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
	sub exchange.EnvelopeSubscription,
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

		logger.Debug("getting payload")

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
		}
	}
}

func (m *manager) handleObjectRequest(
	ctx context.Context,
	env *exchange.Envelope,
) error {
	req := &object.Request{}
	if err := req.FromObject(env.Payload); err != nil {
		return err
	}

	hash := req.ObjectHash
	obj, err := m.store.Get(hash)
	if err != nil {
		return err
	}

	if err := m.exchange.Send(
		ctx,
		obj,
		&peer.Peer{
			Owners: []crypto.PublicKey{
				env.Sender,
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
	env *exchange.Envelope,
) error {
	// TODO check if policy allows requested to retrieve the object
	logger := log.FromContext(ctx)

	req := &stream.Request{}
	if err := req.FromObject(env.Payload); err != nil {
		return err
	}

	// get the entire graph for this stream
	vs, err := m.store.GetByStream(req.Stream)
	if err != nil {
		return err
	}

	// get only the object hashes
	hs := []object.Hash{}
	for _, o := range vs {
		hs = append(hs, o.Hash())
	}

	res := &stream.Response{
		Stream:   req.Stream,
		Nonce:    req.Nonce,
		Children: hs,
	}

	if err := m.exchange.Send(
		ctx,
		res.ToObject(),
		&peer.Peer{
			Owners: []crypto.PublicKey{
				env.Sender,
			},
		},
	); err != nil {
		logger.Warn(
			"streammanager.handleStreamRequest could not send response",
			log.Error(err),
		)
		return err
	}

	return nil
}

// Put stores a given object
// TODO(geoah) what happened if the stream graph is not complete? Do we care?
func (m *manager) Put(o object.Object) (object.Object, error) {
	// add owners
	// TODO should we be adding owners?
	o = o.SetOwners(
		m.keychain.ListPublicKeys(keychain.IdentityKey),
	)
	// figure out if we need to add parents to the object
	streamHash := o.GetStream()
	if !streamHash.IsEmpty() {
		os, err := m.store.GetByStream(streamHash)
		if err != nil && err != objectstore.ErrNotFound {
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
	if err := m.store.Put(o); err != nil {
		return o, err
	}
	// announce to pubsub
	m.pubsub.Publish(o)
	return o, nil
}

func (m *manager) Subscribe(
	lookupOptions ...LookupOption,
) ObjectSubscription {
	options := newLookupOptions(lookupOptions...)
	return m.pubsub.Subscribe(options.Filters...)
}
