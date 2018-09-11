package crypto_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	ucodec "github.com/ugorji/go/codec"

	"nimona.io/go/base58"
	"nimona.io/go/blocks"
	"nimona.io/go/codec"
	"nimona.io/go/crypto"
	"nimona.io/go/peers"
)

func TestSignatureMarshaling(t *testing.T) {
	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	k, err := crypto.NewKey(sk)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	p := &peers.PeerInfo{
		Addresses: []string{
			"a-1",
			"a-2",
		},
	}

	b, err := blocks.PackEncode(p, blocks.SignWith(k))
	assert.NoError(t, err)
	assert.NotEmpty(t, b)

	// test block
	m := map[string]interface{}{}
	d := ucodec.NewDecoderBytes(b, codec.CborHandler())
	err = d.Decode(&m)
	assert.NoError(t, err)

	assert.NotEmpty(t, m["signature"].(string))
	assert.Equal(t, "peer.info", m["type"])
	assert.Equal(t, "a-1", m["payload"].(map[string]interface{})["addresses"].([]interface{})[0])
	assert.Equal(t, "a-2", m["payload"].(map[string]interface{})["addresses"].([]interface{})[1])

	// test block's signature
	bi := m["signature"].(string)
	bbi, err := base58.Decode(bi)
	assert.NoError(t, err)
	m = map[string]interface{}{}
	d = ucodec.NewDecoderBytes(bbi, codec.CborHandler())
	err = d.Decode(&m)
	assert.NoError(t, err)

	assert.Equal(t, "signature", m["type"])
	assert.Equal(t, "ES256", m["payload"].(map[string]interface{})["alg"])

	// test signature's key
	bi = m["payload"].(map[string]interface{})["key"].(string)
	bbi, err = base58.Decode(bi)
	m = map[string]interface{}{}
	d = ucodec.NewDecoderBytes(bbi, codec.CborHandler())
	err = d.Decode(&m)
	assert.NoError(t, err)

	assert.Equal(t, "key", m["type"])
	assert.Equal(t, "P-256", m["payload"].(map[string]interface{})["crv"])
	assert.Equal(t, "EC", m["payload"].(map[string]interface{})["kty"])
}

func TestSignatureVerification(t *testing.T) {
	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	k, err := crypto.NewKey(sk)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	p := &peers.PeerInfo{
		Addresses: []string{
			"a-1",
			"a-2",
		},
	}

	b, err := blocks.PackEncode(p, blocks.SignWith(k))
	assert.NoError(t, err)
	assert.NotEmpty(t, b)

	// test block
	m := map[string]interface{}{}
	d := ucodec.NewDecoderBytes(b, codec.CborHandler())
	err = d.Decode(&m)
	assert.NoError(t, err)

	assert.NotEmpty(t, m["signature"].(string))
	assert.Equal(t, "peer.info", m["type"])
	assert.Equal(t, "a-1", m["payload"].(map[string]interface{})["addresses"].([]interface{})[0])
	assert.Equal(t, "a-2", m["payload"].(map[string]interface{})["addresses"].([]interface{})[1])

	// test verification
	s, err := blocks.UnpackDecode(b, blocks.Verify())
	assert.NoError(t, err)
	assert.NotNil(t, s)
}
