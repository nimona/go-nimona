package tilde

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFluent(t *testing.T) {
	exp := Map{
		"foo": String("baz"),
		"bar": List{
			String("baz"),
			String("qux"),
		},
	}

	ori := Map{
		"foo": String("bar"),
	}
	ori.Fluent().
		Set("foo", String("baz")).
		Set("bar", List{
			String("baz"),
		}).
		Append("bar", String("qux"))

	require.Equal(t, exp, ori)

	foundString := ori.Fluent().Get("foo").String()
	require.Equal(t, String("baz"), foundString)

	missingMap := ori.Fluent().
		Get("nop").
		Map().
		Fluent().
		Get("doubleNop").
		String()
	require.Equal(t, String(""), missingMap)
}
