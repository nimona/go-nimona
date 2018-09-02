package blocks_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/nimona/go-nimona/blocks"
	"github.com/stretchr/testify/assert"
	"github.com/ugorji/go/codec"

	"github.com/nimona/go-nimona/peers"
)

func TestSignatureMarshaling(t *testing.T) {
	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	k, err := blocks.NewKey(sk)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	p := &peers.PeerInfo{
		Addresses: []string{
			"a-1",
			"a-2",
		},
	}

	b, err := blocks.Marshal(p, blocks.SignWith(k))
	assert.NoError(t, err)
	assert.NotEmpty(t, b)

	// test block
	m := map[string]interface{}{}
	d := codec.NewDecoderBytes(b, blocks.CborHandler())
	err = d.Decode(&m)
	assert.NoError(t, err)

	assert.NotEmpty(t, m["signature"].(string))
	assert.Equal(t, m["type"], "peer.info")
	assert.Equal(t, m["payload"].(map[string]interface{})["addresses"].([]interface{})[0], "a-1")
	assert.Equal(t, m["payload"].(map[string]interface{})["addresses"].([]interface{})[1], "a-2")

	// test block's signature
	bi := m["signature"].(string)
	bbi, err := blocks.Base58Decode(bi)
	assert.NoError(t, err)
	m = map[string]interface{}{}
	d = codec.NewDecoderBytes(bbi, blocks.CborHandler())
	err = d.Decode(&m)
	assert.NoError(t, err)

	assert.Equal(t, m["type"], "signature")
	assert.Equal(t, m["payload"].(map[string]interface{})["alg"], "ES256")

	// test signature's key
	bi = m["payload"].(map[string]interface{})["key"].(string)
	bbi, err = blocks.Base58Decode(bi)
	m = map[string]interface{}{}
	d = codec.NewDecoderBytes(bbi, blocks.CborHandler())
	err = d.Decode(&m)
	assert.NoError(t, err)

	assert.Equal(t, m["type"], "key")
	assert.Equal(t, m["payload"].(map[string]interface{})["crv"], "P-256")
	assert.Equal(t, m["payload"].(map[string]interface{})["kty"], "EC")
}

func TestSignatureVerification(t *testing.T) {
	sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	k, err := blocks.NewKey(sk)
	assert.NoError(t, err)
	assert.NotNil(t, sk)

	p := &peers.PeerInfo{
		Addresses: []string{
			"a-1",
			"a-2",
		},
	}

	b, err := blocks.Marshal(p, blocks.SignWith(k))
	assert.NoError(t, err)
	assert.NotEmpty(t, b)

	// test block
	m := map[string]interface{}{}
	d := codec.NewDecoderBytes(b, blocks.CborHandler())
	err = d.Decode(&m)
	assert.NoError(t, err)

	assert.NotEmpty(t, m["signature"].(string))
	assert.Equal(t, m["type"], "peer.info")
	assert.Equal(t, m["payload"].(map[string]interface{})["addresses"].([]interface{})[0], "a-1")
	assert.Equal(t, m["payload"].(map[string]interface{})["addresses"].([]interface{})[1], "a-2")

	// test verification
	s, err := blocks.Unmarshal(b, blocks.Verify())
	assert.NoError(t, err)
	assert.NotNil(t, s)
}
