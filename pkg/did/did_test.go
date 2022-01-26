package did

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDID_MarshalString(t *testing.T) {
	tests := []struct {
		did              string
		wantMarshalErr   bool
		wantUnmarshalErr bool
	}{{
		did: "did:nimona:peer:foo",
	}, {
		did: "did:nimona:keystream:foo",
	}, {
		did:              "did:nimona:foo:foo",
		wantUnmarshalErr: true,
	}, {
		did:              "foo:bar:baz",
		wantUnmarshalErr: true,
	}, {
		did:              "did:baz",
		wantUnmarshalErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.did, func(t *testing.T) {
			did, err := Parse(tt.did)
			if (err != nil) != tt.wantUnmarshalErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantUnmarshalErr)
				return
			}
			if did != nil {
				got, err := did.MarshalString()
				if (err != nil) != tt.wantMarshalErr {
					t.Errorf("error = %v, wantErr %v", err, tt.wantMarshalErr)
					return
				}
				if !tt.wantMarshalErr && !tt.wantUnmarshalErr {
					require.Equal(t, tt.did, got)
				}
			}
		})
	}
}

func TestDID_Equal(t *testing.T) {
	d1 := MustParse("did:nimona:peer:foo")
	d2 := MustParse("did:nimona:peer:foo")
	require.True(t, d1.Equals(*d2))
	require.True(t, *d1 == *d2)
}
