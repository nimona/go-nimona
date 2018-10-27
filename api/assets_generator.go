// +build ignore

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"nimona.io/go/api"
)

func main() {
	err := vfsgen.Generate(api.Assets, vfsgen.Options{
		PackageName:  "api",
		BuildTags:    "!dev",
		VariableName: "Assets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
