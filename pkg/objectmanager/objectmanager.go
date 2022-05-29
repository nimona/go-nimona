package objectmanager

import (
	"fmt"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
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

type (
	ObjectManager interface {
		Put(
			ctx context.Context,
			o *object.Object,
		) error
		Request(
			ctx context.Context,
			hash tilde.Digest,
			id peer.ID,
		) (*object.Object, error)
		Subscribe(
			lookupOptions ...LookupOption,
		) ObjectSubscription
	}

	manager struct {
		network       network.Network
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
		objectstore:   str,
	}

	logger := log.
		FromContext(ctx).
		Named("objectmanager").
		With(
			log.String("method", "objectmanager.New"),
		)

	subs := m.network.Subscribe()

	go func() {
		if err := m.handleObjects(ctx, subs); err != nil {
			logger.Error("handling object requests failed", log.Error(err))
		}
	}()

	return m
}

func (m *manager) isWellKnownEphemeral(
	contentType string,
) bool {
	switch contentType {
	case
		stream.AnnouncementType,
		stream.RequestType,
		stream.ResponseType,
		object.RequestType,
		object.ResponseType:
		return true
	}
	return false
}

func (m *manager) Request(
	ctx context.Context,
	hash tilde.Digest,
	pr peer.ID,
) (*object.Object, error) {
	objCh := make(chan *object.Object)
	errCh := make(chan error)

	rID := m.newRequestID()

	sub := m.network.Subscribe(
		network.FilterByRequestID(rID),
	)
	defer sub.Cancel()

	go func() {
		for {
			e, err := sub.Next()
			if err != nil {
				errCh <- err
				return
			}
			if e == nil {
				return
			}
			res := &object.Response{}
			if err := object.Unmarshal(e.Payload, res); err != nil {
				// errCh <- err // TODO not sure about this one
				continue
			}
			objCh <- res.Object
			return
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
	if err := m.network.Send(
		ctx,
		ro,
		pr,
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
	ctx context.Context,
	sub network.EnvelopeSubscription,
) error {
	for {
		env, err := sub.Next()
		if err != nil {
			return err
		}

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

func (m *manager) handleObjectRequest(
	ctx context.Context,
	env *network.Envelope,
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
			Owner: m.network.GetPeerID(),
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
		if sErr := m.network.Send(
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

	resp.Object = object.Copy(obj)

	ro, err := object.Marshal(resp)
	if err != nil {
		return err
	}
	err = m.network.Send(
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

func (m *manager) Subscribe(
	lookupOptions ...LookupOption,
) ObjectSubscription {
	options := newLookupOptions(lookupOptions...)
	return m.pubsub.Subscribe(options.Filters...)
}
