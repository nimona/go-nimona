package object

import (
	"reflect"

	"github.com/ugorji/go/codec"
)

// CborHandler for un/marshaling objects
func CborHandler() *codec.CborHandle {
	ch := &codec.CborHandle{}
	ch.Canonical = true
	ch.DeleteOnNilMapValue = true
	ch.MapType = reflect.TypeOf(map[string]interface{}(nil))
	ch.TimeRFC3339 = true
	return ch
}

// RawCborHandler for un/marshaling raw objects
func RawCborHandler() *codec.CborHandle {
	ch := CborHandler()
	ch.Raw = true
	return ch
}
