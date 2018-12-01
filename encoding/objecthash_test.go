package encoding

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectHash(t *testing.T) {
	v := map[string]interface{}{
		"str": "foo",
	}

	kh := hash(HintString, []byte("str:s"))
	vh := hash(HintString, []byte("foo"))
	ob := append(kh, vh...)
	oh := hash(HintMap, ob)

	o := NewObjectFromMap(v)
	h, err := ObjectHash(o)
	assert.NoError(t, err)
	assert.Equal(t, oh, h)
}

func TestObjectHashDocs(t *testing.T) {
	v := map[string]interface{}{
		"some-string:s": "bar",
		"nested-object:o": map[string]interface{}{
			"unsigned-number-one:u": 1,
			"array-of-ints:a<i>":    []int{-1, 0, 1},
		},
	}

	o := NewObjectFromMap(v)
	h, err := ObjectHash(o)
	assert.NoError(t, err)

	fmt.Printf("%x", h)
}

func TestLongObjectHash(t *testing.T) {
	v := map[string]interface{}{
		"i":    int(math.MaxInt32),
		"i8":   int8(math.MaxInt8),
		"i16":  int16(math.MaxInt16),
		"i32":  int32(math.MaxInt32),
		"i64":  int64(math.MaxInt64),
		"u":    uint(math.MaxUint32),
		"u8":   uint8(math.MaxUint8),
		"u16":  uint16(math.MaxUint16),
		"u32":  uint32(math.MaxUint32),
		"f32":  float32(math.MaxFloat32),
		"f64":  float64(math.MaxFloat64),
		"Ai8":  []int8{math.MaxInt8, math.MaxInt8 - 1},
		"Ai16": []int16{math.MaxInt16, math.MaxInt16 - 1},
		"Ai32": []int32{math.MaxInt32, math.MaxInt32 - 1},
		"Ai64": []int64{math.MaxInt64, math.MaxInt64 - 1},
		"Au16": []uint16{math.MaxUint16, math.MaxUint16 - 1},
		"Au32": []uint32{math.MaxUint32, math.MaxUint32 - 1},
		"Af32": []float32{math.MaxFloat32, math.MaxFloat32 - 1},
		"Af64": []float64{math.MaxFloat64, math.MaxFloat64 - 1},
		"AAi": [][]int{
			[]int{1, 2},
			[]int{3, 4},
		},
		"AAf": [][]float32{
			[]float32{math.MaxFloat32, math.MaxFloat32 - 1},
			[]float32{math.MaxFloat32, math.MaxFloat32 - 1},
		},
		"O": map[string]interface{}{
			"s": "foo",
			"u": uint64(12),
		},
		"bool": true,
	}

	o := NewObjectFromMap(v)
	h, err := ObjectHash(o)
	assert.NoError(t, err)

	fmt.Printf("% x", h)
}
