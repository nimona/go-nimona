package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/base58"
	"nimona.io/go/encoding"
)

func TestSignWithMandate(t *testing.T) {
	authorityRawKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	assert.NotNil(t, authorityRawKey)
	authorityKey, err := NewKey(authorityRawKey)
	assert.NoError(t, err)
	assert.NotNil(t, authorityKey)

	subjectRawKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	assert.NotNil(t, subjectRawKey)
	subjectKey, err := NewKey(subjectRawKey)
	assert.NoError(t, err)
	assert.NotNil(t, subjectKey)

	p := &MandatePolicy{
		Description: "description",
		Resources: []string{
			"subject1",
			"subject2",
		},
		Actions: []string{
			"action1",
			"action2",
		},
		Effect: "effect",
	}

	m, err := NewMandate(authorityKey, subjectKey, p)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, authorityKey, m.Authority)
	assert.Equal(t, subjectKey, m.Subject)
	assert.Equal(t, p, m.Policy)
	assert.NotNil(t, m.Signature)

	o, err := encoding.NewObjectFromStruct(m)
	assert.NoError(t, err)

	b, _ := encoding.Marshal(o)
	fmt.Println(base58.Encode(b))

}
