package object

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/tilde"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name   string
		source *Object
		want   *Object
	}{{
		name: "same hash, different ptr",
		source: &Object{
			Type: "foo",
			Data: tilde.Map{
				"foo": tilde.String("bar"),
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Copy(tt.source)
			assert.Equal(t, tt.source.Hash(), got.Hash())
			assert.NotSame(t, tt.source, got)
			assert.NotSame(t, tt.source.Data, got.Data)
		})
	}
}
