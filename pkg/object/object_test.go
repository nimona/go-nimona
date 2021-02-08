package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	o := Map{
		"boolArray": BoolArray{
			Bool(false),
			Bool(true),
		},
		"dataArray": DataArray{
			Data("v0"),
			Data("v1"),
		},
		"floatArray": FloatArray{
			Float(0.10),
			Float(1.12),
		},
		"intArray": IntArray{
			Int(0),
			Int(1),
		},
		"mapArray": MapArray{
			Map{"foo0": String("bar0")},
			Map{
				"boolArray": BoolArray{
					Bool(false),
					Bool(true),
				},
				"dataArray": DataArray{
					Data("v0"),
					Data("v1"),
				},
				"floatArray": FloatArray{
					Float(0.10),
					Float(1.12),
				},
				"intArray": IntArray{
					Int(0),
					Int(1),
				},
				"mapArray": MapArray{
					Map{"foo0": String("bar0")},
					Map{"foo1": String("bar1")},
				},
				"stringArray": StringArray{
					String("v0"),
					String("v1"),
				},
				"uintArray": UintArray{
					Uint(0),
					Uint(1),
				},
			},
		},
		"stringArray": StringArray{
			String("v0"),
			String("v1"),
		},
		"uintArray": UintArray{
			Uint(0),
			Uint(1),
		},
		"bool":  Bool(true),
		"data":  Data("foo"),
		"float": Float(1.1),
		"int":   Int(2),
		"map": Map{
			"boolArray": BoolArray{
				Bool(false),
				Bool(true),
			},
			"dataArray": DataArray{
				Data("v0"),
				Data("v1"),
			},
			"floatArray": FloatArray{
				Float(0.10),
				Float(1.12),
			},
			"IntArray": IntArray{
				Int(0),
				Int(1),
			},
			"mapArray": MapArray{
				Map{"foo0": String("bar0")},
				Map{"foo1": String("bar1")},
			},
			"stringArray": StringArray{
				String("v0"),
				String("v1"),
			},
			"UintArray": UintArray{
				Uint(0),
				Uint(1),
			},
			"bool":  Bool(true),
			"data":  Data("foo"),
			"float": Float(1.1),
			"int":   Int(2),
			"map": Map{
				"int": Int(42),
			},
			"string": String("foo"),
			"uint":   Uint(3),
		},
		"string": String("foo"),
		"uint":   Uint(3),
	}
	b, err := json.MarshalIndent(o, "", "  ")
	require.NoError(t, err)

	g := Map{}
	err = json.Unmarshal(b, &g)
	require.NoError(t, err)

	require.Equal(t, o, g)
}
