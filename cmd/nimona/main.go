package main

import (
	"nimona.io/go/cmd/nimona/cmd"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	cmd.Version = Version
	cmd.Commit = Commit
	cmd.Date = Date

	cmd.Execute()
}
