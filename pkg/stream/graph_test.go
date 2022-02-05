package stream

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_DAG_Traverse(t *testing.T) {
	g := NewGraph[string, string]()
	nA := &Node[string, string]{
		Key:   "A",
		Value: "A",
	}
	nB := &Node[string, string]{
		Key:   "B",
		Value: "B",
		Parents: []string{
			"A",
		},
	}
	nC := &Node[string, string]{
		Key:   "C",
		Value: "C",
		Parents: []string{
			"A",
		},
	}
	nD := &Node[string, string]{
		Key:   "D",
		Value: "D",
		Parents: []string{
			"A",
		},
	}
	nE := &Node[string, string]{
		Key:   "E",
		Value: "E",
		Parents: []string{
			"B",
			"C",
		},
	}
	nF := &Node[string, string]{
		Key:   "F",
		Value: "F",
		Parents: []string{
			"E",
		},
	}

	g.AddNode(nE)
	g.AddNode(nB)
	g.AddNode(nC)
	g.AddNode(nD)
	g.AddNode(nF)
	g.AddNode(nA)

	nodes, err := g.TopologicalSort()
	require.NoError(t, err)

	require.Equal(t, 6, len(nodes))
	require.Equal(t, []string{
		nA.Key,
		nB.Key,
		nC.Key,
		nE.Key,
		nF.Key,
		nD.Key,
	}, nodes)

	leaves := g.GetLeaves()
	require.Equal(t, 2, len(leaves))
	require.Equal(t, []string{
		nF.Key,
		nD.Key,
	}, leaves)

	require.Equal(t, 0, g.countToRoot(nA.Key))
	require.Equal(t, 1, g.countToRoot(nB.Key))
	require.Equal(t, 1, g.countToRoot(nC.Key))
	require.Equal(t, 1, g.countToRoot(nD.Key))
	require.Equal(t, 3, g.countToRoot(nE.Key))
	require.Equal(t, 4, g.countToRoot(nF.Key))
}
