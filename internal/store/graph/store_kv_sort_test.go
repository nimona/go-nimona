package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//
//    o
//   / \
//  m1  m2
//  |   | \
//  m3  m4 |
//   \ /   |
//    m6   m5
//

func Test_topographicalSort(t *testing.T) {
	graph := map[string][]string{
		"o":  []string{"m1", "m2"},
		"m1": []string{"m3"},
		"m2": []string{"m4", "m5"},
		"m3": []string{"m6"},
		"m4": []string{"m6"},
	}

	sorted := topographicalSort(graph)
	assert.Equal(t, []string{"o", "m1", "m2", "m3", "m4", "m5", "m6"}, sorted)
}
