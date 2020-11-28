package pubsub

import (
	"sync"

	"github.com/geoah/go-queue"

	"nimona.io/pkg/errors"
)

const (
	ErrSubscriptionCanceled = errors.Error("subscription canceled")
)

type (
	PubSub interface {
		Publisher
		Subscriber
	}
	Filter func(interface{}) bool
	// Subscription is returned for every subscription
	Subscription interface {
		Channel() <-chan interface{}
		Next() (interface{}, error)
		Cancel()
	}
	QueueSubscription struct {
		filters []Filter
		Queue   *queue.Queue
		cancel  func()
	}
	// Publisher deals with the publishing part of our PubSub
	Publisher interface {
		Publish(interface{})
	}
	// Subscriber deals with the subscribing part of our PubSub
	Subscriber interface {
		Subscribe(...Filter) Subscription
	}
	pubsub struct {
		subscriptions *sync.Map
	}
)

// New constructs and returns a new PubSub
func New() PubSub {
	return &pubsub{
		subscriptions: &sync.Map{},
	}
}

func (ps *QueueSubscription) Cancel() {
	ps.Queue.Append(nil)
	ps.cancel()
}

func (ps *QueueSubscription) Channel() <-chan interface{} {
	r := make(chan interface{})
	go func() {
		for {
			next := ps.Queue.Pop()
			if next == nil {
				close(r)
				return
			}
			r <- next
		}
	}()
	return r
}

func (ps *QueueSubscription) Next() (interface{}, error) {
	// get the first item in the queue
	next := ps.Queue.Pop()
	if next == nil {
		// nil should only be allowed from the cancelation function, so
		// assume it's canceled
		return nil, ErrSubscriptionCanceled
	}
	// once we get an non-nil item, return it
	return next, nil
}

// Subscribe to published events with optional filters
func (ps *pubsub) Subscribe(filters ...Filter) Subscription {
	// create a new subscription
	sub := &QueueSubscription{
		Queue:   queue.New(),
		filters: filters,
	}

	// specify the cancelation function
	sub.cancel = func() {
		// delete the subscription
		ps.subscriptions.Delete(sub)
		// add a nil item in the queue that will cause Next() to error
		// with ErrSubscriptionCanceled
		sub.Queue.Append(nil)
	}

	// and store it
	ps.subscriptions.Store(sub, true)
	return sub
}

// Publish to all subscribers
func (ps *pubsub) Publish(v interface{}) {
	// go through our subscriptions
	ps.subscriptions.Range(func(k, _ interface{}) bool {
		// assuming it's ok
		sub := k.(*QueueSubscription)
		// go through the filters
		for _, filter := range sub.filters {
			// and make sure they pass
			if !filter(v) {
				return true
			}
		}
		// and add it to the queue
		sub.Queue.Append(v)
		return true
	})
}
