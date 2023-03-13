package tilde

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSON_Scanner(t *testing.T) {
	testStream := "{\"foo\":\"bar\"}\n{\"bar\":\"baz\"}\n"

	expected := []Map{{
		"foo": String("bar"),
	}, {
		"bar": String("baz"),
	}}

	t.Run("Scan", func(t *testing.T) {
		sc := NewScanner(strings.NewReader(testStream))
		var got []Map
		for {
			m, err := sc.Scan()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			got = append(got, m)
		}
		require.Equal(t, expected, got)
	})
}
