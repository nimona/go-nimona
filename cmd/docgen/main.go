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

	addMapping("client_profile_docgen.go", nimona.Profile{})
	addMapping("client_profile_docgen.go", nimona.ProfileRepository{})
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
	addMapping("handler_network_docgen.go", nimona.NetworkAnnouncePeerRequest{})
	addMapping("handler_network_docgen.go", nimona.NetworkAnnouncePeerResponse{})
	addMapping("handler_network_docgen.go", nimona.NetworkInfoRequest{})
	addMapping("handler_network_docgen.go", nimona.NetworkJoinRequest{})
	addMapping("handler_network_docgen.go", nimona.NetworkJoinResponse{})
	addMapping("handler_network_docgen.go", nimona.NetworkLookupPeerRequest{})
	addMapping("handler_network_docgen.go", nimona.NetworkLookupPeerResponse{})
	addMapping("handler_network_docgen.go", nimona.NetworkResolveHandleRequest{})
	addMapping("handler_network_docgen.go", nimona.NetworkResolveHandleResponse{})
	addMapping("handler_peer_docgen.go", nimona.PeerCapabilitiesRequest{})
	addMapping("handler_peer_docgen.go", nimona.PeerCapabilitiesResponse{})
	addMapping("handler_ping_docgen.go", nimona.Ping{})
	addMapping("handler_ping_docgen.go", nimona.Pong{})
	addMapping("identity_docgen.go", nimona.Identity{})
	addMapping("identity_docgen.go", nimona.IdentityAlias{})
	addMapping("identity_docgen.go", nimona.KeyGraph{})
	addMapping("message_wrapper_docgen.go", nimona.MessageWrapper{})
	addMapping("network_docgen.go", nimona.NetworkAlias{})
	addMapping("network_docgen.go", nimona.NetworkIdentity{})
	addMapping("network_docgen.go", nimona.NetworkInfo{})
	addMapping("peer_addr_docgen.go", nimona.PeerAddr{})
	addMapping("peer_docgen.go", nimona.PeerInfo{})
	addMapping("peer_docgen.go", nimona.PeerKey{})

	for filename, types := range mappings {
		err := nimona.GenerateDocumentMapMethods(
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
