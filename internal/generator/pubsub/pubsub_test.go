package pubsub

import (
	"strings"
	"testing"

	"nimona.io/internal/pubsub"

	"github.com/stretchr/testify/require"
)

func TestPubSubSimple(t *testing.T) {
	ps := NewNamePubSub()
	ps.Publish("one")

	s := ps.Subscribe()
	ps.Publish("two")
	ps.Publish("three")

	next, err := s.Next()
	require.Equal(t, "two", next)
	require.NoError(t, err)

	next, err = s.Next()
	require.Equal(t, "three", next)
	require.NoError(t, err)
}

func TestPubSubFiltered(t *testing.T) {
	ps := NewNamePubSub()

	f1 := func(v ObjectType) bool {
		return strings.HasPrefix(v.(string), "t")
	}

	s := ps.Subscribe(f1)
	ps.Publish("one")
	ps.Publish("two")
	ps.Publish("three")

	next, err := s.Next()
	require.Equal(t, "two", next)
	require.NoError(t, err)

	next, err = s.Next()
	require.Equal(t, "three", next)
	require.NoError(t, err)
}

func TestPubSubFilteredMultiple(t *testing.T) {
	ps := NewNamePubSub()

	f1 := func(v ObjectType) bool {
		return strings.HasPrefix(v.(string), "t")
	}

	f2 := func(v ObjectType) bool {
		return strings.HasSuffix(v.(string), "e")
	}

	s := ps.Subscribe(f1, f2)
	ps.Publish("one")
	ps.Publish("two")
	ps.Publish("three")

	next, err := s.Next()
	require.Equal(t, "three", next)
	require.NoError(t, err)
}

func TestPubSubCancel(t *testing.T) {
	ps := NewNamePubSub()

	s := ps.Subscribe()
	ps.Publish("one")
	ps.Publish("two")

	next, err := s.Next()
	require.Equal(t, "one", next)
	require.NoError(t, err)

	s.Cancel()

	next, err = s.Next()
	require.Equal(t, "two", next)
	require.NoError(t, err)

	next, err = s.Next()
	require.Equal(t, nil, next)
	require.Error(t, err, pubsub.ErrSubscriptionCanceled)
}
