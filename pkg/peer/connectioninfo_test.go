package peer_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestConnectionInfo_MarshalUnmarshal(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	c := &peer.ConnectionInfo{
		Owner:         peer.IDFromPublicKey(k.PublicKey()),
		Timestamp:     time.Now().Format(time.RFC3339),
		Version:       1,
		Addresses:     []string{"foo", "bar"},
		ObjectFormats: []string{"foobar"},
	}
	b, err := json.Marshal(object.MustMarshal(c))
	require.NoError(t, err)

	fmt.Println(string(b))

	g := &object.Object{}
	err = json.Unmarshal(b, g)
	require.NoError(t, err)

	fmt.Println(g.Data["addresses"])

	r := &peer.ConnectionInfo{}
	err = object.Unmarshal(g, r)
	require.NoError(t, err)

	require.Equal(t, c, r)
}
