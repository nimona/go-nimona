package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityAlias(t *testing.T) {
	s0 := "nimona://id:alias:testing.romdo.io/geoah"
	n0 := &IdentityAlias{
		Network: NetworkAlias{
			Hostname: "testing.romdo.io",
		},
		Handle: "geoah",
	}

	require.Equal(t, s0, n0.String())

	n1, err := ParseIdentityAlias(s0)
	require.NoError(t, err)
	require.Equal(t, n0, n1)

	t.Run("marshal unmarshal", func(t *testing.T) {
		b, err := n0.MarshalCBORBytes()
		require.NoError(t, err)

		n1 := &IdentityAlias{}
		err = n1.UnmarshalCBORBytes(b)
		require.NoError(t, err)
		require.EqualValues(t, n0, n1)
		require.Equal(t, s0, n1.String())
	})
}

// func TestIdentityIdentifier(t *testing.T) {
// 	networkInfoRootID := NewTestRandomDocumentID(t)
// 	tests := []struct {
// 		name    string
// 		cborer  Cborer
// 		want    *IdentityIdentifier
// 		wantErr bool
// 	}{{
// 		name: "network alias",
// 		cborer: &IdentityAlias{
// 			Network: NetworkAlias{
// 				Hostname: "nimona.io",
// 			},
// 			Handle: "geoah",
// 		},
// 		want: &IdentityIdentifier{
// 			IdentityAlias: &IdentityAlias{
// 				Network: NetworkAlias{
// 					Hostname: "nimona.io",
// 				},
// 				Handle: "geoah",
// 			},
// 		},
// 	}, {
// 		name: "network identity",
// 		cborer: &Identity{

// 		},
// 		want: &IdentityIdentifier{
// 			IdentityIdentity: &IdentityIdentity{
// 				IdentityInfoRootID: networkInfoRootID,
// 			},
// 		},
// 	}}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cborBytes, err := tt.cborer.MarshalCBORBytes()
// 			require.NoError(t, err)

// 			PrettyPrintCbor(cborBytes)

// 			id := &IdentityIdentifier{}
// 			err = id.UnmarshalCBORBytes(cborBytes)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("UnmarshalCBORBytes() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			require.Equal(t, tt.want, id)
// 		})
// 	}
// }
