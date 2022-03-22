package hyperspace

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

func TestAnnouncement_MarshalWithSignature(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	p := &Announcement{
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: k.PublicKey().DID(),
			},
			Addresses: []string{"foo", "foo"},
		},
		Digests: []tilde.Digest{"foo", "bar"},
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
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	p := &LookupResponse{
		Announcements: []*Announcement{{
			ConnectionInfo: &peer.ConnectionInfo{
				Metadata: object.Metadata{
					Owner: k.PublicKey().DID(),
				},
				Addresses: []string{"foo", "foo"},
			},
			Digests: []tilde.Digest{"foo", "bar"},
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
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	t.Run("should pass, sign announcement ", func(t *testing.T) {
		n := &Announcement{
			Metadata: object.Metadata{
				Owner:     k.PublicKey().DID(),
				Timestamp: "foo",
			},
			ConnectionInfo: &peer.ConnectionInfo{
				Metadata: object.Metadata{
					Owner:     k.PublicKey().DID(),
					Timestamp: "foo",
				},
				Version:       2,
				Addresses:     []string{"1", "2"},
				ObjectFormats: []string{"foo", "bar"},
				Relays: []*peer.ConnectionInfo{{
					Metadata: object.Metadata{
						Owner:     k.PublicKey().DID(),
						Timestamp: "foo",
					},
					Version:       3,
					Addresses:     []string{"1", "2"},
					ObjectFormats: []string{"foo", "bar"},
					Relays:        []*peer.ConnectionInfo{},
				}},
			},
			Digests:          []tilde.Digest{"foo", "bar"},
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
