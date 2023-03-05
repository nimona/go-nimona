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

	b, err := n0.Document().MarshalJSON()
	require.NoError(t, err)

	m1 := &Document{}
	err = m1.UnmarshalJSON(b)
	require.NoError(t, err)

	n1 := &NetworkAlias{}
	err = n1.FromDocumentMap(m1)
	require.NoError(t, err)
	require.Equal(t, n0, n1)
}

func TestNetworkIdentifier(t *testing.T) {
	networkInfoRootID := NewTestRandomDocumentID(t)
	tests := []struct {
		name    string
		doc     DocumentMapper
		want    *NetworkIdentifier
		wantErr bool
	}{{
		name: "network alias",
		doc: &NetworkAlias{
			Hostname: "nimona.io",
		},
		want: &NetworkIdentifier{
			NetworkAlias: &NetworkAlias{
				Hostname: "nimona.io",
			},
		},
	}, {
		name: "network identity",
		doc: &NetworkIdentity{
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
			b, err := tt.doc.Document().MarshalJSON()
			require.NoError(t, err)

			id := &NetworkIdentifier{}
			doc := &Document{}
			err = doc.UnmarshalJSON(b)
			require.NoError(t, err)
			err = id.FromDocumentMap(doc)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, id)
		})
	}
}
