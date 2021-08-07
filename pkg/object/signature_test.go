package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/tilde"
)

func Test_Sign(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	t.Run("should pass, sign root object", func(t *testing.T) {
		o := &Object{
			Type: "foo",
			Metadata: Metadata{
				Owner: k.PublicKey().DID(),
			},
			Data: tilde.Map{
				"foo": tilde.String("bar"),
			},
		}

		err := Sign(k, o)
		assert.NoError(t, err)

		err = Verify(o)
		require.NoError(t, err)

		assert.NotNil(t, o.Metadata.Signature)
		assert.False(t, o.Metadata.Signature.IsEmpty())
		assert.NotNil(t, o.Metadata.Signature.Signer)
	})

	t.Run("should pass, sign nested object", func(t *testing.T) {
		n := &Object{
			Type: "foo",
			Metadata: Metadata{
				Owner: k.PublicKey().DID(),
			},
			Data: tilde.Map{
				"foo": tilde.String("bar"),
			},
		}
		m, err := n.MarshalMap()
		require.NoError(t, err)
		o := &Object{
			Type: "foo",
			Data: tilde.Map{
				"foo": m,
			},
		}

		err = SignDeep(k, o)
		assert.NoError(t, err)

		err = Verify(o)
		require.NoError(t, err)

		assert.True(t, o.Metadata.Signature.IsEmpty())
		assert.Equal(t, crypto.EmptyPublicKey, o.Metadata.Signature.Signer)

		gn := &Object{}
		gm := o.Data["foo"].(tilde.Map)
		err = gn.UnmarshalMap(gm)
		require.NoError(t, err)

		assert.False(t, gn.Metadata.Signature.IsEmpty())
		assert.NotNil(t, gn.Metadata.Signature.Signer)
	})
}
