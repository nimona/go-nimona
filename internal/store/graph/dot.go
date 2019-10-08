package graph

import (
	"encoding/json"
	"fmt"
	"strings"

	"nimona.io/pkg/object"
	"nimona.io/pkg/stream"
)

type graphObject struct {
	ID       string
	NodeType string
	Context  string
	Display  string
	Parents  []string
	Data     string
}

func toGraphObject(v object.Object) (*graphObject, error) {
	b, err := json.Marshal(v.ToMap())
	if err != nil {
		return nil, err
	}
	nType := "object:root"
	parents := stream.Parents(v)
	if len(parents) > 0 {
		nType = "object"
	}
	o := &graphObject{
		ID:       v.Hash().String(),
		NodeType: nType,
		Context:  v.GetType(),
		Parents:  []string{},
		Data:     string(b),
	}
	if d, ok := v.Get("@display").(string); ok {
		o.Display = d
	}
	for _, p := range stream.Parents(v) {
		o.Parents = append(o.Parents, p.Compact())
	}
	return o, nil
}

// Dot returns a graphviz representation of a graph
func Dot(objects []object.Object) (string, error) {
	graphObjects := make([]graphObject, len(objects))
	for i, o := range objects {
		igo, err := toGraphObject(o)
		if err != nil {
			return "", err
		}
		graphObjects[i] = *igo
	}
	return dot(graphObjects), nil
}

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
				p[1:idSize+1],
			)
		}
		id := fmt.Sprintf(
			`<%s>`,
			o.ID[1:idSize+1],
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
