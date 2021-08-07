package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/tilde"
)

func TestObject(t *testing.T) {
	o := tilde.Map{
		"boolArray": tilde.BoolArray{
			tilde.Bool(false),
			tilde.Bool(true),
		},
		"dataArray": tilde.DataArray{
			tilde.Data("v0"),
			tilde.Data("v1"),
		},
		"floatArray": tilde.FloatArray{
			tilde.Float(0.10),
			tilde.Float(1.12),
		},
		"intArray": tilde.IntArray{
			tilde.Int(0),
			tilde.Int(1),
		},
		"mapArray": tilde.MapArray{
			tilde.Map{"foo0": tilde.String("bar0")},
			tilde.Map{
				"boolArray": tilde.BoolArray{
					tilde.Bool(false),
					tilde.Bool(true),
				},
				"dataArray": tilde.DataArray{
					tilde.Data("v0"),
					tilde.Data("v1"),
				},
				"floatArray": tilde.FloatArray{
					tilde.Float(0.10),
					tilde.Float(1.12),
				},
				"intArray": tilde.IntArray{
					tilde.Int(0),
					tilde.Int(1),
				},
				"mapArray": tilde.MapArray{
					tilde.Map{"foo0": tilde.String("bar0")},
					tilde.Map{"foo1": tilde.String("bar1")},
				},
				"stringArray": tilde.StringArray{
					tilde.String("v0"),
					tilde.String("v1"),
				},
				"uintArray": tilde.UintArray{
					tilde.Uint(0),
					tilde.Uint(1),
				},
			},
		},
		"stringArray": tilde.StringArray{
			tilde.String("v0"),
			tilde.String("v1"),
		},
		"uintArray": tilde.UintArray{
			tilde.Uint(0),
			tilde.Uint(1),
		},
		"bool":  tilde.Bool(true),
		"data":  tilde.Data("foo"),
		"float": tilde.Float(1.1),
		"int":   tilde.Int(2),
		"map": tilde.Map{
			"boolArray": tilde.BoolArray{
				tilde.Bool(false),
				tilde.Bool(true),
			},
			"dataArray": tilde.DataArray{
				tilde.Data("v0"),
				tilde.Data("v1"),
			},
			"floatArray": tilde.FloatArray{
				tilde.Float(0.10),
				tilde.Float(1.12),
			},
			"tilde.IntArray": tilde.IntArray{
				tilde.Int(0),
				tilde.Int(1),
			},
			"mapArray": tilde.MapArray{
				tilde.Map{"foo0": tilde.String("bar0")},
				tilde.Map{"foo1": tilde.String("bar1")},
			},
			"stringArray": tilde.StringArray{
				tilde.String("v0"),
				tilde.String("v1"),
			},
			"tilde.UintArray": tilde.UintArray{
				tilde.Uint(0),
				tilde.Uint(1),
			},
			"bool":  tilde.Bool(true),
			"data":  tilde.Data("foo"),
			"float": tilde.Float(1.1),
			"int":   tilde.Int(2),
			"map": tilde.Map{
				"int": tilde.Int(42),
			},
			"string": tilde.String("foo"),
			"uint":   tilde.Uint(3),
		},
		"string": tilde.String("foo"),
		"uint":   tilde.Uint(3),
	}
	b, err := json.MarshalIndent(o, "", "  ")
	require.NoError(t, err)

	g := tilde.Map{}
	err = json.Unmarshal(b, &g)
	require.NoError(t, err)

	require.Equal(t, o, g)
}
