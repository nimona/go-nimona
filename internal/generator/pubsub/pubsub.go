package pubsub

import (
	"github.com/geoah/genny/generic"

	"nimona.io/internal/pubsub"
)

type (
	ObjectType generic.Type // nolint
	// NamePubSub -
	NamePubSub interface {
		Publish(ObjectType)
		Subscribe(...NameFilter) NameSubscription
	}
	NameFilter func(ObjectType) bool
	// NameSubscription is returned for every subscription
	NameSubscription interface {
		Channel() <-chan ObjectType
		Next() (ObjectType, error)
		Cancel()
	}
	nameSubscription struct {
		subscription pubsub.Subscription
	}
	namePubSub struct {
		pubsub pubsub.PubSub
	}
)

// NewName constructs and returns a new Name
func NewNamePubSub() NamePubSub {
	return &namePubSub{
		pubsub: pubsub.New(),
	}
}

// Cancel the subscription
func (s *nameSubscription) Cancel() {
	s.subscription.Cancel()
}

// Channel returns a channel that will be returning the items from the queue
func (s *nameSubscription) Channel() <-chan ObjectType {
	c := s.subscription.Channel()
	r := make(chan ObjectType)
	go func() {
		for {
			v, ok := <-c
			if !ok {
				close(r)
				return
			}
			r <- v.(ObjectType)
		}
	}()
	return r
}

// Next returns the next item from the queue
func (s *nameSubscription) Next() (r ObjectType, err error) {
	next, err := s.subscription.Next()
	if err != nil {
		return
	}
	return next.(ObjectType), nil
}

// Subscribe to published events with optional filters
func (ps *namePubSub) Subscribe(filters ...NameFilter) NameSubscription {
	// cast filters
	iFilters := make([]pubsub.Filter, len(filters))
	for i, filter := range filters {
		filter := filter
		iFilters[i] = func(v interface{}) bool {
			return filter(v.(ObjectType))
		}
	}
	// create a new subscription
	sub := &nameSubscription{
		subscription: ps.pubsub.Subscribe(iFilters...),
	}

	return sub
}

// Publish to all subscribers
func (ps *namePubSub) Publish(v ObjectType) {
	ps.pubsub.Publish(v)
}
