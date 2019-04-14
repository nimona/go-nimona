package object

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectHash(t *testing.T) {
	v := map[string]interface{}{
		"str:s": "foo",
	}

	kh := hash(HintString, []byte("str:s"))
	vh := hash(HintString, []byte("foo"))
	ob := append(kh, vh...)
	oh := hash(HintObject, ob)

	o := FromMap(v)
	h, err := ObjectHash(o)
	assert.NoError(t, err)
	assert.Equal(t, oh, h)
}

func TestObjectHashDocs(t *testing.T) {
	v := map[string]interface{}{
		"some-string": "bar",
		"nested-object": map[string]interface{}{
			"unsigned-number-one": 1,
			"array-of-ints:a<i>":  []int{-1, 0, 1},
		},
	}

	o := FromMap(v)
	_, err := ObjectHash(o)
	assert.NoError(t, err)
}

func TestLongObjectHash(t *testing.T) {
	v := map[string]interface{}{
		"i:i":       int(math.MaxInt32),
		"i8:i":      int8(math.MaxInt8),
		"i16:i":     int16(math.MaxInt16),
		"i32:i":     int32(math.MaxInt32),
		"i64:i":     int64(math.MaxInt64),
		"u:u":       uint(math.MaxUint32),
		"u8:u":      uint8(math.MaxUint8),
		"u16:u":     uint16(math.MaxUint16),
		"u32:u":     uint32(math.MaxUint32),
		"f32:f":     float32(math.MaxFloat32),
		"f64:f":     float64(math.MaxFloat64),
		"Ai8:a<i>":  []int8{math.MaxInt8, math.MaxInt8 - 1},
		"Ai16:a<i>": []int16{math.MaxInt16, math.MaxInt16 - 1},
		"Ai32:a<i>": []int32{math.MaxInt32, math.MaxInt32 - 1},
		"Ai64:a<i>": []int64{math.MaxInt64, math.MaxInt64 - 1},
		"Au16:a<u>": []uint16{math.MaxUint16, math.MaxUint16 - 1},
		"Au32:a<u>": []uint32{math.MaxUint32, math.MaxUint32 - 1},
		"Af32:a<f>": []float32{math.MaxFloat32, math.MaxFloat32 - 1},
		"Af64:a<f>": []float64{math.MaxFloat64, math.MaxFloat64 - 1},
		"AAi:a<a<i>>": [][]int{
			[]int{1, 2},
			[]int{3, 4},
		},
		"AAf:a<a<f>>": [][]float32{
			[]float32{math.MaxFloat32, math.MaxFloat32 - 1},
			[]float32{math.MaxFloat32, math.MaxFloat32 - 1},
		},
		"O:o": map[string]interface{}{
			"s:s": "foo",
			"u:u": uint64(12),
		},
		"bool:b": true,
	}

	o := FromMap(v)
	_, err := ObjectHash(o)
	assert.NoError(t, err)
}

func TestLongObjectHashInterfaces(t *testing.T) {
	v := map[string]interface{}{
		"I:i":     1,
		"Ai:a<i>": []interface{}{1, 2},
		"S:s":     "a",
		"As:a<s>": []interface{}{"a", "b"},
	}

	o := FromMap(v)
	h, err := ObjectHash(o)
	assert.NoError(t, err)

	b := `{"I:i":1,"Ai:a\u003ci\u003e":[1,2],"S:s":"a","As:a\u003cs\u003e":["a","b"]}` // nolint
	nv := map[string]interface{}{}
	json.Unmarshal([]byte(b), &nv) // nolint

	no := FromMap(nv)
	nh, err := ObjectHash(no)
	assert.NoError(t, err)

	assert.Equal(t, h, nh)
}
