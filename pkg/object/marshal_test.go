package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/tilde"
)

type (
	TestMarshalStruct struct {
		Type     string   `nimona:"@type:s"`
		Metadata Metadata `nimona:"@metadata:m"`
		TestMarshalMap
	}
	TestMarshalMap struct {
		String       string                 `nimona:"string:s"`
		Bool         bool                   `nimona:"bool:b"`
		Float32      float32                `nimona:"float32:f"`
		Float64      float64                `nimona:"float64:f"`
		Int          int                    `nimona:"int:i"`
		Int8         int8                   `nimona:"int8:i"`
		Int16        int16                  `nimona:"int16:i"`
		Int32        int32                  `nimona:"int32:i"`
		Int64        int64                  `nimona:"int64:i"`
		Uint         uint                   `nimona:"uint:u"`
		Uint8        uint8                  `nimona:"uint8:u"`
		Uint16       uint16                 `nimona:"uint16:u"`
		Uint32       uint32                 `nimona:"uint32:u"`
		Uint64       uint64                 `nimona:"uint64:u"`
		StringArray  []string               `nimona:"stringArray:as"`
		BoolArray    []bool                 `nimona:"boolArray:ab"`
		Float32Array []float32              `nimona:"float32Array:af"`
		Float64Array []float64              `nimona:"float64Array:af"`
		IntArray     []int                  `nimona:"intArray:ai"`
		Int8Array    []int8                 `nimona:"int8Array:ai"`
		Int16Array   []int16                `nimona:"int16Array:ai"`
		Int32Array   []int32                `nimona:"int32Array:ai"`
		Int64Array   []int64                `nimona:"int64Array:ai"`
		UintArray    []uint                 `nimona:"uintArray:au"`
		Uint8Array   []uint8                `nimona:"uint8Array:au"`
		Uint16Array  []uint16               `nimona:"uint16Array:au"`
		Uint32Array  []uint32               `nimona:"uint32Array:au"`
		Uint64Array  []uint64               `nimona:"uint64Array:au"`
		GoMap        map[string]interface{} `nimona:"gomap:m"`
		Stringer     crypto.PublicKey       `nimona:"stringer:s"`
	}
)

func TestMarshal(t *testing.T) {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	s := &TestMarshalStruct{
		Type: "some-type",
		Metadata: Metadata{
			Timestamp: "foo",
			Parents: Parents{
				"*": tilde.DigestArray{
					tilde.Digest("foo"),
					tilde.Digest("bar"),
				},
			},
		},
		TestMarshalMap: TestMarshalMap{
			String:       "string",
			Bool:         true,
			Float32:      0.0,
			Float64:      1.1,
			Int:          -2,
			Int8:         -3,
			Int16:        -4,
			Int32:        -5,
			Int64:        -6,
			Uint:         7,
			Uint8:        8,
			Uint16:       9,
			Uint32:       10,
			Uint64:       11,
			StringArray:  []string{"string"},
			BoolArray:    []bool{true},
			Float32Array: []float32{0.0},
			Float64Array: []float64{1.1},
			IntArray:     []int{-2},
			Int8Array:    []int8{-3},
			Int16Array:   []int16{-4},
			Int32Array:   []int32{-5},
			Int64Array:   []int64{-6},
			UintArray:    []uint{7},
			Uint8Array:   []uint8{8},
			Uint16Array:  []uint16{9},
			Uint32Array:  []uint32{10},
			Uint64Array:  []uint64{11},
			GoMap: map[string]interface{}{
				"string:s":        "string",
				"bool:b":          true,
				"float32:f":       float32(0.0),
				"float64:f":       float64(1.1),
				"int:i":           int(-2),
				"int8:i":          int8(-3),
				"int16:i":         int16(-4),
				"int32:i":         int32(-5),
				"int64:i":         int64(-6),
				"uint:u":          uint(7),
				"uint8:u":         uint8(8),
				"uint16:u":        uint16(9),
				"uint32:u":        uint32(10),
				"uint64:u":        uint64(11),
				"stringArray:as":  []string{"string"},
				"boolArray:ab":    []bool{true},
				"float32Array:af": []float32{0.0},
				"float64Array:af": []float64{1.1},
				"intArray:ai":     []int{-2},
				"int8Array:ai":    []int8{-3},
				"int16Array:ai":   []int16{-4},
				"int32Array:ai":   []int32{-5},
				"int64Array:ai":   []int64{-6},
				"uintArray:au":    []uint{7},
				"uint8Array:au":   []uint8{8},
				"uint16Array:au":  []uint16{9},
				"uint32Array:au":  []uint32{10},
				"uint64Array:au":  []uint64{11},
			},
			Stringer: k.PublicKey(),
		},
	}
	e := &Object{
		Type: "some-type",
		Metadata: Metadata{
			Timestamp: "foo",
			Parents: Parents{
				"*": tilde.DigestArray{
					tilde.Digest("foo"),
					tilde.Digest("bar"),
				},
			},
		},
		Data: tilde.Map{
			"string":       tilde.String("string"),
			"bool":         tilde.Bool(true),
			"float32":      tilde.Float(0.0),
			"float64":      tilde.Float(1.1),
			"int":          tilde.Int(-2),
			"int8":         tilde.Int(-3),
			"int16":        tilde.Int(-4),
			"int32":        tilde.Int(-5),
			"int64":        tilde.Int(-6),
			"uint":         tilde.Uint(7),
			"uint8":        tilde.Uint(8),
			"uint16":       tilde.Uint(9),
			"uint32":       tilde.Uint(10),
			"uint64":       tilde.Uint(11),
			"stringArray":  tilde.StringArray{"string"},
			"boolArray":    tilde.BoolArray{true},
			"float32Array": tilde.FloatArray{0.0},
			"float64Array": tilde.FloatArray{1.1},
			"intArray":     tilde.IntArray{-2},
			"int8Array":    tilde.IntArray{-3},
			"int16Array":   tilde.IntArray{-4},
			"int32Array":   tilde.IntArray{-5},
			"int64Array":   tilde.IntArray{-6},
			"uintArray":    tilde.UintArray{7},
			"uint8Array":   tilde.UintArray{8},
			"uint16Array":  tilde.UintArray{9},
			"uint32Array":  tilde.UintArray{10},
			"uint64Array":  tilde.UintArray{11},
			"gomap": tilde.Map{
				"string":       tilde.String("string"),
				"bool":         tilde.Bool(true),
				"float32":      tilde.Float(0.0),
				"float64":      tilde.Float(1.1),
				"int":          tilde.Int(-2),
				"int8":         tilde.Int(-3),
				"int16":        tilde.Int(-4),
				"int32":        tilde.Int(-5),
				"int64":        tilde.Int(-6),
				"uint":         tilde.Uint(7),
				"uint8":        tilde.Uint(8),
				"uint16":       tilde.Uint(9),
				"uint32":       tilde.Uint(10),
				"uint64":       tilde.Uint(11),
				"stringArray":  tilde.StringArray{"string"},
				"boolArray":    tilde.BoolArray{true},
				"float32Array": tilde.FloatArray{0.0},
				"float64Array": tilde.FloatArray{1.1},
				"intArray":     tilde.IntArray{-2},
				"int8Array":    tilde.IntArray{-3},
				"int16Array":   tilde.IntArray{-4},
				"int32Array":   tilde.IntArray{-5},
				"int64Array":   tilde.IntArray{-6},
				"uintArray":    tilde.UintArray{7},
				"uint8Array":   tilde.UintArray{8},
				"uint16Array":  tilde.UintArray{9},
				"uint32Array":  tilde.UintArray{10},
				"uint64Array":  tilde.UintArray{11},
			},
			"stringer": tilde.String(k.PublicKey().String()),
		},
	}
	g, err := Marshal(s)
	require.NoError(t, err)
	assert.NotNil(t, g)
	assert.Equal(t, e, g)
}

type TestObjectWithTags struct {
	Metadata Metadata `nimona:"@metadata:m,type=foo"`
	String   string   `nimona:"string:s"`
}

func TestMarshal_ObjectWithTags(t *testing.T) {
	s := &TestObjectWithTags{
		Metadata: Metadata{
			Timestamp: "now",
		},
		String: "hello",
	}

	e := &Object{
		Type: "foo",
		Metadata: Metadata{
			Timestamp: "now",
		},
		Data: tilde.Map{
			"string": tilde.String("hello"),
		},
	}

	g, err := Marshal(s)
	require.NoError(t, err)
	require.Equal(t, e, g)
}
