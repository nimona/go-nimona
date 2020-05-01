package eventbus

import (
	"sync"

	"nimona.io/internal/pubsub"
)

var (
	// DefaultEventbus -
	DefaultEventbus = New()
)

type (
	localEvent interface {
		isLocalEvent()
	}
	// Eventbus for passing information around the various nimona packages.
	Eventbus interface {
		Publish(localEvent)
		Subscribe() pubsub.Subscription
	}
	// eventbus -
	eventbus struct {
		lock    sync.RWMutex
		history []localEvent
		pubsub  pubsub.PubSub
	}
)

// New constructs a new event bus
func New() Eventbus {
	return &eventbus{
		lock:    sync.RWMutex{},
		history: []localEvent{},
		pubsub:  pubsub.New(),
	}
}

// Publish a local event
func (eb *eventbus) Publish(e localEvent) {
	eb.lock.Lock()
	eb.pubsub.Publish(e)
	eb.history = append(eb.history, e)
	eb.lock.Unlock()
}

// Subscribe to all local events
func (eb *eventbus) Subscribe() pubsub.Subscription {
	eb.lock.Lock()
	sub := eb.pubsub.Subscribe().(*pubsub.QueueSubscription)
	for _, e := range eb.history {
		sub.Queue.Append(e)
	}
	eb.lock.Unlock()
	return sub
}
