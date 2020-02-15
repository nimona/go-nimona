package object

import (
	"encoding/json"
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

	osig := copyObjectThroughJSON(t, sig.ToObject())
	nsig := &Signature{}
	err = nsig.FromObject(osig)
	assert.NoError(t, err)
	assert.Equal(t, sig, nsig)

	h := NewHash(o)
	err = sig.Signer.Verify(h.Bytes(), sig.X)
	assert.NoError(t, err)

	err = nsig.Signer.Verify(h.Bytes(), nsig.X)
	assert.NoError(t, err)

	err = Sign(o, sk)
	assert.NoError(t, err)

	err = Verify(o)
	assert.NoError(t, err)
}

func copyObjectThroughJSON(
	t *testing.T,
	o Object,
) Object {
	j, err := json.Marshal(o.ToMap())
	assert.NoError(t, err)
	m := map[string]interface{}{}
	err = json.Unmarshal(j, &m)
	assert.NoError(t, err)
	return FromMap(m)
}
