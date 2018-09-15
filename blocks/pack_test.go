package blocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/crypto"
)

func init() {
	RegisterContentType(&TestPackStructA{})
	RegisterContentType(&TestPackStructB{})
}

type TestPackStructA struct {
	A  string             `json:"a"`
	B  *TestPackStructB   `json:"b"`
	BS []*TestPackStructB `json:"bs"`
	C  int                `json:"c"`
	D  []byte             `json:"d"`
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
			"b": map[string]interface{}{
				"payload": map[string]interface{}{
					"aa": "aa-value",
					"bb": 1234,
				},
				"type": "b",
			},
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

	ep := "BafYRz7KhUdXVegjc94yBfMew2hJZPT2mPu2pYwGXAYLL7VXik45E9b3MbtFLLDTyYMcjcvoHPVbL61eD8hP5uyxXDN9ssEqZSdN"

	p, err := PackEncodeBase58(s)
	assert.NoError(t, err)
	assert.Equal(t, ep, p)
}

func TestPackEncodeNestedSliceTyped(t *testing.T) {
	s := &TestPackStructA{
		A: "a-value",
		BS: []*TestPackStructB{
			&TestPackStructB{
				AA: "aa-value-1",
				BB: 1,
			},
			&TestPackStructB{
				AA: "aa-value-2",
				BB: 2,
			},
		},
		C: 12,
		D: []byte{1, 2, 3},
	}

	ep := &Block{
		Type: "a",
		Payload: map[string]interface{}{
			"a": "a-value",
			"bs": []interface{}{
				map[string]interface{}{
					"payload": map[string]interface{}{
						"aa": "aa-value-1",
						"bb": 1,
					},
					"type": "b",
				},
				map[string]interface{}{
					"payload": map[string]interface{}{
						"aa": "aa-value-2",
						"bb": 2,
					},
					"type": "b",
				},
			},
			"c": 12,
			"d": []byte{1, 2, 3},
		},
	}

	p, err := Pack(s)
	assert.NoError(t, err)
	assert.Equal(t, ep, p)
}
