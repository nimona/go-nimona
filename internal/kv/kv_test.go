package kv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct1 struct {
	Name string
	Age  int
}

type testStruct2 struct {
	X, Y float64
}

type testStruct3 struct {
	First, Last string
}

func TestKey(t *testing.T) {
	tests := []struct {
		input  interface{}
		output string
	}{{
		input:  testStruct1{Name: "John", Age: 30},
		output: "John/30/",
	}, {
		input:  testStruct1{Name: "John"},
		output: "John/0/",
	}, {
		input:  testStruct1{Name: "Alice", Age: 25},
		output: "Alice/25/",
	}, {
		input:  testStruct2{X: 1.0, Y: 2.5},
		output: "1/2.5/",
	}, {
		input:  testStruct3{First: "John"},
		output: "John/",
	}, {
		input:  []int{1, 2, 3},
		output: "1/2/3/",
	}, {
		input:  []string{"apple", "banana", "cherry"},
		output: "apple/banana/cherry/",
	}, {
		input:  []float64{1.0, 2.5},
		output: "1/2.5/",
	}, {
		input:  []bool{true, false, true},
		output: "true/false/true/",
	}, {
		input:  123,
		output: "123/",
	}, {
		input:  3.14,
		output: "3.14/",
	}, {
		input:  "hello",
		output: "hello/",
	}, {
		input:  true,
		output: "true/",
	}}

	for _, test := range tests {
		output := keyToString(test.input)
		require.Equal(t, test.output, output)
	}
}
