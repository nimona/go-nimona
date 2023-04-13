package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityStore_E2E(t *testing.T) {
	db := NewTestDocumentDB(t)
	st, err := NewKeygraphStore(db)
	require.NoError(t, err)

	kg, err := st.NewKeygraph("test")
	require.NoError(t, err)

	got, err := st.GetKeygraph(kg.ID())
	require.NoError(t, err)
	require.EqualValues(t, kg, got)

	kpc, kpn, err := st.GetKeyPairs(kg.ID())
	require.NoError(t, err)
	require.EqualValues(t, kg.Keys, kpc.PublicKey)
	require.EqualValues(t, kg.Next, kpn.PublicKey)
}
