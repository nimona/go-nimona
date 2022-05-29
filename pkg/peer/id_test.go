package peer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestID_MarshalString(t *testing.T) {
	tests := []struct {
		id               string
		wantMarshalErr   bool
		wantUnmarshalErr bool
	}{{
		id: "nimona:peer:foo",
	}, {
		id: "nimona:keystream:foo",
	}, {
		id:               "nimona:foo:foo",
		wantUnmarshalErr: true,
	}, {
		id:               "foo:bar:baz",
		wantUnmarshalErr: true,
	}, {
		id:               "baz",
		wantUnmarshalErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			id, err := NewID(tt.id)
			if (err != nil) != tt.wantUnmarshalErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantUnmarshalErr)
				return
			}
			if id != nil {
				got, err := id.MarshalString()
				if (err != nil) != tt.wantMarshalErr {
					t.Errorf("error = %v, wantErr %v", err, tt.wantMarshalErr)
					return
				}
				if !tt.wantMarshalErr && !tt.wantUnmarshalErr {
					require.Equal(t, tt.id, got)
				}
			}
		})
	}
}

func TestID_Equal(t *testing.T) {
	d1 := MustNewID("nimona:peer:foo")
	d2 := MustNewID("nimona:peer:foo")
	require.True(t, d1.Equals(*d2))
	require.True(t, *d1 == *d2)
}
