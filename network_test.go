package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetworkAlias_ParseNetworkAlias(t *testing.T) {
	s0 := "nimona://net:alias:testing.reamde.dev"
	n0 := NetworkAlias{
		Hostname: "testing.reamde.dev",
	}

	require.Equal(t, s0, n0.String())

	n1, err := ParseNetworkAlias(s0)
	require.NoError(t, err)
	require.Equal(t, n0, n1)
}

func TestNetworkAlias_MarshalUnmarshal(t *testing.T) {
	n0 := &NetworkAlias{
		Hostname: "testing.reamde.dev",
	}

	b, err := MarshalCBORBytes(n0)
	require.NoError(t, err)

	n1 := &NetworkAlias{}
	err = UnmarshalCBORBytes(b, n1)
	require.NoError(t, err)

	require.Equal(t, n0, n1)
}

func TestNetworkIdentifier(t *testing.T) {
	networkInfoRootID := NewTestRandomDocumentID(t)
	tests := []struct {
		name    string
		cborer  Cborer
		want    *NetworkIdentifier
		wantErr bool
	}{{
		name: "network alias",
		cborer: &NetworkAlias{
			Hostname: "nimona.io",
		},
		want: &NetworkIdentifier{
			NetworkAlias: &NetworkAlias{
				Hostname: "nimona.io",
			},
		},
	}, {
		name: "network identity",
		cborer: &NetworkIdentity{
			NetworkInfoRootID: networkInfoRootID,
		},
		want: &NetworkIdentifier{
			NetworkIdentity: &NetworkIdentity{
				NetworkInfoRootID: networkInfoRootID,
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cborBytes, err := MarshalCBORBytes(tt.cborer)
			require.NoError(t, err)

			id := &NetworkIdentifier{}
			err = UnmarshalCBORBytes(cborBytes, id)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalCBORBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, id)
		})
	}
}
