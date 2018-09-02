package blocks

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

func (b *bMarshallable) MarshalBlock() (string, error) {
	return Base58Encode([]byte{1, 2, 3}), nil
}

func (b *bMarshallable) UnmarshalBlock(b58bytes string) error {
	bytes, err := Base58Decode(b58bytes)
	if err != nil {
		return err
	}
	return UnmarshalInto(bytes, bytes)
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
		"b": Base58Encode([]byte{1, 2, 3}),
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
		X string     `nimona:"x-h,header"`
	}{
		A: "a-value",
		T: "a-type",
		S: &Signature{
			Alg: "a-alg",
		},
		X: "x-header",
	}

	eb := &Block{
		Type:      "a-type",
		Signature: "952dJcyEyxSbDRYD6WtMeFmxqBJ3FqaCvGv9NKcFeMTgh996UAya42x",
		Payload: map[string]interface{}{
			"a": "a-value",
		},
		Headers: map[string]string{
			"x-h": "x-header",
		},
	}

	b := New(s).Block()
	assert.Equal(t, eb, b)
}

func TestEncodeBlockMeta(t *testing.T) {
	s := struct {
		A      string `nimona:"a"`
		Parent string `nimona:",parent"`
	}{
		A:      "a-value",
		Parent: "p",
	}

	eb := &Block{
		Payload: map[string]interface{}{
			"a": "a-value",
		},
		Metadata: &Metadata{
			Parent: "p",
		},
	}

	b := New(s).Block()
	assert.Equal(t, eb, b)
}
