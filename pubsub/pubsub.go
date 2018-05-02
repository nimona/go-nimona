package pubsub

import (
	"regexp"
)

type PubSub interface {
	Publish(msg interface{}, topic string) error
	Subscribe(topic string) (chan interface{}, error)
	Unsubscribe(chan interface{}) error
}

type pubSub struct {
	incMessages              chan *messageWithTopic
	incSubscriptions         chan *subscription
	incUnsubscriptions       chan *subscription
	subscriptionTopicMatches map[string]*regexp.Regexp
	subscriptions            map[string]map[chan interface{}]bool
	channelSize              int
}

type messageWithTopic struct {
	topic   string
	message interface{}
}

type subscription struct {
	topic   *regexp.Regexp
	channel chan interface{}
}

func NewPubSub() (PubSub, error) {
	ps := &pubSub{
		incMessages:              make(chan *messageWithTopic, 1000),
		incSubscriptions:         make(chan *subscription, 10),
		incUnsubscriptions:       make(chan *subscription, 10),
		subscriptionTopicMatches: map[string]*regexp.Regexp{},
		subscriptions:            map[string]map[chan interface{}]bool{},
		channelSize:              100,
	}

	go func() {
		for {
			select {
			case msg := <-ps.incMessages:
				for subscribedTopic, subscriptionChs := range ps.subscriptions {
					topicMatch := ps.subscriptionTopicMatches[subscribedTopic]
					if !topicMatch.MatchString(msg.topic) {
						continue
					}
					for subscriptionCh, ok := range subscriptionChs {
						if ok {
							subscriptionCh <- msg.message
						}
					}
				}
			case sub := <-ps.incSubscriptions:
				topic := sub.topic.String()
				ps.subscriptionTopicMatches[topic] = sub.topic
				subscriptions, ok := ps.subscriptions[topic]
				if !ok {
					ps.subscriptions[topic] = map[chan interface{}]bool{}
					subscriptions = ps.subscriptions[topic]
				}
				subscriptions[sub.channel] = true
			case sub := <-ps.incUnsubscriptions:
				for _, subscriptions := range ps.subscriptions {
					if _, ok := subscriptions[sub.channel]; !ok {
						continue
					}
					close(sub.channel)
					delete(subscriptions, sub.channel)
				}
			}
		}
	}()

	return ps, nil
}

func (ps *pubSub) Publish(msg interface{}, topic string) error {
	msgWithTopic := &messageWithTopic{
		topic:   topic,
		message: msg,
	}
	ps.incMessages <- msgWithTopic
	return nil
}

func (ps *pubSub) Subscribe(topic string) (chan interface{}, error) {
	topicMatch, err := regexp.Compile(topic)
	if err != nil {
		return nil, err
	}
	ch := make(chan interface{}, ps.channelSize)
	sub := &subscription{
		topic:   topicMatch,
		channel: ch,
	}
	ps.incSubscriptions <- sub
	return ch, nil
}

func (ps *pubSub) Unsubscribe(ch chan interface{}) error {
	sub := &subscription{
		channel: ch,
	}
	ps.incUnsubscriptions <- sub
	return nil
}
