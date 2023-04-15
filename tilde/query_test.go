package tilde

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuery_Map(t *testing.T) {
	m := Map{
		"foo": Map{
			"bar": List{
				Map{
					"baz": String("qux"),
				},
				Map{
					"baz": String("quux"),
				},
				Map{
					"baz": String("xyzzy"),
				},
			},
		},
		"quuz": Int64(42),
	}

	t.Run("Select all foo.bar", func(t *testing.T) {
		q1, err := m.Query().
			Select("foo.bar").
			Exec()
		x1 := List{
			Map{
				"baz": String("qux"),
			},
			Map{
				"baz": String("quux"),
			},
			Map{
				"baz": String("xyzzy"),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x1, q1)
	})

	t.Run("Select foo.bar where baz is qux", func(t *testing.T) {
		q1, err := m.Query().
			Select("foo.bar").
			Where(
				Eq("baz", String("qux")),
			).
			Exec()
		x1 := List{
			Map{
				"baz": String("qux"),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x1, q1)
	})

	t.Run("Select whole object where quuz gt 41", func(t *testing.T) {
		q2, err := m.Query().
			Where(
				Gt("quuz", Int64(41)),
			).
			Exec()
		require.NoError(t, err)
		assert.Equal(t, m, q2)
	})

	t.Run("Select whole object where quuz gt 41", func(t *testing.T) {
		q2, err := m.Query().
			Where(
				Gt("quuz", Int64(41)),
			).
			Exec()
		require.NoError(t, err)
		assert.Equal(t, m, q2)
	})

	t.Run("Select foo.bar where baz like q%", func(t *testing.T) {
		q3, err := m.Query().
			Select("foo.bar").
			Where(
				Like("baz", "q%"),
			).
			Exec()
		x3 := List{
			Map{
				"baz": String("qux"),
			},
			Map{
				"baz": String("quux"),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x3, q3)
	})
}

func TestQuery_List(t *testing.T) {
	m := List{
		Map{
			"foo": Map{
				"bar": List{
					Map{
						"baz":  String("qux"),
						"quux": Int64(10),
					},
					Map{
						"baz":  String("xyzzy"),
						"quux": Int64(12),
					},
				},
			},
			"quuz": Int64(42),
		},
		Map{
			"foo": Map{
				"bar": List{
					Map{
						"baz":  String("quz"),
						"quux": Int64(14),
					},
				},
			},
			"quuz": Int64(39),
		},
	}

	t.Run("Select . where quuz gt 40", func(t *testing.T) {
		q1, err := m.Query().
			Where(
				Gt("quuz", Int64(40)),
			).
			Exec()
		x1 := List{
			Map{
				"foo": Map{
					"bar": List{
						Map{
							"baz":  String("qux"),
							"quux": Int64(10),
						},
						Map{
							"baz":  String("xyzzy"),
							"quux": Int64(12),
						},
					},
				},
				"quuz": Int64(42),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x1, q1)
	})

	t.Run("Select all foo.bar", func(t *testing.T) {
		q1, err := m.Query().
			Select("foo.bar").
			Exec()
		x1 := List{
			Map{
				"baz":  String("qux"),
				"quux": Int64(10),
			},
			Map{
				"baz":  String("xyzzy"),
				"quux": Int64(12),
			},
			Map{
				"baz":  String("quz"),
				"quux": Int64(14),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x1, q1)
	})

	t.Run("Select foo.bar where quux < 11", func(t *testing.T) {
		q1, err := m.Query().
			Select("foo.bar").
			Where(
				Lt("quux", Int64(11)),
			).
			Exec()
		x1 := List{
			Map{
				"baz":  String("qux"),
				"quux": Int64(10),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x1, q1)
	})

	t.Run("Select foo.bar where quux <= 10", func(t *testing.T) {
		q1, err := m.Query().
			Select("foo.bar").
			Where(
				Lte("quux", Int64(10)),
			).
			Exec()
		x1 := List{
			Map{
				"baz":  String("qux"),
				"quux": Int64(10),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x1, q1)
	})

	t.Run("Select foo.bar where quux > 11", func(t *testing.T) {
		q1, err := m.Query().
			Select("foo.bar").
			Where(
				Gt("quux", Int64(11)),
			).
			Exec()
		x1 := List{
			Map{
				"baz":  String("xyzzy"),
				"quux": Int64(12),
			},
			Map{
				"baz":  String("quz"),
				"quux": Int64(14),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x1, q1)
	})

	t.Run("Select foo.bar where baz >= x", func(t *testing.T) {
		q1, err := m.Query().
			Select("foo.bar").
			Where(
				Gte("baz", String("x")),
			).
			Exec()
		x1 := List{
			Map{
				"baz":  String("xyzzy"),
				"quux": Int64(12),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x1, q1)
	})

	t.Run("Select foo.bar where quux >= 14 and baz like q%", func(t *testing.T) {
		q1, err := m.Query().
			Select("foo.bar").
			Where(
				Gte("quux", Int64(14)),
				Like("baz", "q%"),
			).
			Exec()
		x1 := List{
			Map{
				"baz":  String("quz"),
				"quux": Int64(14),
			},
		}
		require.NoError(t, err)
		assert.Equal(t, x1, q1)
	})
}
