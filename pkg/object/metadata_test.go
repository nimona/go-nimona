package object

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	want := &Metadata{
		Owner:    pk0,
		Datetime: "foo",
		Parents: Parents{
			"*": CIDArray{
				"QmY9QbAQ2kJ67tms5t63QWPjXQ5pB5Zb7nsUa6UcTtCsxX",
			},
			"foo.*": CIDArray{
				"QmY9QbAQ2kJ67tms5t63QWPjXQ5pB5Zb7nsUa6UcTtCsxX",
				"QmY9QbAQ2kJ67tms5t63QWPjXQ5pB5Zb7nsUa6UcTtCsxX",
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
		Stream: "QmY9QbAQ2kJ67tms5t63QWPjXQ5pB5Zb7nsUa6UcTtCsxX",
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
				Starts:  time.Now().UTC().Format(time.RFC3339),
				Expires: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	t.Run("metadata of object", func(t *testing.T) {
		o := &Object{
			Metadata: *want,
			Data:     Map{},
		}
		b, err := json.MarshalIndent(o, "", "  ")
		require.NoError(t, err)

		fmt.Println(string(b))

		g := &Object{}
		err = json.Unmarshal(b, g)
		require.NoError(t, err)
		assert.Equal(t, o, g)
	})

	t.Run("metadata of nested object", func(t *testing.T) {
		o := &Object{
			Data: Map{
				"foo": &Object{
					Metadata: *want,
					Data:     Map{},
				},
			},
		}
		b, err := json.Marshal(o)
		require.NoError(t, err)

		g := &Object{}
		err = json.Unmarshal(b, g)
		require.NoError(t, err)
		assert.Equal(t, o, g)
	})
}
