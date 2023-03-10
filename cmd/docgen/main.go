package main

import (
	"fmt"

	"nimona.io"
)

func main() {
	mappings := map[string][]any{}
	addMapping := func(f string, t any) {
		if _, ok := mappings[f]; !ok {
			mappings[f] = []any{}
		}
		mappings[f] = append(mappings[f], t)
	}

	addMapping("document_id_docgen.go", nimona.DocumentID{})
	addMapping("document_metadata_docgen.go", nimona.Metadata{})
	addMapping("document_metadata_docgen.go", nimona.Permissions{})
	addMapping("document_metadata_docgen.go", nimona.PermissionsAllow{})
	addMapping("document_metadata_docgen.go", nimona.PermissionsCondition{})
	addMapping("document_metadata_docgen.go", nimona.Signature{})
	addMapping("document_patch_docgen.go", nimona.DocumentPatch{})
	addMapping("document_patch_docgen.go", nimona.DocumentPatchOperation{})
	addMapping("fixtures_docgen.go", nimona.CborFixture{})
	addMapping("handler_document_docgen.go", nimona.DocumentRequest{})
	addMapping("handler_document_docgen.go", nimona.DocumentResponse{})
	addMapping("handler_document_graph_docgen.go", nimona.DocumentGraphRequest{})
	addMapping("handler_document_graph_docgen.go", nimona.DocumentGraphResponse{})
	addMapping("handler_peer_docgen.go", nimona.PeerCapabilitiesRequest{})
	addMapping("handler_peer_docgen.go", nimona.PeerCapabilitiesResponse{})
	addMapping("handler_ping_docgen.go", nimona.Ping{})
	addMapping("handler_ping_docgen.go", nimona.Pong{})
	addMapping("identity_docgen.go", nimona.Identity{})
	addMapping("identity_docgen.go", nimona.IdentityAlias{})
	addMapping("identity_docgen.go", nimona.KeyGraph{})
	addMapping("model_profile_docgen.go", nimona.Profile{})
	addMapping("model_profile_docgen.go", nimona.ProfileRepository{})
	addMapping("peer_addr_docgen.go", nimona.PeerAddr{})
	addMapping("peer_docgen.go", nimona.PeerInfo{})
	addMapping("peer_docgen.go", nimona.PeerKey{})

	for filename, types := range mappings {
		err := nimona.GenerateDocumentMethods(
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
