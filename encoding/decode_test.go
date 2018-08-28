package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type aPayload struct {
	A string      `nimona:"a"`
	B interface{} `nimona:"b"`
	C int         `nimona:"c"`
	D []byte      `nimona:"d"`
	T string      `nimona:",type"`
	S *Signature  `nimona:",signature"`
	E *Signature  `nimona:"e"`
}

type bPayload struct {
	AA string `nimona:"aa"`
	BB int    `nimona:"bb"`
}

func TestDecode(t *testing.T) {
	b := &Block{
		Type:      "a-type",
		Signature: quickBase64Decode("oWdwYXlsb2FkoWNhbGdlYS1hbGc="),
		Payload: map[string]interface{}{
			"a": "a-value",
			"b": map[string]interface{}{
				"aa": "aa-value",
				"bb": 1212,
			},
		},
	}

	es := &aPayload{
		A: "a-value",
		B: &bPayload{
			AA: "aa-value",
			BB: 1212,
		},
		T: "a-type",
		S: &Signature{
			Alg: "a-alg",
		},
	}

	s := &aPayload{
		B: &bPayload{},
	}
	Decode(b, s)
	assert.Equal(t, es, s)
}

func TestDecodeNestedMarshaler(t *testing.T) {
	b := &Block{
		Type:      "a-type",
		Signature: quickBase64Decode("oWdwYXlsb2FkoWNhbGdlYS1hbGc="),
		Payload: map[string]interface{}{
			"a": "a-value",
			"b": map[string]interface{}{
				"aa": "aa-value",
				"bb": 1212,
			},
			"s": quickBase64Decode("oWdwYXlsb2FkoWNhbGdlYS1hbGc="),
			"e": quickBase64Decode("oWdwYXlsb2FkoWNhbGdlYS1hbGc="),
		},
	}

	es := &aPayload{
		A: "a-value",
		B: &bPayload{
			AA: "aa-value",
			BB: 1212,
		},
		T: "a-type",
		E: &Signature{
			Alg: "a-alg",
		},
		S: &Signature{
			Alg: "a-alg",
		},
	}

	s := &aPayload{
		B: &bPayload{},
	}
	Decode(b, s)
	assert.Equal(t, es, s)
}
