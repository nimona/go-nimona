package main

import (
	"fmt"

	cbg "github.com/whyrusleeping/cbor-gen"
)

var mappings = map[string][]any{}

func AddMapping(f string, t any) {
	if _, ok := mappings[f]; !ok {
		mappings[f] = []any{}
	}
	mappings[f] = append(mappings[f], t)
}

func main() {
	for filename, types := range mappings {
		err := cbg.WriteMapEncodersToFile(
			filename,
			"nimona",
			types...,
		)
		if err != nil {
			panic(
				fmt.Sprintf("error writing %s, err: %s", filename, err),
			)
		}
	}
}
