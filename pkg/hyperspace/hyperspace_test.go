package hyperspace

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestAnnounce_EncodeDecodeWithSignature(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	p := &Announcement{
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: k.PublicKey(),
			Addresses: []string{"foo", "foo"},
		},
		PeerVector: []uint64{
			546078562,
			884891506,
			1158584717,
			2540824933,
			3828740739,
			4138058784,
		},
	}

	p.Metadata.Signature, err = object.NewSignature(k, p.ToObject())
	require.NoError(t, err)

	err = object.Verify(p.ToObject())
	require.NoError(t, err)

	b, err := json.MarshalIndent(p.ToObject().ToMap(), "", "  ")
	require.NoError(t, err)

	o := &object.Object{}
	err = json.Unmarshal(b, o)
	require.NoError(t, err)

	g := &Announcement{}
	err = g.FromObject(o)
	require.NoError(t, err)

	assert.Equal(t, p, g)

	err = object.Verify(o)
	require.NoError(t, err)
}

func TestResponse_EncodeDecodeWithSignature(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	p := &LookupResponse{
		Announcements: []*Announcement{{
			ConnectionInfo: &peer.ConnectionInfo{
				PublicKey: k.PublicKey(),
				Addresses: []string{"foo", "foo"},
			},
			PeerVector: []uint64{
				546078562,
				884891506,
				1158584717,
				2540824933,
				3828740739,
				4138058784,
			},
		}},
	}

	p.Metadata.Signature, err = object.NewSignature(k, p.ToObject())
	require.NoError(t, err)

	err = object.Verify(p.ToObject())
	require.NoError(t, err)

	b, err := json.MarshalIndent(p.ToObject().ToMap(), "", "  ")
	require.NoError(t, err)

	fmt.Println(string(b))

	o := &object.Object{}
	err = json.Unmarshal(b, o)
	require.NoError(t, err)

	g := &LookupResponse{}
	err = g.FromObject(o)
	require.NoError(t, err)

	assert.Equal(t, p, g)

	err = object.Verify(o)
	require.NoError(t, err)
}
