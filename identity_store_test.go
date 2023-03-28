package nimona

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityStore_E2E(t *testing.T) {
	db := NewTestDocumentDB(t)
	st, err := NewIdentityStore(db)
	require.NoError(t, err)

	id, err := st.NewIdentity("test")
	require.NoError(t, err)
	fmt.Println(id.String())

	kg, err := st.GetKeyGraph(id)
	require.NoError(t, err)

	kpc, kpn, err := st.GetKeyPairs(id)
	require.NoError(t, err)

	require.EqualValues(t, kg.Keys, kpc.PublicKey)
	require.EqualValues(t, kg.Next, kpn.PublicKey)
}
