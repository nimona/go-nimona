package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/nimona/go-nimona/peers"
)

func main() {
	names := []string{
		"andromeda",
		"borealis",
		"cassiopeia",
		"draco",
		"eridanus",
		"fornax",
		"gemini",
		"hydra",
		"indus",
		"lacerta",
		"mensa",
		"norma",
		"orion",
		"pyxis",
	}
	for _, name := range names {
		configPath := "bootstraps/" + name

		if configPath == "" {
			usr, _ := user.Current()
			configPath = path.Join(usr.HomeDir, ".nimona")
		}

		if err := os.MkdirAll(configPath, 0777); err != nil {
			log.Fatal("could not create config dir", err)
		}

		reg, err := peers.NewAddressBook(configPath)
		if err != nil {
			log.Fatal("could not load key", err)
		}

		pi := reg.GetLocalPeerInfo()
		pi.Addresses = []string{
			"tcp:" + name + ".nimona.io:21013",
		}
		pp, _ := pi.GetPeerInfo().MarshalBlock()
		fmt.Printf(`// %s.nimona.io
// "%s",
`, name, pp)
	}
}
