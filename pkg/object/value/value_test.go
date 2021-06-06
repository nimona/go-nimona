package value

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Value(t *testing.T) {
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

func Test_IsX(t *testing.T) {
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

func Test_Marshal(t *testing.T) {
	testData, err := base64.StdEncoding.DecodeString("Zm9v")
	require.NoError(t, err)

	tests := []struct {
		name  string
		value Map
		json  string
	}{{
		name: "string",
		value: Map{
			"string": String("bar"),
		},
		json: `{"string:s":"bar"}`,
	}, {
		name: "data",
		value: Map{
			"data": Data(testData),
		},
		json: `{"data:d":"Zm9v"}`,
	}, {
		name: "bool",
		value: Map{
			"bool": Bool(true),
		},
		json: `{"bool:b":true}`,
	}, {
		name: "float",
		value: Map{
			"float": Float(1.1),
		},
		json: `{"float:f":1.1}`,
	}, {
		name: "int",
		value: Map{
			"int": Int(-2),
		},
		json: `{"int:i":-2}`,
	}, {
		name: "uint",
		value: Map{
			"uint": Uint(7),
		},
		json: `{"uint:u":7}`,
	}, {
		name: "stringArray",
		value: Map{
			"stringArray": StringArray{"foo", "bar"},
		},
		json: `{"stringArray:as":["foo","bar"]}`,
	}, {
		name: "boolArray",
		value: Map{
			"boolArray": BoolArray{true, false},
		},
		json: `{"boolArray:ab":[true,false]}`,
	}, {
		name: "floatArray",
		value: Map{
			"floatArray": FloatArray{1.0, 1.1},
		},
		json: `{"floatArray:af":[1,1.1]}`,
	}, {
		name: "intArray",
		value: Map{
			"intArray": IntArray{-2, 1},
		},
		json: `{"intArray:ai":[-2,1]}`,
	}, {
		name: "uintArray",
		value: Map{
			"uintArray": UintArray{6, 7},
		},
		json: `{"uintArray:au":[6,7]}`,
	}}

	for _, tt := range tests {
		t.Run("marshal to json: "+tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.json, string(body))
		})
		t.Run("unmarshal from json: "+tt.name, func(t *testing.T) {
			got := Map{}
			err := json.Unmarshal([]byte(tt.json), &got)
			require.NoError(t, err)
			assert.EqualValues(t, tt.value, got)
		})
	}
}
