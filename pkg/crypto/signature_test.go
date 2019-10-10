package crypto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/hash"
	"nimona.io/pkg/object"
)

func TestNewSignature(t *testing.T) {
	sk, err := GenerateKey()
	assert.NoError(t, err)

	o := object.FromMap(map[string]interface{}{
		"foo:s": "bar",
	})

	sig, err := NewSignature(sk, AlgorithmObjectHash, o)
	assert.NoError(t, err)
	assert.Equal(t, sk.PublicKey, sig.PublicKey)

	osig := copyObjectThroughJSON(t, sig.ToObject())
	nsig := &Signature{}
	err = nsig.FromObject(osig)
	assert.NoError(t, err)
	assert.Equal(t, sig, nsig)

	h := hash.New(o)
	err = verify(sig, h.D)
	assert.NoError(t, err)

	err = verify(nsig, h.D)
	assert.NoError(t, err)

	err = Sign(o, sk)
	assert.NoError(t, err)

	err = Verify(o)
	assert.NoError(t, err)
}

func copyObjectThroughJSON(
	t *testing.T,
	o object.Object,
) object.Object {
	j, err := json.Marshal(o.ToMap())
	assert.NoError(t, err)
	m := map[string]interface{}{}
	err = json.Unmarshal(j, &m)
	assert.NoError(t, err)
	return object.FromMap(m)
}
