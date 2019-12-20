package main

import (
	"nimona.io/pkg/daemon"
)

func main() {
	d := daemon.New()
	d.Start()
}
