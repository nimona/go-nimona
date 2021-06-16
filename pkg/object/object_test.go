package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/chore"
)

func TestObject(t *testing.T) {
	o := chore.Map{
		"boolArray": chore.BoolArray{
			chore.Bool(false),
			chore.Bool(true),
		},
		"dataArray": chore.DataArray{
			chore.Data("v0"),
			chore.Data("v1"),
		},
		"floatArray": chore.FloatArray{
			chore.Float(0.10),
			chore.Float(1.12),
		},
		"intArray": chore.IntArray{
			chore.Int(0),
			chore.Int(1),
		},
		"mapArray": chore.MapArray{
			chore.Map{"foo0": chore.String("bar0")},
			chore.Map{
				"boolArray": chore.BoolArray{
					chore.Bool(false),
					chore.Bool(true),
				},
				"dataArray": chore.DataArray{
					chore.Data("v0"),
					chore.Data("v1"),
				},
				"floatArray": chore.FloatArray{
					chore.Float(0.10),
					chore.Float(1.12),
				},
				"intArray": chore.IntArray{
					chore.Int(0),
					chore.Int(1),
				},
				"mapArray": chore.MapArray{
					chore.Map{"foo0": chore.String("bar0")},
					chore.Map{"foo1": chore.String("bar1")},
				},
				"stringArray": chore.StringArray{
					chore.String("v0"),
					chore.String("v1"),
				},
				"uintArray": chore.UintArray{
					chore.Uint(0),
					chore.Uint(1),
				},
			},
		},
		"stringArray": chore.StringArray{
			chore.String("v0"),
			chore.String("v1"),
		},
		"uintArray": chore.UintArray{
			chore.Uint(0),
			chore.Uint(1),
		},
		"bool":  chore.Bool(true),
		"data":  chore.Data("foo"),
		"float": chore.Float(1.1),
		"int":   chore.Int(2),
		"map": chore.Map{
			"boolArray": chore.BoolArray{
				chore.Bool(false),
				chore.Bool(true),
			},
			"dataArray": chore.DataArray{
				chore.Data("v0"),
				chore.Data("v1"),
			},
			"floatArray": chore.FloatArray{
				chore.Float(0.10),
				chore.Float(1.12),
			},
			"chore.IntArray": chore.IntArray{
				chore.Int(0),
				chore.Int(1),
			},
			"mapArray": chore.MapArray{
				chore.Map{"foo0": chore.String("bar0")},
				chore.Map{"foo1": chore.String("bar1")},
			},
			"stringArray": chore.StringArray{
				chore.String("v0"),
				chore.String("v1"),
			},
			"chore.UintArray": chore.UintArray{
				chore.Uint(0),
				chore.Uint(1),
			},
			"bool":  chore.Bool(true),
			"data":  chore.Data("foo"),
			"float": chore.Float(1.1),
			"int":   chore.Int(2),
			"map": chore.Map{
				"int": chore.Int(42),
			},
			"string": chore.String("foo"),
			"uint":   chore.Uint(3),
		},
		"string": chore.String("foo"),
		"uint":   chore.Uint(3),
	}
	b, err := json.MarshalIndent(o, "", "  ")
	require.NoError(t, err)

	g := chore.Map{}
	err = json.Unmarshal(b, &g)
	require.NoError(t, err)

	require.Equal(t, o, g)
}
