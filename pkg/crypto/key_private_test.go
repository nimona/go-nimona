package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivateKey(t *testing.T) {
	emsk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	// create new SecretKey
	sk, err := NewPrivateKey(emsk)
	assert.NoError(t, err)
	assert.Equal(t, emsk, sk.Key)
	assert.Equal(t, &emsk.PublicKey, sk.PublicKey.Key)

	// convert SecretKey to object and back
	nsk := &PrivateKey{}
	nsk.FromObject(sk.ToObject())
	assert.Equal(t, emsk, nsk.Key)
	assert.Equal(t, &emsk.PublicKey, nsk.PublicKey.Key)
	assert.Equal(t, sk.Hash, nsk.Hash)
	assert.Equal(t, sk.PublicKey.Fingerprint(), nsk.PublicKey.Fingerprint())

	// convert PublicKey to object
	npk := &PublicKey{}
	npk.FromObject(sk.PublicKey.ToObject())
	assert.Equal(t, &emsk.PublicKey, npk.Key)
	assert.Equal(t, sk.PublicKey.Fingerprint(), nsk.PublicKey.Fingerprint())
}
