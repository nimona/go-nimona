package encoding

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"nimona.io/go/base58"
)

const (
	testTypeFoo       = "test/foo"
	testTypeComposite = "test/composite"
)

type TestFoo struct {
	Foo string `json:"foo"`
}

type TestComposite struct {
	Foo       string  `json:"foo"`
	Signer    *Object `json:"@signer"`
	RawObject *Object `json:"@"`
}

// func TestObjectMap(t *testing.T) {
// 	em := map[string]interface{}{
// 		"@ctx:s":          "test/message",
// 		"simple-string:s": "hello world",
// 		"nested-object:o": map[string]interface{}{
// 			"@ctx:s": "test/something-random",
// 			"foo:s":  "bar",
// 		},
// 		"@signer:O": map[string]interface{}{
// 			"@ctx:s": testTypeFoo,
// 			"crv:s":  "P-256",
// 			"kty:s":  "EC",
// 		},
// 	}

// 	o := NewObjectFromMap(em)
// 	assert.NotNil(t, o)

// 	assert.NotNil(t, o.GetSignerKey())
// 	assert.IsType(t, o.GetSignerKey(), &Object{})

// 	m := o.Map()
// 	assert.Equal(t, em, m)
// }

// func TestCompositeObjectMap(t *testing.T) {
// 	em := map[string]interface{}{
// 		"@ctx:s": testTypeComposite,
// 		"foo:s":  "hello world",
// 		"@signer:O": map[string]interface{}{
// 			"@ctx:s": testTypeFoo,
// 			"crv:s":  "P-256",
// 			"kty:s":  "EC",
// 		},
// 	}

// 	o := NewObjectFromMap(em)
// 	assert.NotNil(t, o)

// 	assert.NotNil(t, o.GetSignerKey())
// 	assert.IsType(t, &Object{}, o.GetSignerKey())

// 	m := o.Map()
// 	assert.Equal(t, em, m)
// }

// func TestObjectUnmarshal(t *testing.T) {
// 	type TestStruct struct {
// 		Foo string `json:"foo:s"`
// 	}

// 	ev := &TestStruct{
// 		Foo: "bar",
// 	}

// 	o, err := NewObjectFromMapFromStruct(ev)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, o)
// 	assert.Equal(t, "bar", o.GetRaw("foo"))

// 	v := &TestStruct{}
// 	err = o.Unmarshal(v)
// 	assert.NoError(t, err)
// 	assert.Equal(t, ev, v)
// }

func TestObjectEncodeSelf(t *testing.T) {
	type TestNestedStruct struct {
		Schema    string `json:"@ctx:s"`
		NestedFoo string `json:"nestedFoo:s"`
	}

	type TestStruct struct {
		Schema       string            `json:"@ctx:s"`
		Foo          string            `json:"foo:s"`
		NestedStruct *TestNestedStruct `json:"nestedStruct:O"`
	}

	ev := &TestStruct{
		Schema: "test/a",
		Foo:    "bar",
		NestedStruct: &TestNestedStruct{
			Schema:    "test/b",
			NestedFoo: "nested_bar",
		},
	}

	em := map[string]interface{}{
		"@ctx:s": "test/a",
		"foo:s":  "bar",
		"nestedStruct:O": map[string]interface{}{
			"@ctx:s":      "test/b",
			"nestedFoo:s": "nested_bar",
		},
	}

	o := NewObjectFromMap(em)
	assert.Equal(t, em, o.Map())
	assert.IsType(t, &Object{}, o)
	assert.IsType(t, &Object{}, o.GetRaw("nestedStruct"))

	v := &TestStruct{}
	err := o.Unmarshal(v)
	assert.NoError(t, err)
	assert.Equal(t, ev, v)

	b, err := Marshal(o)
	assert.NoError(t, err)

	fmt.Println(base58.Encode(b))

	no, err := Unmarshal(b)
	assert.NoError(t, err)
	assert.Equal(t, o.data, no.data)
}
