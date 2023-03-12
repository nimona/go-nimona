package main

import (
	"encoding/json"
	"fmt"

	"nimona.io"
)

func main() {
	asimovPublicKey, asimovPrivateKey, _ := nimona.GenerateKey()
	banksPublicKey, banksPrivateKey, _ := nimona.GenerateKey()
	currentPublic, currentPrivateKey, _ := nimona.GenerateKey()
	nextPublic, nextPrivateKey, _ := nimona.GenerateKey()

	keyGraph := nimona.KeyGraph{
		Keys: currentPublic,
		Next: nextPublic,
	}

	identity := &nimona.Identity{
		KeyGraphID: nimona.NewDocumentID(
			keyGraph.Document(),
		),
	}

	identityInfo := nimona.IdentityInfo{
		Alias: nimona.IdentityAlias{
			Hostname: "nimona.dev",
		},
		Identity: *identity,
		PeerAddresses: []nimona.PeerAddr{{
			Address:   "asimov.testing.reamde.dev:1013",
			Transport: "utp",
			PublicKey: asimovPublicKey,
		}, {
			Address:   "banks.testing.reamde.dev:1013",
			Transport: "utp",
			PublicKey: banksPublicKey,
		}},
	}

	doc := identityInfo.Document()

	rctx := &nimona.RequestContext{
		Identity:   identity,
		PublicKey:  currentPublic,
		PrivateKey: currentPrivateKey,
	}

	nimona.SignDocument(rctx, doc)

	bytes, _ := json.MarshalIndent(doc, "", "  ")
	fmt.Println(string(bytes))

	bytes, _ = json.MarshalIndent(map[string]interface{}{
		"peers": map[string]interface{}{
			"asimov": asimovPrivateKey,
			"banks":  banksPrivateKey,
		},
		"identity": map[string]interface{}{
			"current": currentPrivateKey,
			"next":    nextPrivateKey,
		},
	}, "", "  ")
	fmt.Println(string(bytes))

	fmt.Println("identity", identity.String())
	fmt.Println("asimov pk", asimovPublicKey.String())
	fmt.Println("banks pk", banksPublicKey.String())
}
