package primitives

import (
	"reflect"

	"github.com/ugorji/go/codec"
)

// CborHandler for un/marshaling blocks
func CborHandler() *codec.CborHandle {
	ch := &codec.CborHandle{}
	ch.Canonical = true
	ch.DeleteOnNilMapValue = true
	ch.MapType = reflect.TypeOf(map[string]interface{}(nil))
	// blockExt := BlockExt{}
	// ch.SetExt(reflect.TypeOf(Block{}), uint64(1013), blockExt)
	return ch
}
