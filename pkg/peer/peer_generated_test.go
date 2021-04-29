package peer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func TestEncoding(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	c := &ConnectionInfo{
		Metadata: object.Metadata{
			Owner:    k.PublicKey(),
			Datetime: time.Now().Format(time.RFC3339),
		},
		Version:       1,
		PublicKey:     k.PublicKey(),
		Addresses:     []string{"foo", "bar"},
		ObjectFormats: []string{"foobar"},
	}
	b, err := json.Marshal(c.ToObject())
	require.NoError(t, err)

	g := &object.Object{}
	err = json.Unmarshal(b, g)
	require.NoError(t, err)

	r := &ConnectionInfo{}
	err = r.FromObject(g)
	require.NoError(t, err)

	require.Equal(t, c, r)
}
