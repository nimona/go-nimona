package object

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/did"
)

func TestMetadata_Map(t *testing.T) {
	k0, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	pk0 := k0.PublicKey()

	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	pk1 := k1.PublicKey()

	want := &Metadata{
		Owner:    *did.MustParse("did:nimona:foo"),
		Datetime: "foo",
		Parents: Parents{
			"*": chore.HashArray{
				chore.Hash("foo"),
			},
			"foo.*": chore.HashArray{
				chore.Hash("foo"),
				chore.Hash("foo"),
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
		Root: chore.Hash("foo"),
		Signature: Signature{
			Signer: pk1,
			Alg:    "alg",
			X:      []byte{0, 1, 2},
		},
	}

	t.Run("metadata of object", func(t *testing.T) {
		o := &Object{
			Metadata: *want,
			Data:     chore.Map{},
		}

		b, err := json.MarshalIndent(o, "", "  ")
		require.NoError(t, err)

		fmt.Println(string(b))

		g := &Object{}
		err = json.Unmarshal(b, g)
		require.NoError(t, err)

		require.NoError(t, err)
		assert.Equal(t, o, g)
	})

	t.Run("metadata of nested object", func(t *testing.T) {
		no := &Object{
			Metadata: *want,
			Data:     chore.Map{},
		}
		nm, err := no.MarshalMap()
		require.NoError(t, err)
		o := &Object{
			Data: chore.Map{
				"foo": nm,
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
