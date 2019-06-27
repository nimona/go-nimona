package subscription

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
)

func TestSubscription_ToObject(t *testing.T) {
	es := &Subscription{
		Subscriber: crypto.Fingerprint("foo"),
		Parents:    []string{"a", "b"},
	}

	o := es.ToObject()
	s := &Subscription{}
	err := s.FromObject(o)
	require.NoError(t, err)
	require.Equal(t, es.Subscriber, s.Subscriber)
	require.Equal(t, es.Parents, s.Parents)
}
