package publisher

import (
	"github.com/cheekybits/genny/generic"
)

type (
	ObservableType generic.Type // nolint
	// PubSub -
	PubSub interface {
		Publisher
		Subscriber
	}
	// Publisher deals with the publishing part of our PubSub
	Publisher interface {
		Publish(ObservableType)
	}
	// Subscriber deals with the subscribing part of our PubSub
	Subscriber interface {
		Subscribe(chan ObservableType, ...filter)
		Unsubscribe(chan ObservableType)
	}
	filter     func(ObservableType) bool
	subscriber struct {
		filters []filter
		out     chan ObservableType
	}
	publisher struct {
		outgoing             chan ObservableType
		registerSubscriber   chan *subscriber
		unregisterSubscriber chan chan ObservableType
		subscriptions        map[*subscriber]bool
	}
)

// NewPubSub constructs and returns a new PubSub
func NewPubSub() PubSub {
	o := &publisher{
		outgoing:             make(chan ObservableType, 100),
		registerSubscriber:   make(chan *subscriber, 1),
		unregisterSubscriber: make(chan chan ObservableType, 1),
		subscriptions:        map[*subscriber]bool{},
	}

	go o.process()

	return o
}

// Subscribe to published events with optional filters
func (o *publisher) Subscribe(out chan ObservableType, filters ...filter) {
	o.registerSubscriber <- &subscriber{
		out:     out,
		filters: filters,
	}
}

// Unsubscribe given a subscribed channel
func (o *publisher) Unsubscribe(out chan ObservableType) {
	o.unregisterSubscriber <- out
}

// Publish to all subscribers
func (o *publisher) Publish(v ObservableType) {
	o.outgoing <- v
}

func (o *publisher) process() {
	for {
		select {
		case s := <-o.registerSubscriber:
			o.subscriptions[s] = true

		case s := <-o.unregisterSubscriber:
			for k := range o.subscriptions {
				if k.out == s {
					delete(o.subscriptions, k)
				}
			}

		case e := <-o.outgoing:
			for s := range o.subscriptions {
				publish := true
				for _, f := range s.filters {
					if f(e) == false {
						publish = false
						break
					}
				}
				if publish {
					s.out <- e
				}
			}
		}
	}
}
