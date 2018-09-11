package blocks

import (
	"testing"

	"nimona.io/go/crypto"
	"github.com/stretchr/testify/assert"
)

func init() {
	RegisterContentType(&TestPackStructA{})
	RegisterContentType(&TestPackStructB{})
}

type TestPackStructA struct {
	A string           `json:"a"`
	B *TestPackStructB `json:"b"`
	C int              `json:"c"`
	D []byte           `json:"d"`
}

func (t *TestPackStructA) GetType() string {
	return "a"
}

func (t *TestPackStructA) GetSignature() *crypto.Signature {
	// no signature
	return nil
}

func (t *TestPackStructA) SetSignature(s *crypto.Signature) {
	// no signature
}

func (t *TestPackStructA) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (t *TestPackStructA) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

type TestPackStructB struct {
	AA string `json:"aa"`
	BB int    `json:"bb"`
}

func (t *TestPackStructB) GetType() string {
	return "b"
}

func (s *TestPackStructB) GetSignature() *crypto.Signature {
	// no signature
	return nil
}

func (s *TestPackStructB) SetSignature(*crypto.Signature) {
	// no signature
}

func (s *TestPackStructB) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (s *TestPackStructB) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

func TestPack(t *testing.T) {
	s := &TestPackStructA{
		A: "a-value",
		C: 12,
		D: []byte{1, 2, 3},
	}

	ep := &Block{
		Type: "a",
		Payload: map[string]interface{}{
			"a": "a-value",
			"c": 12,
			"d": []byte{1, 2, 3},
		},
	}

	p, err := Pack(s)
	assert.NoError(t, err)
	assert.Equal(t, ep, p)
}

func TestPackNestedTyped(t *testing.T) {
	s := &TestPackStructA{
		A: "a-value",
		B: &TestPackStructB{
			AA: "aa-value",
			BB: 1234,
		},
		C: 12,
		D: []byte{1, 2, 3},
	}

	ep := &Block{
		Type: "a",
		Payload: map[string]interface{}{
			"a": "a-value",
			"b": "HCsVC4MT3AwJMXPUoHfMgQSG6GbaMDMJxpueaBFFbD6YVqF7",
			"c": 12,
			"d": []byte{1, 2, 3},
		},
	}

	p, err := Pack(s)
	assert.NoError(t, err)
	assert.Equal(t, ep, p)
}

func TestPackEncodeNestedTyped(t *testing.T) {
	s := &TestPackStructA{
		A: "a-value",
		B: &TestPackStructB{
			AA: "aa-value",
			BB: 1234,
		},
		C: 12,
		D: []byte{1, 2, 3},
	}

	ep := "2Jn7th59YMCscZggcVxGbpBVEboF4QdP2JBj9CkPJP3j7FqKC6736z6iHPFvWBop9kccPgyQ6523qXwZtcMk727eJhEhrXG4t54gkz3z4YaZYWwv6GALbcZRz"

	p, err := PackEncodeBase58(s)
	assert.NoError(t, err)
	assert.Equal(t, ep, p)
}

// func TestEncodeBlock(t *testing.T) {
// 	s := struct {
// 		A string                `json:"a"`
// 		T string                `json:",type"`
// 		S *crypto.Signature `json:"-"`
// 		// X string     `json:"x-h,header"`
// 	}{
// 		A: "a-value",
// 		T: "a-type",
// 		S: &crypto.Signature{
// 			Alg: "a-alg",
// 		},
// 		// X: "x-header",
// 	}

// 	eb := &Block{
// 		Type:      "a-type",
// 		Signature: "952dJcyEyxSbDRYD6WtMeFmxqBJ3FqaCvGv9NKcFeMTgh996UAya42x",
// 		Payload: map[string]interface{}{
// 			"a": "a-value",
// 		},
// 		// Headers: map[string]string{
// 		// 	"x-h": "x-header",
// 		// },
// 	}

// 	b := New(s).Block()
// 	assert.Equal(t, eb, b)
// }

// func TestEncodeBlockMeta(t *testing.T) {
// 	s := struct {
// 		A      string `json:"a"`
// 		Parent string `json:",parent"`
// 	}{
// 		A:      "a-value",
// 		Parent: "p",
// 	}

// 	eb := &Block{
// 		Payload: map[string]interface{}{
// 			"a": "a-value",
// 		},
// 		Metadata: &Metadata{
// 			Parent: "p",
// 		},
// 	}

// 	b := New(s).Block()
// 	assert.Equal(t, eb, b)
// }
