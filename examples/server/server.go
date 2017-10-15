package main

import fabric "github.com/nimona/go-nimona-fabric"

func main() {
	f := fabric.New("SERVER")
	f.Listen()
}
