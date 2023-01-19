package main

import (
	"fmt"

	cbg "github.com/whyrusleeping/cbor-gen"

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

	addMapping("document_metadata_gen.go", nimona.Metadata{})
	addMapping("document_metadata_gen.go", nimona.Permissions{})
	addMapping("document_metadata_gen.go", nimona.PermissionsAllow{})
	addMapping("document_metadata_gen.go", nimona.PermissionsCondition{})
	addMapping("document_metadata_gen.go", nimona.Signature{})
	addMapping("fixtures_gen.go", nimona.CborFixture{})
	addMapping("handler_network_gen.go", nimona.NetworkInfo{})
	addMapping("handler_network_gen.go", nimona.NetworkInfoRequest{})
	addMapping("handler_peer_gen.go", nimona.PeerCapabilitiesRequest{})
	addMapping("handler_peer_gen.go", nimona.PeerCapabilitiesResponse{})
	addMapping("handler_stream_gen.go", nimona.StreamRequest{})
	addMapping("handler_stream_gen.go", nimona.StreamResponse{})
	addMapping("identifier_document_gen.go", nimona.DocumentID{})
	addMapping("identifier_network_gen.go", nimona.NetworkID{})
	addMapping("identifier_peer_gen.go", nimona.PeerID{})
	addMapping("message_ping_gen.go", nimona.Ping{})
	addMapping("message_ping_gen.go", nimona.Pong{})
	addMapping("message_wrapper_gen.go", nimona.MessageWrapper{})
	addMapping("peer_addr_gen.go", nimona.PeerAddr{})
	addMapping("peer_info_gen.go", nimona.PeerInfo{})
	addMapping("session_request_gen.go", nimona.Request{})
	addMapping("session_response_gen.go", nimona.Response{})
	addMapping("stream_gen.go", nimona.StreamOperation{})
	addMapping("stream_gen.go", nimona.StreamPatch{})

	for filename, types := range mappings {
		err := cbg.WriteMapEncodersToFile(
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
