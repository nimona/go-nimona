package tilde

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

var testSchema = Schema{
	Kind: KindMap,
	Properties: map[string]*Schema{
		"foo": {
			Kind: KindString,
		},
		"bar": {
			Kind: KindMap,
			Properties: map[string]*Schema{
				"baz": {
					Kind: KindString,
				},
			},
		},
		"qux": {
			Kind: KindList,
			Elements: &Schema{
				Kind: KindString,
			},
		},
		"quux": {
			Kind: KindList,
			Elements: &Schema{
				Kind: KindMap,
				Properties: map[string]*Schema{
					"quuz": {
						Kind: KindString,
					},
				},
			},
		},
	},
}

var testMap = Map{
	"foo": String("bar"),
	"bar": Map{
		"baz": String("qux"),
	},
	"qux": List{
		String("quux"),
		String("quuz"),
	},
	"quux": List{
		Map{
			"quuz": String("qux"),
		},
		Map{
			"quuz": String("quux"),
		},
	},
}

var testMapComplete = Map{
	"s": String("bar"),
	"i": Int64(-42),
	"u": Uint64(42),
	"b": Bool(true),
	"d": Bytes([]byte("bar")),
	"r": Ref("foo"),
	"m": Map{
		"s": String("qux"),
	},
	"as": List{
		String("quux"),
		String("quuz"),
	},
	"ai": List{
		Int64(-1),
		Int64(-2),
	},
	"au": List{
		Uint64(1),
		Uint64(2),
	},
	"ab": List{
		Bool(true),
		Bool(false),
	},
	"ad": List{
		Bytes([]byte("foo")),
		Bytes([]byte("bar")),
	},
	"ar": List{
		Ref("foo"),
		Ref("bar"),
	},
	"am": List{
		Map{
			"s": String("qux"),
			"i": Int64(-42),
		},
		Map{
			"s": String("quux"),
			"u": Uint64(42),
		},
	},
	"aas": List{
		List{
			String("quux"),
			String("quuz"),
		},
		List{
			String("quuz"),
			String("qux"),
		},
	},
	"aai": List{
		List{
			Int64(-1),
			Int64(-2),
		},
		List{
			Int64(-2),
			Int64(-1),
		},
	},
	"aau": List{
		List{
			Uint64(1),
			Uint64(2),
		},
		List{
			Uint64(2),
			Uint64(1),
		},
	},
	"aab": List{
		List{
			Bool(true),
			Bool(false),
		},
		List{
			Bool(false),
			Bool(true),
		},
	},
	"aad": List{
		List{
			Bytes([]byte("foo")),
			Bytes([]byte("bar")),
		},
		List{
			Bytes([]byte("bar")),
			Bytes([]byte("foo")),
		},
	},
	"aar": List{
		List{
			Ref("foo"),
			Ref("bar"),
		},
		List{
			Ref("bar"),
			Ref("foo"),
		},
		List{
			Ref("foobar"),
			Ref("foobaz"),
		},
	},
	"aam": List{
		List{
			Map{
				"s": String("qux"),
				"m": Map{
					"aaaas": List{
						List{
							List{
								List{
									List{
										String("qux"),
										String("quux"),
										String("quuz"),
									},
								},
							},
						},
					},
				},
				"aab": List{
					List{
						Bool(true),
						Bool(false),
					},
				},
			},
		},
		List{
			Map{
				"s": String("quux"),
				"m": Map{
					"am": List{
						Map{
							"s": String("qux"),
							"i": Int64(-42),
						},
						Map{
							"s": String("quux"),
							"u": Uint64(42),
						},
					},
				},
			},
		},
	},
}

func TestGetters(t *testing.T) {
	t.Run("Get foo", func(t *testing.T) {
		got, gotErr := testMap.Get("foo")
		require.NoError(t, gotErr)
		require.Equal(t, String("bar"), got)
	})

	t.Run("Get bar", func(t *testing.T) {
		got, gotErr := testMap.Get("bar")
		require.NoError(t, gotErr)
		require.Equal(t, Map{
			"baz": String("qux"),
		}, got)
	})

	t.Run("Get bar.baz", func(t *testing.T) {
		got, gotErr := testMap.Get("bar.baz")
		require.NoError(t, gotErr)
		require.Equal(t, String("qux"), got)
	})

	t.Run("Get qux", func(t *testing.T) {
		got, gotErr := testMap.Get("qux")
		require.NoError(t, gotErr)
		require.Equal(t, List{
			String("quux"),
			String("quuz"),
		}, got)
	})

	t.Run("Get qux.0", func(t *testing.T) {
		got, gotErr := testMap.Get("qux.0")
		require.NoError(t, gotErr)
		require.Equal(t, String("quux"), got)
	})

	t.Run("Get qux.1", func(t *testing.T) {
		got, gotErr := testMap.Get("qux.1")
		require.NoError(t, gotErr)
		require.Equal(t, String("quuz"), got)
	})

	t.Run("Get quux", func(t *testing.T) {
		got, gotErr := testMap.Get("quux")
		require.NoError(t, gotErr)
		require.Equal(t, List{
			Map{
				"quuz": String("qux"),
			},
			Map{
				"quuz": String("quux"),
			},
		}, got)
	})

	t.Run("Get quux.0", func(t *testing.T) {
		got, gotErr := testMap.Get("quux.0")
		require.NoError(t, gotErr)
		require.Equal(t, Map{
			"quuz": String("qux"),
		}, got)
	})

	t.Run("Get quux.1", func(t *testing.T) {
		got, gotErr := testMap.Get("quux.1")
		require.NoError(t, gotErr)
		require.Equal(t, Map{
			"quuz": String("quux"),
		}, got)
	})

	t.Run("Get quux.0.quuz", func(t *testing.T) {
		got, gotErr := testMap.Get("quux.0.quuz")
		require.NoError(t, gotErr)
		require.Equal(t, String("qux"), got)
	})

	t.Run("Get quux.1.quuz", func(t *testing.T) {
		got, gotErr := testMap.Get("quux.1.quuz")
		require.NoError(t, gotErr)
		require.Equal(t, String("quux"), got)
	})

	t.Run("Get quux.2", func(t *testing.T) {
		_, gotErr := testMap.Get("quux.2")
		require.Error(t, gotErr)
	})
}

func TestSetters(t *testing.T) {
	m := Map{}

	t.Run("Set foo", func(t *testing.T) {
		err := m.Set("foo", String("bar"))
		require.NoError(t, err)
		require.Equal(t, Map{
			"foo": String("bar"),
		}, m)
	})

	t.Run("Set bar", func(t *testing.T) {
		err := m.Set("bar", Map{
			"baz": String("qux"),
			"qux": String("quux"),
		})
		require.NoError(t, err)
		require.Equal(t, Map{
			"foo": String("bar"),
			"bar": Map{
				"baz": String("qux"),
				"qux": String("quux"),
			},
		}, m)
	})

	t.Run("Set bar.baz", func(t *testing.T) {
		err := m.Set("bar.baz", String("quxbb"))
		require.NoError(t, err)
		require.Equal(t, Map{
			"foo": String("bar"),
			"bar": Map{
				"baz": String("quxbb"),
				"qux": String("quux"),
			},
		}, m)
	})

	t.Run("Append qux", func(t *testing.T) {
		err := m.Append("qux", String("quux"))
		require.NoError(t, err)
		require.Equal(t, Map{
			"foo": String("bar"),
			"bar": Map{
				"baz": String("quxbb"),
				"qux": String("quux"),
			},
			"qux": List{
				String("quux"),
			},
		}, m)
	})

	t.Run("Set qux.0", func(t *testing.T) {
		err := m.Set("qux.0", String("quuz"))
		require.NoError(t, err)
		require.Equal(t, Map{
			"foo": String("bar"),
			"bar": Map{
				"baz": String("quxbb"),
				"qux": String("quux"),
			},
			"qux": List{
				String("quuz"),
			},
		}, m)
	})

	t.Run("Append quux", func(t *testing.T) {
		err := m.Append("quux", Map{
			"quuz": String("qux"),
		})
		require.NoError(t, err)
		require.Equal(t, Map{
			"foo": String("bar"),
			"bar": Map{
				"baz": String("quxbb"),
				"qux": String("quux"),
			},
			"qux": List{
				String("quuz"),
			},
			"quux": List{
				Map{
					"quuz": String("qux"),
				},
			},
		}, m)
	})

	t.Run("Set quux.0.quuz", func(t *testing.T) {
		err := m.Set("quux.0.quuz", String("quux"))
		require.NoError(t, err)
		require.Equal(t, Map{
			"foo": String("bar"),
			"bar": Map{
				"baz": String("quxbb"),
				"qux": String("quux"),
			},
			"qux": List{
				String("quuz"),
			},
			"quux": List{
				Map{
					"quuz": String("quux"),
				},
			},
		}, m)
	})

	t.Run("Set quux.0.quuz on empty map", func(t *testing.T) {
		m2 := Map{}
		err := m2.Set("quux.0.quuz", String("quux"))
		require.Error(t, err)

		// TODO: We might be able to support this in the future.
		// require.NoError(t, err)
		// require.Equal(t, Map{
		// 	"quux": List{
		// 		Map{
		// 			"quuz": String("quux"),
		// 		},
		// 	},
		// }, m2)
	})
}

func TestKind(t *testing.T) {
	k := KindBool
	require.Equal(t, "bool", k.String())
	require.Equal(t, "Bool", k.Name())
}

func TestJSON_Unmarshal(t *testing.T) {
	testJSON, err := json.MarshalIndent(testMapComplete, "", "  ")
	require.NoError(t, err)

	t.Run("Unmarshal", func(t *testing.T) {
		var got Map
		err := json.Unmarshal(testJSON, &got)
		require.NoError(t, err)
		require.EqualValues(t, testMapComplete, got)
	})
}

func TestCopy(t *testing.T) {
	t.Run("Copy", func(t *testing.T) {
		got := Copy(testMapComplete)
		require.NotSame(t, testMapComplete, got)
		for k, v := range testMapComplete {
			gotV, ok := got[k]
			require.True(t, ok)
			require.NotSame(t, v, gotV)
		}
		require.EqualValues(t, testMapComplete, got)
	})
}
