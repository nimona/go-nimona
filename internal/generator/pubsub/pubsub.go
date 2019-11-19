package pubsub

import (
	"github.com/cheekybits/genny/generic"

	"nimona.io/internal/pubsub"
)

type (
	ObjectType generic.Type // nolint
	PubSubName string       // nolint
	// PubSubNamePubSub -
	PubSubNamePubSub interface {
		Publish(ObjectType)
		Subscribe(...PubSubNameFilter) PubSubNameSubscription
	}
	PubSubNameFilter func(ObjectType) bool
	// PubSubNameSubscription is returned for every subscription
	PubSubNameSubscription interface {
		Next() (ObjectType, error)
		Cancel()
	}
	psPubSubNameSubscription struct {
		subscription pubsub.Subscription
	}
	psPubSubName struct {
		pubsub pubsub.PubSub
	}
)

// NewPubSubName constructs and returns a new PubSubNamePubSub
func NewPubSubNamePubSub() PubSubNamePubSub {
	return &psPubSubName{
		pubsub: pubsub.New(),
	}
}

// Cancel the subscription
func (s *psPubSubNameSubscription) Cancel() {
	s.subscription.Cancel()
}

// Next returns the an item from the queue
func (s *psPubSubNameSubscription) Next() (ObjectType, error) {
	next, err := s.subscription.Next()
	if err != nil {
		return nil, err
	}
	return next.(ObjectType), nil
}

// Subscribe to published events with optional filters
func (ps *psPubSubName) Subscribe(filters ...PubSubNameFilter) PubSubNameSubscription {
	// cast filters
	iFilters := make([]pubsub.Filter, len(filters))
	for i, filter := range filters {
		filter := filter
		iFilters[i] = func(v interface{}) bool {
			return filter(v.(ObjectType))
		}
	}
	// create a new subscription
	sub := &psPubSubNameSubscription{
		subscription: ps.pubsub.Subscribe(iFilters...),
	}

	return sub
}

// Publish to all subscribers
func (ps *psPubSubName) Publish(v ObjectType) {
	ps.pubsub.Publish(v)
}
