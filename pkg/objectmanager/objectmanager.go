package objectmanager

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

var (
	requestType = new(Request).GetType()
	ErrTimeout  = errors.New("request timed out")
)

//go:generate $GOBIN/mockgen -destination=../objectmanagermock/objectmanagermock_generated.go -package=objectmanagermock -source=objectmanager.go

type (
	Requester interface {
		Request(
			ctx context.Context,
			hash object.Hash,
			peer *peer.Peer,
		) (*object.Object, error)
	}
	manager struct {
		store    *sqlobjectstore.Store
		exchange exchange.Exchange
	}
	Option func(*manager)
)

func New(
	ctx context.Context,
	opts ...Option,
) Requester {
	m := &manager{}

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
		exchange.FilterByObjectType(requestType),
	)

	go func() {
		if err := m.handleObjects(subs); err != nil {
			logger.Error("handling object requests failed", log.Error(err))
		}
	}()

	return m
}

func (m *manager) Request(
	ctx context.Context,
	hash object.Hash,
	pr *peer.Peer,
) (*object.Object, error) {
	objCh := make(chan *object.Object)
	errCh := make(chan error)

	sub := m.exchange.Subscribe(
		func(e *exchange.Envelope) bool {
			return e.Payload.Hash() == hash
		},
	)
	defer sub.Cancel()

	go func() {
		for {
			e, err := sub.Next()
			if err != nil {
				errCh <- err
				break
			}

			if e.Payload.Hash() == hash {
				objCh <- &e.Payload
				break
			}
		}
	}()
	if err := m.exchange.Send(ctx, Request{
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
		e, err := sub.Next()
		if err != nil {
			return err
		}

		logger := log.
			FromContext(context.Background()).
			Named("objectmanager").
			With(
				log.String("method", "objectmanager.handleObjects"),
				log.String("payload", e.Payload.GetType()),
			)

		logger.Debug("getting payload")

		if e.Payload.GetType() == requestType {
			req := &Request{}
			if err := req.FromObject(e.Payload); err != nil {
				return err
			}

			hash := req.ObjectHash
			obj, err := m.store.Get(hash)
			if err != nil {
				return err
			}

			if err := m.exchange.Send(
				context.Background(),
				obj,
				&peer.Peer{
					Owners: []crypto.PublicKey{
						e.Sender,
					},
				},
			); err != nil {
				return errors.Wrap(
					errors.Error("could not send object"),
					err,
				)
			}
		}
	}
}
