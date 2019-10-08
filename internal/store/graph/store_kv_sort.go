package graph

import (
	"nimona.io/pkg/object"
	"nimona.io/pkg/stream"
)

func topographicalSortObjects(os []object.Object) []object.Object {
	osm := map[string]object.Object{}
	graph := map[string][]string{}
	for _, o := range os {
		key := o.Hash().String()
		osm[key] = o
		parentKeys := stream.Parents(o)
		if _, ok := graph[key]; !ok {
			graph[key] = []string{}
		}
		for _, parentKey := range parentKeys {
			h := parentKey.Compact()
			if _, ok := graph[h]; !ok {
				graph[h] = []string{
					key,
				}
			} else {
				graph[h] = append(graph[h], key)
			}
		}
	}

	ord := topographicalSort(graph)

	// if len(ord) != len(os) {
	// 	return os
	// }

	oos := make([]object.Object, len(ord))
	for i, k := range ord {
		oos[i] = osm[k]
	}
	return oos
}

func topographicalSort(g map[string][]string) []string {
	deg := map[string]int{}
	for n := range g {
		deg[n] = 0
	}
	for _, adjacent := range g {
		for _, v := range adjacent {
			deg[v]++
		}
	}
	nxt := []string{}
	for u, v := range deg {
		if v == 0 {
			nxt = append(nxt, u)
		}
	}
	ord := []string{}
	for len(nxt) > 0 {
		u := nxt[0]
		nxt = nxt[1:]
		ord = append(ord, u)
		for _, v := range g[u] {
			deg[v]--
			if deg[v] == 0 {
				nxt = append(nxt, v)
			}
		}
	}
	return ord
}
