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

	subjectKey, err := NewKey(subjectRawKey)
	assert.NoError(t, err)
	assert.NotNil(t, subjectKey)

	m := map[string]interface{}{
		"@ctx": "test/signed",
		"foo":  "bar",
	}

	eo := object.NewObjectFromMap(m)
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

	description := "description"
	resources := []string{
		"subject1",
		"subject2",
	}
	actions := []string{
		"action1",
		"action2",
	}
	effect := "effect"

	m, err := NewMandate(authorityKey, subjectKey, description, resources, actions, effect)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, authorityKey.GetPublicKey().HashBase58(), m.Signer.HashBase58())
	assert.Equal(t, subjectKey.GetPublicKey().HashBase58(), m.Subject.HashBase58())
	assert.Equal(t, description, m.Description)
	assert.Equal(t, resources, m.Resources)
	assert.Equal(t, actions, m.Actions)
	assert.Equal(t, effect, m.Effect)

	o := m.ToObject()
	err = Verify(o)
	assert.NoError(t, err)
}
