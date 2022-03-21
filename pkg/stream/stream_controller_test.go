package stream

import (
	"database/sql"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"nimona.io/pkg/context"
	"nimona.io/pkg/object"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

// Graph of the test stream:
//
//     (A)
//   /  |  \
// (D) (B) (C)
//      |  /
//     (E)
//      |
//     (F)
//
func Test_Controller(t *testing.T) {
	sqlStoreDB, err := sql.Open(
		"sqlite",
		path.Join(t.TempDir(), "db.sqlite"),
	)
	require.NoError(t, err)

	sqlStore, err := sqlobjectstore.New(sqlStoreDB)
	require.NoError(t, err)

	nA := &object.Object{
		Type: "test/root",
		Data: tilde.Map{
			"name": tilde.String("nA"),
		},
	}

	hA := nA.Hash()

	c := NewController(hA, nil, sqlStore)
	require.NotNil(t, c)

	nAh, err := c.Insert(nA)
	require.NoError(t, err)
	require.Equal(t, nA.Hash(), nAh)

	nB := &object.Object{
		Type: "test/event",
		Data: tilde.Map{
			"name": tilde.String("nB"),
		},
	}

	fmt.Println(">>> applying nB")
	nBh, err := c.Insert(nB)
	require.NoError(t, err)
	require.NotEqual(t, tilde.EmptyDigest, nBh)

	nC := &object.Object{
		Type: "test/event",
		Metadata: object.Metadata{
			Parents: object.Parents{
				"*": []tilde.Digest{
					nAh,
				},
			},
		},
		Data: tilde.Map{
			"name": tilde.String("nC"),
		},
	}

	fmt.Println(">>> applying nC")
	nCh, err := c.Insert(nC)
	require.NoError(t, err)
	require.NotEqual(t, tilde.EmptyDigest, nCh)

	nE := &object.Object{
		Type: "test/event",
		Data: tilde.Map{
			"name": tilde.String("nE"),
		},
	}

	fmt.Println(">>> applying nE")
	nEh, err := c.Insert(nE)
	require.NoError(t, err)
	require.NotEqual(t, tilde.EmptyDigest, nEh)

	nF := &object.Object{
		Type: "test/event",
		Data: tilde.Map{
			"name": tilde.String("nF"),
		},
	}

	fmt.Println(">>> applying nF")
	nFh, err := c.Insert(nF)
	require.NoError(t, err)
	require.NotEqual(t, tilde.EmptyDigest, nFh)

	nD := &object.Object{
		Type: "test/event",
		Metadata: object.Metadata{
			Parents: object.Parents{
				"*": []tilde.Digest{
					nAh,
				},
			},
		},
		Data: tilde.Map{
			"name": tilde.String("nD"),
		},
	}

	fmt.Println(">>> applying nD")
	nDh, err := c.Insert(nD)
	require.NoError(t, err)
	require.NotEqual(t, tilde.EmptyDigest, nDh)

	gotOrder, err := c.(*controller).GetObjectDigests()
	require.NoError(t, err)
	require.Equal(t, []tilde.Digest{
		nAh,
		// C comes before B because due to the alphabetical sorting of their
		// digests.
		nCh, nBh,
		nEh,
		nFh,
		nDh,
	}, gotOrder)

	// nX := &object.Object{
	// 	Type: "test/event",
	// 	Metadata: object.Metadata{
	// 		Parents: object.Parents{
	// 			"*": []tilde.Digest{
	// 				"doesn't exist",
	// 			},
	// 		},
	// 	},
	// 	Data: tilde.Map{
	// 		"name": tilde.String("nX"),
	// 	},
	// }

	// fmt.Println(">>> applying nX")
	// nXh, err := c.Insert(nX)
	// require.NoError(t, err)
	// require.NotEqual(t, tilde.EmptyDigest, nXh)

	fmt.Println("-------------")

	t.Run("apply all events", func(t *testing.T) {
		r, err := sqlStore.GetByStream(nAh)
		require.NoError(t, err)

		c := NewController(nAh, nil, sqlStore)
		require.NotNil(t, c)

		i := 0
		for {
			obj, err := r.Read()
			if err != nil {
				break
			}
			fmt.Printf(">>> applying %d %s\n", i, obj.Hash())
			print(obj)
			err = c.Apply(obj)
			require.NoError(t, err)
			i++
		}

		gotOrder, err := c.(*controller).GetObjectDigests()
		require.NoError(t, err)
		require.Equal(t, []tilde.Digest{
			nAh,
			// C comes before B because due to the alphabetical sorting of their
			// digests.
			nCh, nBh,
			nEh,
			nFh,
			nDh,
		}, gotOrder)
	})

	fmt.Println("-------------")

	t.Run("controller from manager", func(t *testing.T) {
		m, err := NewManager(context.New(), nil, nil, sqlStore)
		require.NoError(t, err)

		c, err := m.GetController(nAh)
		require.NotNil(t, c)
		require.NoError(t, err)

		gotOrder, err := c.(*controller).GetObjectDigests()
		require.NoError(t, err)
		require.Equal(t, []tilde.Digest{
			nAh,
			// C comes before B because due to the alphabetical sorting of their
			// digests.
			nCh, nBh,
			nEh,
			nFh,
			nDh,
		}, gotOrder)
	})
}

func print(o *object.Object) {
	m, err := o.MarshalMap()
	if err != nil {
		panic(err)
	}
	y, err := yaml.Marshal(m)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(y))
}
