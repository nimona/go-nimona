package object

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/object/value"
)

func TestCopy(t *testing.T) {
	tests := []struct {
		name   string
		source *Object
		want   *Object
	}{{
		name: "same cid, different ptr",
		source: &Object{
			Type: "foo",
			Data: value.Map{
				"foo": value.String("bar"),
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Copy(tt.source)
			assert.Equal(t, tt.source.CID(), got.CID())
			assert.NotSame(t, tt.source, got)
			assert.NotSame(t, tt.source.Data, got.Data)
		})
	}
}
