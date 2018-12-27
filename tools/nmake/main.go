package main

import "nimona.io/tools/nmake/cmd"

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
