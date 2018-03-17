package pubsub

import (
	"regexp"
	"sync"
)

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
		subscriptions[ch] = false
	}
	return nil
}
