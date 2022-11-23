package xsync

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueue_E2E(t *testing.T) {
	q := NewQueue[string](10)

	t.Run("push", func(t *testing.T) {
		q.Push("foo")
		q.Push("bar")
		q.Push("baz")
	})

	t.Run("pop", func(t *testing.T) {
		v, err := q.Pop()
		require.NoError(t, err)
		require.Equal(t, "foo", v)

		v, err = q.Pop()
		require.NoError(t, err)
		require.Equal(t, "bar", v)
	})

	t.Run("select", func(t *testing.T) {
		v := <-q.Select()
		require.Equal(t, "baz", v)
	})

	q.Close()

	t.Run("pop closed", func(t *testing.T) {
		v, err := q.Pop()
		require.Zero(t, v)
		require.Error(t, err)
	})

	t.Run("push closed", func(t *testing.T) {
		err := q.Push("qux")
		require.Error(t, err)
	})
}
