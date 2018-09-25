package primitives

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignatureVerification(t *testing.T) {
	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	k, err := NewKey(sk)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	p := &Block{
		Type: "test.nimona.io/dummy",
		Payload: map[string]interface{}{
			"foo": "bar2",
		},
	}

	err = Sign(p, k)
	assert.NoError(t, err)
	assert.NotEmpty(t, p.Signature)

	// test verification
	digest, err := getDigest(p)
	assert.NoError(t, err)
	err = Verify(p.Signature, digest)
	assert.NoError(t, err)
}
