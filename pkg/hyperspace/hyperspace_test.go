package hyperspace

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestAnnouncement_MarshalWithSignature(t *testing.T) {
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

	p.Metadata.Signature, err = object.NewSignature(k, object.MustMarshal(p))
	require.NoError(t, err)

	err = object.Verify(object.MustMarshal(p))
	require.NoError(t, err)

	b, err := json.MarshalIndent(object.MustMarshal(p), "", "  ")
	require.NoError(t, err)

	o := &object.Object{}
	err = json.Unmarshal(b, o)
	require.NoError(t, err)

	g := &Announcement{}
	err = object.Unmarshal(o, g)
	require.NoError(t, err)

	assert.Equal(t, p, g)

	err = object.Verify(o)
	require.NoError(t, err)
}

func TestResponse_MarshalWithSignature(t *testing.T) {
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

	p.Metadata.Signature, err = object.NewSignature(k, object.MustMarshal(p))
	require.NoError(t, err)

	err = object.Verify(object.MustMarshal(p))
	require.NoError(t, err)

	b, err := json.MarshalIndent(object.MustMarshal(p), "", "  ")
	require.NoError(t, err)

	o := &object.Object{}
	err = json.Unmarshal(b, o)
	require.NoError(t, err)

	g := &LookupResponse{}
	err = object.Unmarshal(o, g)
	require.NoError(t, err)

	assert.Equal(t, p, g)

	err = object.Verify(o)
	require.NoError(t, err)
}

// Test_SignDeep is testing announcement deep signing as the connection info
// is pretty deeply embedded and is a good edge case for signing.
func TestAnnouncement_SignDeep(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	t.Run("should pass, sign announcement ", func(t *testing.T) {
		n := &Announcement{
			Metadata: object.Metadata{
				Owner:    k.PublicKey(),
				Datetime: "foo",
			},
			ConnectionInfo: &peer.ConnectionInfo{
				Metadata: object.Metadata{
					Owner:    k.PublicKey(),
					Datetime: "foo",
				},
				Version:       2,
				PublicKey:     k.PublicKey(),
				Addresses:     []string{"1", "2"},
				ObjectFormats: []string{"foo", "bar"},
				Relays: []*peer.ConnectionInfo{{
					Metadata: object.Metadata{
						Owner:    k.PublicKey(),
						Datetime: "foo",
					},
					Version:       3,
					PublicKey:     k.PublicKey(),
					Addresses:     []string{"1", "2"},
					ObjectFormats: []string{"foo", "bar"},
					Relays:        []*peer.ConnectionInfo{},
				}},
			},
			PeerVector:       []uint64{0, 1, 2},
			Version:          1,
			PeerCapabilities: []string{"a", "b"},
		}

		// marshal to object
		no, err := object.Marshal(n)
		assert.NoError(t, err)

		// sign
		err = object.SignDeep(k, no)
		assert.NoError(t, err)

		// verify
		err = object.Verify(no)
		require.NoError(t, err)

		// marshal to json
		b, err := json.MarshalIndent(no, "", "  ")
		assert.NoError(t, err)

		// unmarshal to object
		o := &object.Object{}
		err = json.Unmarshal(b, o)
		require.NoError(t, err)

		// verify
		err = object.Verify(o)
		require.NoError(t, err)

		// unmarshal to struct
		nn := &Announcement{}
		err = object.Unmarshal(no, nn)
		require.NoError(t, err)
		require.Equal(t, no.Metadata, nn.Metadata)

		// marshal to object
		ng, err := object.Marshal(nn)
		assert.NoError(t, err)
		assert.Equal(t, no, ng)
	})
}
