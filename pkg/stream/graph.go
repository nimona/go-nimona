package stream

import (
	"errors"
	"fmt"
	"sync"
)

var errNodeNotFound = errors.New("node not found")

type (
	Node[Key comparable, Value any] struct {
		Key     Key
		Value   Value
		Parents []Key
	}
	Graph[Key comparable, Value any] struct {
		nodes map[Key]*Node[Key, Value]
		lock  sync.RWMutex
	}
)

func NewGraph[Key comparable, Value any]() *Graph[Key, Value] {
	return &Graph[Key, Value]{
		nodes: map[Key]*Node[Key, Value]{},
	}
}

func (g *Graph[Key, Value]) AddNode(n *Node[Key, Value]) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.nodes[n.Key] = n
}

func (g *Graph[Key, Value]) Add(key Key, node Value, parents []Key) {
	g.AddNode(&Node[Key, Value]{
		Key:     key,
		Value:   node,
		Parents: parents,
	})
}

func (g *Graph[Key, Value]) Contains(key Key) bool {
	_, ok := g.nodes[key]
	return ok
}

// Verify that the graph is acyclic and no nodes are missing.
func (g *Graph[Key, Value]) Verify() error {
	g.lock.RLock()
	defer g.lock.RUnlock()
	foundRoot := false
	for key, node := range g.nodes {
		if node.Parents == nil {
			if foundRoot {
				return fmt.Errorf("multiple roots found")
			}
			foundRoot = true
			continue
		}
		for _, parent := range node.Parents {
			if _, ok := g.nodes[parent]; !ok {
				return fmt.Errorf("node %v is missing parent %v", key, parent)
			}
		}
	}
	return nil
}

// TopologicalSort returns a slice of nodes in topological order.
// If there are cycles in the graph, it will return an error.
//
// The algorithm is based on Kahn's algorithm.
// The implementation is based on Algorithms Unlocked [0] by Thomas H. Cormen.
// The go implementation is based on Topological Sorting [1] by Tyler Cipriani.
//
// 0. https://mitpress.mit.edu/books/algorithms-unlocked
// 1. https://tylercipriani.com/blog/2017/09/13/topographical-sorting-in-golang
//
func (g *Graph[Key, Value]) TopologicalSort() ([]Key, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	err := g.Verify()
	if err != nil {
		return nil, err
	}

	linearOrder := []Key{}
	inDegree := map[Key]int{}
	for n := range g.nodes {
		inDegree[n] = 0
	}

	for _, adjacent := range g.nodes {
		for _, v := range adjacent.Parents {
			inDegree[v]++
		}
	}

	next := []Key{}
	for u, v := range inDegree {
		if v != 0 {
			continue
		}
		next = append(next, u)
	}

	for len(next) > 0 {
		u := next[0]
		next = next[1:]
		linearOrder = append(linearOrder, u)
		for _, v := range g.nodes[u].Parents {
			inDegree[v]--
			if inDegree[v] == 0 {
				next = append(next, v)
			}
		}
	}

	for i, j := 0, len(linearOrder)-1; i < j; i, j = i+1, j-1 {
		linearOrder[i], linearOrder[j] = linearOrder[j], linearOrder[i]
	}

	return linearOrder, nil
}
