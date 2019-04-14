package graph

import (
	"fmt"
	"strings"
)

func dot(objects []graphObject) string {
	idSize := 5
	s := ""
	objectIDs := []string{}
	mutationIDs := []string{}
	for _, o := range objects {
		parents := make([]string, len(o.Parents))
		for i, p := range o.Parents {
			parents[i] = fmt.Sprintf(
				`<%s>`,
				p.String()[1:idSize+1],
			)
		}
		id := fmt.Sprintf(
			`<%s>`,
			o.ID.String()[1:idSize+1],
		)
		if len(parents) == 0 {
			s += fmt.Sprintf(
				"\t%s -> {} [shape=doublecircle];\n",
				id,
			)
			objectIDs = append(objectIDs, id)
		} else {
			s += fmt.Sprintf(
				"\t%s -> {%s} [shape=circle,label=\" mutates\"];\n",
				id,
				strings.Join(parents, " "),
			)
			mutationIDs = append(mutationIDs, id)
		}
	}
	m := "\trankdir=TB;\n"
	m += "\tsize=\"5,4\"\n"
	m += "\tgraph [bgcolor=white, fontname=Helvetica, fontsize=11];\n"
	m += "\tedge [fontname=Helvetica, fontcolor=grey, fontsize=9];\n"
	m += fmt.Sprintf(
		"\tnode [shape=doublecircle, fontname=Monospace, fontsize=11]; %s\n",
		strings.Join(objectIDs, " "),
	)
	m += fmt.Sprintf(
		"\tnode [shape=circle, fontname=Monospace, fontsize=11]; %s\n",
		strings.Join(mutationIDs, " "),
	)
	return fmt.Sprintf("digraph G {\n%s%s}", m, s)
}
