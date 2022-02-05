package stream

import (
	"fmt"
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
	fmt.Println(nodes)
	require.NoError(t, err)

	require.Equal(t, len(nodes), 6)
	require.Equal(t, []string{
		nA.Key,
		nC.Key,
		nB.Key,
		nE.Key,
		nF.Key,
		nD.Key,
	}, nodes)
}
