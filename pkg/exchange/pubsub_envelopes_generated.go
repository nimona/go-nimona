// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/geoah/genny

package exchange

import (
	"nimona.io/internal/pubsub"
)

type (
	// EnvelopePubSub -
	EnvelopePubSub interface {
		Publish(*Envelope)
		Subscribe(...EnvelopeFilter) EnvelopeSubscription
	}
	EnvelopeFilter func(*Envelope) bool
	// EnvelopeSubscription is returned for every subscription
	EnvelopeSubscription interface {
		Next() (*Envelope, error)
		Cancel()
	}
	envelopeSubscription struct {
		subscription pubsub.Subscription
	}
	envelopePubSub struct {
		pubsub pubsub.PubSub
	}
)

// NewEnvelope constructs and returns a new Envelope
func NewEnvelopePubSub() EnvelopePubSub {
	return &envelopePubSub{
		pubsub: pubsub.New(),
	}
}

// Cancel the subscription
func (s *envelopeSubscription) Cancel() {
	s.subscription.Cancel()
}

// Next returns the an item from the queue
func (s *envelopeSubscription) Next() (r *Envelope, err error) {
	next, err := s.subscription.Next()
	if err != nil {
		return
	}
	return next.(*Envelope), nil
}

// Subscribe to published events with optional filters
func (ps *envelopePubSub) Subscribe(filters ...EnvelopeFilter) EnvelopeSubscription {
	// cast filters
	iFilters := make([]pubsub.Filter, len(filters))
	for i, filter := range filters {
		filter := filter
		iFilters[i] = func(v interface{}) bool {
			return filter(v.(*Envelope))
		}
	}
	// create a new subscription
	sub := &envelopeSubscription{
		subscription: ps.pubsub.Subscribe(iFilters...),
	}

	return sub
}

// Publish to all subscribers
func (ps *envelopePubSub) Publish(v *Envelope) {
	ps.pubsub.Publish(v)
}
