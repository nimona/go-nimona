package subscription

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/orchestrator"
)

//go:generate $GOBIN/objectify -schema /graph.subscription -type Subscription -in subscription.go -out subscription_generated.go

// Subscription provides a way for users and peers to subscribe to graph updates
type Subscription struct {
	Subscriber crypto.Fingerprint `json:"subscriber:s"`
	Root       string             `json:"@root:s"`
	Parents    []string           `json:"@parents:as"`
}

// New construct a subscription
func New(subscriber *crypto.PublicKey, parents []string) *Subscription {
	return &Subscription{
		Subscriber: subscriber.Fingerprint(),
		Parents:    parents,
	}
}

func GetSubscribers(g *orchestrator.Graph) ([]crypto.Fingerprint, error) {
	sm := map[crypto.Fingerprint]bool{}
	for _, o := range g.Objects {
		if o.GetType() == SubscriptionType {
			s := &Subscription{}
			if err := s.FromObject(o); err != nil {
				// TODO log error
				continue
			}
			sm[s.Subscriber] = true
		}
	}
	sa := make([]crypto.Fingerprint, 0, len(sm))
	for s := range sm {
		sa = append(sa, s)
	}
	return sa, nil
}
