package stream

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var errNodeNotFound = errors.New("node not found")

type (
	keyable interface {
		string
	}
	Node[Key keyable, Value any] struct {
		Key     Key
		Value   Value
		Parents []Key
	}
	Graph[Key keyable, Value any] struct {
		nodes map[Key]*Node[Key, Value]
		lock  sync.RWMutex
	}
)

func NewGraph[Key keyable, Value any]() *Graph[Key, Value] {
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
// The algorithm is based on Kahn's algorithm, but has been modified to return
// consistent order for nodes with the same number of adjacent nodes.
// In case of ties, the order is determined by the following checks:
// a. The number of nodes in the path from the node to the root of the graph.
// b. The number of nodes in all paths from the node to all leaf nodes.
// c. Alphanumeric ordering of the node key.
//
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

	inVotes := map[Key]*kv[Key]{}
	for n := range g.nodes {
		inVotes[n] = &kv[Key]{
			Key:           n,
			VotesToRoot:   g.countToRoot(n),
			VotesToLeaves: 0,
		}
	}

	// find all the leaf nodes
	next := []Key{}
	for u, v := range inDegree {
		if v != 0 {
			continue
		}
		next = append(next, u)
	}

	// continue until there are no more nodes
	for len(next) > 0 {
		// sort the leaf nodes by votes
		next = sortByVotes(inVotes, next)
		// remove the first node from the list
		u := next[0]
		next = next[1:]
		// add it to the linear order
		linearOrder = append(linearOrder, u)
		// update the votes for all the nodes that it depends on
		for _, v := range g.nodes[u].Parents {
			inVotes[v].VotesToLeaves++
		}
		// go through the adjacent nodes and decrement their in-degree
		for _, v := range g.nodes[u].Parents {
			inDegree[v]--
			// if the in-degree of the node is 0, add it to the next list
			if inDegree[v] == 0 {
				next = append(next, v)
			}
		}
	}

	// reverse the linear order
	for i, j := 0, len(linearOrder)-1; i < j; i, j = i+1, j-1 {
		linearOrder[i], linearOrder[j] = linearOrder[j], linearOrder[i]
	}

	return linearOrder, nil
}

func (g *Graph[Key, Value]) countToRoot(key Key) int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	n, ok := g.nodes[key]
	if !ok {
		return 0
	}
	if n.Parents == nil {
		return 0
	}
	count := 1
	for _, parent := range n.Parents {
		count += g.countToRoot(parent)
	}
	return count
}

type kv[Key keyable] struct {
	Key           Key
	VotesToRoot   int
	VotesToLeaves int
}

func sortByVotes[Key keyable](votes map[Key]*kv[Key], keys []Key) []Key {
	ss := []kv[Key]{}
	for _, k := range keys {
		v := votes[k]
		ss = append(ss, *v)
	}

	sort.Slice(ss, func(i, j int) bool {
		if ss[i].VotesToRoot < ss[j].VotesToRoot {
			return true
		}
		if ss[i].VotesToRoot == ss[j].VotesToRoot {
			if ss[i].VotesToLeaves > ss[j].VotesToLeaves {
				return true
			}
			if ss[i].VotesToLeaves == ss[j].VotesToLeaves {
				return ss[i].Key > ss[j].Key
			}
		}
		return false
	})

	rr := []Key{}
	for _, kv := range ss {
		rr = append(rr, kv.Key)
	}

	return rr
}
