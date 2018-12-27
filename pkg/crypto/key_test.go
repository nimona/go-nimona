package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/encoding"
)

func TestKeyMaterialize(t *testing.T) {
	emk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	key, err := NewKey(emk)
	assert.NoError(t, err)

	nmk := key.Materialize()
	assert.Equal(t, emk, nmk)
}

func TestKeyHash(t *testing.T) {
	emk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	key, err := NewKey(emk)
	assert.NoError(t, err)

	eskh := key.ToObject().HashBase58()
	epkh := key.GetPublicKey().ToObject().HashBase58()
	assert.NotEqual(t, eskh, epkh)

	pkb, err := encoding.Marshal(key.ToObject())
	assert.NoError(t, err)
	nok, err := encoding.Unmarshal(pkb)
	assert.NoError(t, err)
	nk := &Key{}
	nk.FromObject(nok)
	assert.Equal(t, key, nk)

	nskh := nk.ToObject().HashBase58()
	assert.Equal(t, eskh, nskh)

	npkh := nk.GetPublicKey().ToObject().HashBase58()
	assert.Equal(t, epkh, npkh)
}
