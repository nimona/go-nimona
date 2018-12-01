package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"nimona.io/go/base58"
	"nimona.io/go/encoding"
	"nimona.io/go/peers"
)

func main() {
	names := []string{
		// "stats",
		// "andromeda",
		// "borealis",
		// "cassiopeia",
		// "draco",
		// "eridanus",
		// "fornax",
		// "gemini",
		// "hydra",
		// "indus",
		// "lacerta",
		// "mensa",
		// "norma",
		// "orion",
		// "pyxis",
		"local",
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
			// "tcp:" + name + ".nimona.io:21013",
			"tcp:localhost:21013",
		}
		b, _ := encoding.Marshal(pi)
		pp := base58.Encode(b)
		fmt.Printf(`// %s.nimona.io
// "%s",
`, name, pp)
	}
}
