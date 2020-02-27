package object

import (
	"testing"

	"nimona.io/pkg/crypto"

	"github.com/stretchr/testify/assert"
)

func TestNewSignature(t *testing.T) {
	sk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	o := FromMap(map[string]interface{}{
		"foo:s": "bar",
	})

	sig, err := NewSignature(sk, o)
	assert.NoError(t, err)
	assert.Equal(t, sk.PublicKey(), sig.Signer)

	o = o.AddSignature(sig)
	err = Verify(o)
	assert.NoError(t, err)
}
