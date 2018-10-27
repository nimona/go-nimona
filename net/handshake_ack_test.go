package net // import "nimona.io/go/net"

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/primitives"
)

func TestKeyBlock(t *testing.T) {
	k := &HandshakeAck{
		Nonce: "foo",
		Signature: &primitives.Signature{
			Key: &primitives.Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	eb := &primitives.Block{
		Type: "nimona.io/handshake.ack",
		Payload: map[string]interface{}{
			"nonce": "foo",
		},
		Signature: &primitives.Signature{
			Key: &primitives.Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	b := k.Block()

	assert.Equal(t, eb, b)
}

func TestHandshakeAckFromBlock(t *testing.T) {
	ek := &HandshakeAck{
		Nonce: "foo",
		Signature: &primitives.Signature{
			Key: &primitives.Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	b := &primitives.Block{
		Type: "nimona.io/handshake.ack",
		Payload: map[string]interface{}{
			"nonce": "foo",
		},
		Signature: &primitives.Signature{
			Key: &primitives.Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	k := &HandshakeAck{}
	k.FromBlock(b)

	assert.Equal(t, ek, k)
}
