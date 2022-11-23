package xsync

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap_E2E(t *testing.T) {
	m := NewMap[string, string]()

	t.Run("store", func(t *testing.T) {
		m.Store("foo", "bar")
		m.Store("baz", "qux")
	})

	t.Run("range", func(t *testing.T) {
		vs := []string{}
		m.Range(func(key string, value string) bool {
			vs = append(vs, key, value)
			return true
		})
		require.ElementsMatch(t, []string{"foo", "bar", "baz", "qux"}, vs)
	})

	t.Run("load", func(t *testing.T) {
		v, ok := m.Load("foo")
		require.True(t, ok)
		require.Equal(t, "bar", v)

		v, ok = m.Load("baz")
		require.True(t, ok)
		require.Equal(t, "qux", v)
	})

	t.Run("load invalid", func(t *testing.T) {
		v, ok := m.Load("qux")
		require.False(t, ok)
		require.Equal(t, "", v)
	})

	t.Run("load or store existing", func(t *testing.T) {
		v, ok := m.LoadOrStore("foo", "qux")
		require.True(t, ok)
		require.Equal(t, "bar", v)
	})

	t.Run("load or store missing", func(t *testing.T) {
		v, ok := m.LoadOrStore("qux", "quux")
		require.False(t, ok)
		require.Equal(t, "quux", v)
	})

	t.Run("delete", func(t *testing.T) {
		m.Delete("foo")
	})

	t.Run("load deleted", func(t *testing.T) {
		v, ok := m.Load("foo")
		require.False(t, ok)
		require.Equal(t, "", v)
	})
}
