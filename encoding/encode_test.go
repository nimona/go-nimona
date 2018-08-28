package encoding

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func quickBase58Decode(s string) []byte {
	b, err := Base58Decode(s)
	if err != nil {
		panic(err)
	}

	return b
}

func quickBase64Decode(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}

	return b
}

func TestMarshal(t *testing.T) {
	s := struct {
		A string      `nimona:"a"`
		B interface{} `nimona:"b"`
		C int         `nimona:"c"`
		D []byte      `nimona:"d"`
		E *Signature  `nimona:"e_sig"`
		S *Signature  `nimona:",signature"`
		T string      `nimona:",type"`
	}{
		A: "a-value",
		B: struct {
			AA string `nimona:"aa"`
			BB int    `nimona:"bb"`
		}{
			AA: "aa-value",
			BB: 1212,
		},
		C: 12,
		D: []byte{1, 2, 3},
		E: &Signature{
			Alg: "e-alg",
		},
		S: &Signature{
			Alg: "a-alg",
		},
		T: "a-type",
	}

	b, err := Marshal(s)
	assert.NoError(t, err)
	assert.NotNil(t, b)
}

func TestEncodeMap(t *testing.T) {
	s := struct {
		A string      `nimona:"a"`
		B interface{} `nimona:"b"`
		C int         `nimona:"c"`
		D []byte      `nimona:"d"`
	}{
		A: "a-value",
		B: struct {
			AA string `nimona:"aa"`
			BB int    `nimona:"bb"`
		}{
			AA: "aa-value",
			BB: 1212,
		},
		C: 12,
		D: []byte{1, 2, 3},
	}

	es := map[string]interface{}{
		"a": "a-value",
		"b": map[string]interface{}{
			"aa": "aa-value",
			"bb": 1212,
		},
		"c": 12,
		"d": []byte{1, 2, 3},
	}

	assert.Equal(t, es, New(&s).Map())
	assert.Equal(t, es, New(s).Map())
}

type bMarshallable struct {
	AA string `nimona:"aa"`
	BB int    `nimona:"bb"`
}

func (b *bMarshallable) MarshalBlock() ([]byte, error) {
	// return Marshal(b)
	return []byte{1, 2, 3}, nil
}

func (b *bMarshallable) UnmarshalBlock(bytes []byte) error {
	return Unmarshal(bytes, b)
}

func TestEncodeMapNestedMarshaler(t *testing.T) {
	s := struct {
		A string      `nimona:"a"`
		B interface{} `nimona:"b"`
		C int         `nimona:"c"`
		D []byte      `nimona:"d"`
	}{
		A: "a-value",
		B: &bMarshallable{
			AA: "aa-value",
			BB: 1212,
		},
		C: 12,
		D: []byte{1, 2, 3},
	}

	es := map[string]interface{}{
		"a": "a-value",
		"b": []byte{1, 2, 3},
		"c": 12,
		"d": []byte{1, 2, 3},
	}

	assert.EqualValues(t, es, New(s).Map())
}

func TestEncodeBlock(t *testing.T) {
	s := struct {
		A string     `nimona:"a"`
		T string     `nimona:",type"`
		S *Signature `nimona:",signature"`
	}{
		A: "a-value",
		T: "a-type",
		S: &Signature{
			Alg: "a-alg",
		},
	}

	eb := &Block{
		Type:      "a-type",
		Signature: quickBase64Decode("oWdwYXlsb2FkoWNhbGdlYS1hbGc="),
		Payload: map[string]interface{}{
			"a": "a-value",
		},
	}

	assert.Equal(t, eb, New(s).Block())
}
