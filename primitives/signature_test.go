package primitives

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/codec"
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

func TestSignatureBlock(t *testing.T) {
	k := &Signature{
		Key: &Key{
			Algorithm: "key-alg",
		},
		Alg: "sig-alg",
	}

	eb := &Block{
		Type: "nimona.io/signature",
		Payload: map[string]interface{}{
			"alg": "sig-alg",
			"key": &Key{
				Algorithm: "key-alg",
			},
		},
	}

	b := k.Block()

	assert.Equal(t, eb, b)
}

func TestSignatureFromBlock(t *testing.T) {
	ek := &Signature{
		Key: &Key{
			Algorithm: "key-alg",
		},
		Alg: "sig-alg",
	}

	b := &Block{
		Type: "nimona.io/signature",
		Payload: map[string]interface{}{
			"alg": "sig-alg",
			"key": &Block{
				Type: "nimona.io/key",
				Payload: map[string]interface{}{
					"alg": "key-alg",
				},
			},
		},
	}

	k := &Signature{}
	k.FromBlock(b)

	assert.Equal(t, ek, k)
}

func TestSignatureSelfEncode(t *testing.T) {
	eb := &Signature{
		Key: &Key{
			Algorithm: "key-alg",
		},
		Alg: "sig-alg",
	}

	bs, err := codec.Marshal(eb)
	assert.NoError(t, err)

	b := &Signature{}
	err = codec.Unmarshal(bs, b)
	assert.NoError(t, err)

	assert.Equal(t, eb, b)
}
