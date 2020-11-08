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
	k, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	p := &Peer{
		Metadata: object.Metadata{
			Owner: k.PublicKey(),
		},
		Addresses:    []string{"foo", "foo"},
		QueryVector:  []uint64{1, 2},
		ContentTypes: []string{"foo", "bar"},
	}

	b, err := json.Marshal(p.ToObject())
	require.NoError(t, err)

	o := &object.Object{}
	err = json.Unmarshal(b, o)
	require.NoError(t, err)

	g := &Peer{}
	err = g.FromObject(o)
	require.NoError(t, err)

	assert.Equal(t, p, g)
}

func TestPeer_EncodeDecodeWithSignature(t *testing.T) {
	k, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	p := &Peer{
		Metadata: object.Metadata{
			Owner: k.PublicKey(),
		},
		Addresses: []string{"foo", "foo"},
		QueryVector: []uint64{
			546078562,
			884891506,
			1158584717,
			2540824933,
			3828740739,
			4138058784,
		},
		ContentTypes: []string{"foo", "bar"},
	}

	p.Metadata.Signature, err = object.NewSignature(k, p.ToObject())
	require.NoError(t, err)

	err = object.Verify(p.ToObject())
	require.NoError(t, err)

	b, err := json.Marshal(p.ToObject())
	require.NoError(t, err)

	o := &object.Object{}
	err = json.Unmarshal(b, o)
	require.NoError(t, err)

	g := &Peer{}
	err = g.FromObject(o)
	require.NoError(t, err)

	assert.Equal(t, p, g)

	err = object.Verify(o)
	require.NoError(t, err)
}
