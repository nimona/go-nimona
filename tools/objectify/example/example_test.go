package example

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/object"
)

func TestGenerated(t *testing.T) {
	ff := &InnerFoo{
		InnerBar:  "foo",
		InnerBars: []string{"foo"},
		// MoreInnerFoos: []*InnerFoo{{
		// 	InnerBar: "foo",
		// }},
		// I:    int(math.MaxInt32),
		// I8:   int8(math.MaxInt8),
		// I16:  int16(math.MaxInt16),
		// I32:  int32(math.MaxInt32),
		// I64:  int64(math.MaxInt64),
		// U:    uint(math.MaxUint32),
		// U8:   uint8(math.MaxUint8),
		// U16:  uint16(math.MaxUint16),
		// U32:  uint32(math.MaxUint32),
		// F32:  float32(math.MaxFloat32),
		// F64:  float64(math.MaxFloat64),
		// Ai8:  []int8{math.MaxInt8, math.MaxInt8 - 1},
		// Ai16: []int16{math.MaxInt16, math.MaxInt16 - 1},
		// Ai32: []int32{math.MaxInt32, math.MaxInt32 - 1},
		// Ai64: []int64{math.MaxInt64, math.MaxInt64 - 1},
		// Au16: []uint16{math.MaxUint16, math.MaxUint16 - 1},
		// Au32: []uint32{math.MaxUint32, math.MaxUint32 - 1},
		// Af32: []float32{math.MaxFloat32, math.MaxFloat32 - 1},
		// Af64: []float64{math.MaxFloat64, math.MaxFloat64 - 1},
		// AAi: [][]int{
		// 	{1, 2},
		// 	{3, 4},
		// },
		// AAf: [][]float32{
		// 	{math.MaxFloat32, math.MaxFloat32 - 1},
		// 	{math.MaxFloat32, math.MaxFloat32 - 1},
		// },
		// AAs: [][]string{{"foo", "bar"}},
		// O: map[string]interface{}{
		// 	"s": "foo",
		// 	"u": uint64(12),
		// },
		B: true,
	}

	// o := object.Object{}
	// o.SetType("foo")
	// o.SetRaw("bar", "bar")

	f := &Foo{
		Bar:      "foo",
		Bars:     []string{"foo", "bar"},
		InnerFoo: ff,
		// InnerFoos: []*InnerFoo{ff},
		// Object: o,
		// Objects   []object.Object `json:"objects"`
	}

	m := f.ToObject().ToMap()
	b, _ := json.MarshalIndent(m, "", "  ")
	fmt.Println(string(b))

	uo := object.Object{}
	err := uo.FromMap(m)
	assert.NoError(t, err)

	b, _ = json.MarshalIndent(uo.ToMap(), "", "  ")
	fmt.Println(string(b))

	uf := &Foo{}
	err = uf.FromObject(uo)
	assert.NoError(t, err)
	assert.Equal(t, f, uf)
}
