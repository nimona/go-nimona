package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/pkg/object"
)

func TestSignAndVerify(t *testing.T) {
	subjectRawKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	assert.NotNil(t, subjectRawKey)

	subjectKey, err := NewPrivateKey(subjectRawKey)
	assert.NoError(t, err)
	assert.NotNil(t, subjectKey)

	m := map[string]interface{}{
		"@ctx:s": "test/signed",
		"foo:s":  "bar",
	}

	eo := object.FromMap(m)
	assert.NotNil(t, eo)

	err = Sign(eo, subjectKey)
	assert.NoError(t, err)

	es, err := GetObjectSignature(eo)
	assert.NoError(t, err)
	assert.NotNil(t, es)
	assert.NotNil(t, es.PublicKey)

	err = Verify(eo)
	assert.NoError(t, err)

	eo.Set("something-new:s", "some-new-value")
	err = Verify(eo)
	assert.Error(t, err)
}
