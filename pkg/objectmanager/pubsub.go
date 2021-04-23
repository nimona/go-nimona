package objectmanager

import (
	"nimona.io/internal/pubsub"
	"nimona.io/pkg/object"
)

type (
	// ObjectPubSub -
	ObjectPubSub interface {
		Publish(*object.Object)
		Subscribe(...ObjectFilter) ObjectSubscription
	}
	ObjectFilter func(*object.Object) bool
	// ObjectSubscription is returned for every subscription
	ObjectSubscription interface {
		object.ReadCloser
		Channel() <-chan *object.Object
	}
	objectSubscription struct {
		subscription pubsub.Subscription
	}
	objectPubSub struct {
		pubsub pubsub.PubSub
	}
)

// NewObject constructs and returns a new Object
func NewObjectPubSub() ObjectPubSub {
	return &objectPubSub{
		pubsub: pubsub.New(),
	}
}

// Cancel the subscription
func (s *objectSubscription) Close() {
	s.subscription.Cancel()
}

// Channel returns a channel that will be returning the items from the queue
func (s *objectSubscription) Channel() <-chan *object.Object {
	c := s.subscription.Channel()
	r := make(chan *object.Object)
	go func() {
		for {
			v, ok := <-c
			if !ok {
				close(r)
				return
			}
			r <- v.(*object.Object)
		}
	}()
	return r
}

// Next returns the next item from the queue
func (s *objectSubscription) Read() (r *object.Object, err error) {
	next, err := s.subscription.Next()
	if err != nil {
		return
	}
	return next.(*object.Object), nil
}

// Subscribe to published events with optional filters
func (ps *objectPubSub) Subscribe(filters ...ObjectFilter) ObjectSubscription {
	// cast filters
	iFilters := make([]pubsub.Filter, len(filters))
	for i, filter := range filters {
		filter := filter
		iFilters[i] = func(v interface{}) bool {
			return filter(v.(*object.Object))
		}
	}
	// create a new subscription
	sub := &objectSubscription{
		subscription: ps.pubsub.Subscribe(iFilters...),
	}

	return sub
}

// Publish to all subscribers
func (ps *objectPubSub) Publish(v *object.Object) {
	ps.pubsub.Publish(v)
}
