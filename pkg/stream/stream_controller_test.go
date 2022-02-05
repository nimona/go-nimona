package stream

import (
	"database/sql"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
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

	c := NewController(nil, sqlStore)
	require.NotNil(t, c)

	nA := &object.Object{
		Type: "test/root",
		Data: tilde.Map{
			"name": tilde.String("nA"),
		},
	}

	fmt.Println(">>> applying nA")
	nAh, err := c.Apply(nA)
	require.NoError(t, err)
	require.Equal(t, nA.Hash(), nAh)

	nB := &object.Object{
		Type: "test/event",
		Data: tilde.Map{
			"name": tilde.String("nB"),
		},
	}

	fmt.Println(">>> applying nB")
	nBh, err := c.Apply(nB)
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
	nCh, err := c.Apply(nC)
	require.NoError(t, err)
	require.NotEqual(t, tilde.EmptyDigest, nCh)

	nE := &object.Object{
		Type: "test/event",
		Data: tilde.Map{
			"name": tilde.String("nE"),
		},
	}

	fmt.Println(">>> applying nE")
	nEh, err := c.Apply(nE)
	require.NoError(t, err)
	require.NotEqual(t, tilde.EmptyDigest, nEh)

	nF := &object.Object{
		Type: "test/event",
		Data: tilde.Map{
			"name": tilde.String("nF"),
		},
	}

	fmt.Println(">>> applying nF")
	nFh, err := c.Apply(nF)
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
	nDh, err := c.Apply(nD)
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
}
