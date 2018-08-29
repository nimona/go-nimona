package blocks

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
	X string      `nimona:"x-h,header"`
	P string      `nimona:",parent"`
}

type bPayload struct {
	AA string `nimona:"aa"`
	BB int    `nimona:"bb"`
}

func TestDecode(t *testing.T) {
	b := &Block{
		Type:      "a-type",
		Signature: quickBase58Decode("952dJcyEyxSbDRYD6WtMeFmxqBJ3FqaCvGv9NKcFeMTgh996UAya42x"),
		Payload: map[string]interface{}{
			"a": "a-value",
			"b": map[string]interface{}{
				"aa": "aa-value",
				"bb": 1212,
			},
		},
		Headers: map[string]string{
			"x-h": "x-header",
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
		X: "x-header",
	}

	s := &aPayload{
		B: &bPayload{},
	}

	DecodeInto(b, s)
	assert.Equal(t, es, s)
}
func TestDecodeMetadata(t *testing.T) {
	b := &Block{
		Type:      "a-type",
		Signature: quickBase58Decode("952dJcyEyxSbDRYD6WtMeFmxqBJ3FqaCvGv9NKcFeMTgh996UAya42x"),
		Payload: map[string]interface{}{
			"a": "a-value",
			"b": map[string]interface{}{
				"aa": "aa-value",
				"bb": 1212,
			},
		},
		Headers: map[string]string{
			"x-h": "x-header",
		},
		Metadata: &Metadata{
			Parent: "p",
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
		X: "x-header",
		P: "p",
	}

	s := &aPayload{
		B: &bPayload{},
	}

	DecodeInto(b, s)
	assert.Equal(t, es, s)
}

func TestDecodeNestedMarshaler(t *testing.T) {
	b := &Block{
		Type:      "a-type",
		Signature: quickBase58Decode("952dJcyEyxSbDRYD6WtMeFmxqBJ3FqaCvGv9NKcFeMTgh996UAya42x"),
		Payload: map[string]interface{}{
			"a": "a-value",
			"b": map[string]interface{}{
				"aa": "aa-value",
				"bb": 1212,
			},
			"s": quickBase58Decode("952dJcyEyxSbDRYD6WtMeFmxqBJ3FqaCvGv9NKcFeMTgh996UAya42x"),
			"e": quickBase58Decode("952dJcyEyxSbDRYD6WtMeFmxqBJ3FqaCvGv9NKcFeMTgh996UAya42x"),
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
	DecodeInto(b, s)
	assert.Equal(t, es, s)
}
