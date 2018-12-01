package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/base58"
)

type TestKey struct {
	Algorithm string `json:"alg:s,omitempty"`
	X         []byte `json:"x:d,omitempty"`
	Y         []byte `json:"y:d,omitempty"`
	D         []byte `json:"d:d,omitempty"`
}

func (t *TestKey) Type() string {
	return "type:key"
}

type TestSignature struct {
	Key       *TestKey `json:"key:O"`
	Alg       string   `json:"alg:s"`
	Signature []byte   `json:"sig:d"`
}

func (t *TestSignature) Type() string {
	return "type:sig"
}

type TestMessage struct {
	Body      string         `json:"body:s"`
	Timestamp string         `json:"timestamp:s"`
	Signature *TestSignature `json:"@sig:O"`
}

func (t *TestMessage) Type() string {
	return "type:msg"
}

func TestMarshalUnmarshal(t *testing.T) {
	ek := &TestKey{
		Algorithm: "a",
		X:         []byte{1, 2, 3},
		Y:         []byte{4, 5, 6},
		D:         []byte{7, 8, 9},
	}

	es := &TestSignature{
		Key:       ek,
		Alg:       "b",
		Signature: []byte{10, 11, 12},
	}

	em := &TestMessage{
		Body:      "hello",
		Timestamp: "2018-11-09T22:07:21Z", // TODO support timestamp `:t`
		Signature: es,
	}

	eo, err := NewObjectFromStruct(em)
	assert.NoError(t, err)

	bs, err := Marshal(eo)
	assert.NoError(t, err)

	assert.Equal(t, "5zrZoD7TStgnkh36YpWWitkFsUxmqfNHRp2UofB2vahpL8SAsAfEYvm4y"+
		"VuwdN5z82DNj9yuzePDXWYZdam21vTrM5B8338Z14b6RmHd2Ppj9x5DTiwVuFZdRjkhqx"+
		"tc4Hj4vEbydAcQhhWRnJ98V1jUpKHTt7Vf7Hjp7oFfsyEzMa65TeYTRKcUi3jumJqcVTs"+
		"6wnoGTKRTxaEpossrHHEK4DP", base58.Encode(bs))

	o, err := Unmarshal(bs)
	assert.NoError(t, err)
	assert.NotNil(t, o)
}
