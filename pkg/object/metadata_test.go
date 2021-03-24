package object

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
)

func TestMetadata_Map(t *testing.T) {
	k0, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	pk0 := k0.PublicKey()

	k1, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	pk1 := k1.PublicKey()

	want := Metadata{
		Owner:    pk0,
		Datetime: "foo",
		Parents: Parents{
			"*": CIDArray{
				"bah0",
			},
			"foo.*": CIDArray{
				"bah1",
				"bah2",
			},
		},
		Policies: Policies{{
			Type:      SignaturePolicy,
			Subjects:  []crypto.PublicKey{pk0, pk1},
			Resources: []string{"foo", "bar"},
			Actions:   []PolicyAction{ReadAction, "foo", "bar"},
			Effect:    AllowEffect,
		}, {
			Type:      SignaturePolicy,
			Subjects:  []crypto.PublicKey{pk0},
			Resources: []string{"foo"},
			Actions:   []PolicyAction{ReadAction},
			Effect:    DenyEffect,
		}},
		Stream: "bah1",
		Signature: Signature{
			Signer: pk1,
			Alg:    "alg",
			X:      []byte{0, 1, 2},
			Certificate: &Certificate{
				Metadata: Metadata{
					Owner: pk1,
					Policies: Policies{{
						Type:    SignaturePolicy,
						Actions: []PolicyAction{ReadAction},
						Effect:  AllowEffect,
					}},
				},
				Nonce:   "nonce",
				Created: time.Now().UTC().Format(time.RFC3339),
				Expires: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}
	got := MetadataFromMap(want.Map())
	require.Equal(t, want, got)
}
