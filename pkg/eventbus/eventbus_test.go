package eventbus

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/pubsub"
)

func Test_eventbus_Subscribe(t *testing.T) {
	eb := New()

	e1 := NetworkAddressAdded{Address: "a"}
	e2 := NetworkAddressAdded{Address: "b"}

	eb.Publish(e1)
	s1 := eb.Subscribe()
	require.Equal(t, []localEvent{e1}, drainSubscription(s1, 1))

	s2 := eb.Subscribe()
	eb.Publish(e2)
	require.Equal(t, []localEvent{e2}, drainSubscription(s1, 1))
	require.Equal(t, []localEvent{e1, e2}, drainSubscription(s2, 2))
}

func drainSubscription(s pubsub.Subscription, expectedCount int) []localEvent {
	es := []localEvent{}
	for i := 0; i < expectedCount; i++ {
		e, _ := s.Next()
		es = append(es, e.(localEvent))
	}
	return es
}
