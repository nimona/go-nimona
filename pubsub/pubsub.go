package pubsub

import (
	"regexp"
	"sync"

	"github.com/nimona/go-nimona/net"
)

// type Registry struct {
// 	subscriptions map[string]chan interface{}
// }

// func (r *Registry) Subscribe(peerID string, topics ...string) error {
// 	if ch, ok :=r.subscriptions[peerID]; ok {

// 	}
// }

type PubSub interface {
	Publish(msg interface{}, topics ...string) error
	Subscribe(topics ...string) (chan interface{}, error)
	Unsubscribe(chan interface{}) error
}

type pubSub struct {
	sync.RWMutex
	subscriptionTopicMatches map[string]*regexp.Regexp
	subscriptions            map[string]map[chan interface{}]bool
	channelSize              int

	net net.Net
	// streams []net.Conn
}

func NewPubSub(nn net.Net) (PubSub, error) {
	ps := &pubSub{
		subscriptionTopicMatches: map[string]*regexp.Regexp{},
		subscriptions:            map[string]map[chan interface{}]bool{},
		channelSize:              100,
		net:                      nn,
	}

	// go func() {
	// 	ch, _ := ps.Subscribe("(.*)")
	// 	for {
	// 		msg := <-ch

	// 	}
	// }()

	return ps, nil
}

func (ps *pubSub) Publish(msg interface{}, topics ...string) error {
	ps.RLock()
	defer ps.RUnlock()

	for subscribedTopic, subscriptionChs := range ps.subscriptions {
		topicMatch := ps.subscriptionTopicMatches[subscribedTopic]
		for _, topic := range topics {
			if !topicMatch.MatchString(topic) {
				continue
			}
			for subscriptionCh, ok := range subscriptionChs {
				if ok {
					subscriptionCh <- msg
				}
			}
		}
	}
	return nil
}

func (ps *pubSub) Subscribe(topics ...string) (chan interface{}, error) {
	ps.Lock()
	defer ps.Unlock()

	ch := make(chan interface{}, ps.channelSize)
	for _, topic := range topics {
		topicMatch, err := regexp.Compile(topic)
		if err != nil {
			return nil, err
		}
		ps.subscriptionTopicMatches[topic] = topicMatch
		subscriptions, ok := ps.subscriptions[topic]
		if !ok {
			ps.subscriptions[topic] = map[chan interface{}]bool{}
			subscriptions = ps.subscriptions[topic]
		}
		subscriptions[ch] = true
	}
	return ch, nil
}

func (ps *pubSub) Unsubscribe(ch chan interface{}) error {
	ps.Lock()
	defer ps.Unlock()

	for _, subscriptions := range ps.subscriptions {
		if _, ok := subscriptions[ch]; !ok {
			continue
		}
		close(ch)
		delete(subscriptions, ch)
	}
	return nil
}
