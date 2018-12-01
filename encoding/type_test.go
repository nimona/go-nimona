package encoding

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeHint(t *testing.T) {
	v := map[string]interface{}{
		"str":   "The quick brown fox jumps over the lazy dog",
		"bytes": []byte{0x7c, 0xc3, 0x6f, 0x04, 0x0a, 0x82, 0x47, 0xbb},
		"i":     int(math.MaxInt32),
		"i8":    int8(math.MaxInt8),
		"i16":   int16(math.MaxInt16),
		"i32":   int32(math.MaxInt32),
		"i64":   int64(math.MaxInt64),
		"u":     uint(math.MaxUint32),
		"u8":    uint8(math.MaxUint8),
		"u16":   uint16(math.MaxUint16),
		"u32":   uint32(math.MaxUint32),
		"f32":   float32(math.MaxFloat32),
		"f64":   float64(math.MaxFloat64),
		"Ai8":   []int8{math.MaxInt8, math.MaxInt8 - 1},
		"Ai16":  []int16{math.MaxInt16, math.MaxInt16 - 1},
		"Ai32":  []int32{math.MaxInt32, math.MaxInt32 - 1},
		"Ai64":  []int64{math.MaxInt64, math.MaxInt64 - 1},
		// "Au8":   []uint8{math.MaxUint8, math.MaxUint8 - 1},
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
		// "AO": []map[string]interface{}{
		// 	map[string]interface{}{
		// 		"s": "foo",
		// 		"u": uint64(12),
		// 	},
		// },
		// "AO2": []interface{}{
		// 	map[string]interface{}{
		// 		"s": "foo",
		// 		"u": uint64(12),
		// 	},
		// },
		"bool": true,
	}

	ev := map[string]interface{}{
		"str:s":     "The quick brown fox jumps over the lazy dog",
		"bytes:d":   []byte{0x7c, 0xc3, 0x6f, 0x04, 0x0a, 0x82, 0x47, 0xbb},
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
		"Ai8:A<i>":  []int8{math.MaxInt8, math.MaxInt8 - 1},
		"Ai16:A<i>": []int16{math.MaxInt16, math.MaxInt16 - 1},
		"Ai32:A<i>": []int32{math.MaxInt32, math.MaxInt32 - 1},
		"Ai64:A<i>": []int64{math.MaxInt64, math.MaxInt64 - 1},
		// "Au8:A<u>":  []uint8{math.MaxUint8, math.MaxUint8 - 1},
		"Au16:A<u>": []uint16{math.MaxUint16, math.MaxUint16 - 1},
		"Au32:A<u>": []uint32{math.MaxUint32, math.MaxUint32 - 1},
		"Af32:A<f>": []float32{math.MaxFloat32, math.MaxFloat32 - 1},
		"Af64:A<f>": []float64{math.MaxFloat64, math.MaxFloat64 - 1},
		"AAi:A<A<i>>": [][]int{
			[]int{1, 2},
			[]int{3, 4},
		},
		"AAf:A<A<f>>": [][]float32{
			[]float32{math.MaxFloat32, math.MaxFloat32 - 1},
			[]float32{math.MaxFloat32, math.MaxFloat32 - 1},
		},
		"O:O": map[string]interface{}{
			"s:s": "foo",
			"u:u": uint64(12),
		},
		// "AO:A<O>": []map[string]interface{}{
		// 	map[string]interface{}{
		// 		"s:s": "foo",
		// 		"u:u": uint64(12),
		// 	},
		// },
		// "AO2:A<O>": []interface{}{
		// 	map[string]interface{}{
		// 		"s:s": "foo",
		// 		"u:u": uint64(12),
		// 	},
		// },
		"bool:b": true,
	}

	nv, err := TypeMap(v)
	assert.NoError(t, err)
	assert.Equal(t, ev, nv)
}
