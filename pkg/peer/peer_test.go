package peer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func TestPeer_EncodeDecode(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	p := &ConnectionInfo{
		PublicKey: k.PublicKey(),
		Addresses: []string{"foo", "foo"},
	}

	b, err := json.Marshal(p.ToObject())
	require.NoError(t, err)

	o := &object.Object{}
	err = json.Unmarshal(b, o)
	require.NoError(t, err)

	g := &ConnectionInfo{}
	err = g.FromObject(o)
	require.NoError(t, err)

	assert.Equal(t, p, g)
}
