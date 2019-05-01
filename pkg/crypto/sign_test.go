package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/internal/encoding/base58"
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
		"@ctx": "test/signed",
		"foo":  "bar",
	}

	eo := object.FromMap(m)
	assert.NotNil(t, eo)

	err = Sign(eo, subjectKey)
	assert.NoError(t, err)

	assert.NotNil(t, eo.GetSignerKey())
	assert.NotNil(t, eo.GetSignature())

	err = Verify(eo)
	assert.NoError(t, err)

	eo.SetRaw("something-new", "some-new-value")
	err = Verify(eo)
	assert.Error(t, err)

	b, _ := object.Marshal(eo)
	fmt.Println(base58.Encode(b))
}
