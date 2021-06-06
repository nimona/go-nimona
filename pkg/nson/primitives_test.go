package nson

import (
	"reflect"
	"testing"
)

func TestValue(t *testing.T) {
	tests := []Value{
		new(Bool),
		new(Data),
		new(Float),
		new(Int),
		new(Map),
		new(String),
		new(Uint),
		new(CID),
	}
	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt).Name(), func(t *testing.T) {
			tt.Hint()
			tt._isValue()
		})
	}
}

func TestIsX(t *testing.T) {
	tests := []ArrayValue{
		make(BoolArray, 2),
		make(DataArray, 2),
		make(FloatArray, 2),
		make(IntArray, 2),
		make(MapArray, 2),
		make(StringArray, 2),
		make(UintArray, 2),
		make(CIDArray, 2),
		// make(ObjectArray, 2),
	}
	for _, tt := range tests {
		t.Run(reflect.TypeOf(tt).Name(), func(t *testing.T) {
			tt.Hint()
			tt._isValue()
			tt._isArray()
			tt.Len()
			done := false
			tt.Range(func(_ int, _ Value) bool {
				done = true
				return done
			})
		})
	}
}
